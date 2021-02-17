package apex

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/apex/log"

	"github.com/axiomhq/axiom-go/axiom"
)

var _ log.Handler = (*Handler)(nil)

const (
	batchSize    = 1024
	sendInterval = time.Second
)

// An Option modifies the behaviour of the Axiom handler.
type Option func(*Handler) error

// IngestOptions specifies the ingestion options to use for ingesting the logs.
func IngestOptions(opts axiom.IngestOptions) Option {
	return func(h *Handler) error {
		h.ingestOptions = opts
		return nil
	}
}

// Handler implements a log.Handler used for shipping logs to Axiom.
type Handler struct {
	client      *axiom.Client
	datasetName string

	ingestOptions axiom.IngestOptions

	eventCh   chan axiom.Event
	cancel    context.CancelFunc
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new Handler configured to talk to the specified Axiom
// deployment and authenticating with the given access token. An ingest token is
// sufficient enough. The logs will be ingested into the specified dataset.
// Additional options can be supplied to configure the Handler. A Handler needs
// to be closed properly to make sure all logs are sent by calling Close().
func New(baseURL, accessToken, datasetName string, options ...Option) (*Handler, error) {
	client, err := axiom.NewClient(baseURL, accessToken)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName, options...)
}

// NewCloud is like New() but configures the Handler to talk to Axiom Cloud.
func NewCloud(accessToken, orgID, datasetName string, options ...Option) (*Handler, error) {
	client, err := axiom.NewCloudClient(accessToken, orgID)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName, options...)
}

// NewWithClient behaves like New() but utilizes an already configured
// axiom.Client to talk to a deployment.
func NewWithClient(client *axiom.Client, datasetName string, options ...Option) (*Handler, error) {
	handler := &Handler{
		client:      client,
		datasetName: datasetName,

		eventCh: make(chan axiom.Event, 1),
		closeCh: make(chan struct{}),
	}

	// Apply supplied options.
	if err := handler.Options(options...); err != nil {
		return nil, err
	}

	// Run background scheduler.
	var ctx context.Context
	ctx, handler.cancel = context.WithCancel(context.Background())
	go handler.run(ctx, handler.closeCh)

	return handler, nil
}

// Options applies Options to the Handler.
func (h *Handler) Options(options ...Option) error {
	for _, option := range options {
		if err := option(h); err != nil {
			return err
		}
	}
	return nil
}

// Close the handler and make sure all events are flushed. Closing the handler
// renders it unusable for further use.
func (h *Handler) Close() {
	h.closeOnce.Do(func() {
		close(h.eventCh)
		h.cancel()
		<-h.closeCh
	})
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(entry *log.Entry) error {
	event := axiom.Event{}

	// Set fields first.
	for k, v := range entry.Fields {
		event[k] = v
	}

	// Set timestamp, severity and actual message.
	event[axiom.TimestampField] = entry.Timestamp.Format(time.RFC3339Nano)
	event["severity"] = entry.Level.String()
	event["message"] = entry.Message

	h.eventCh <- event

	return nil
}

func (h *Handler) run(ctx context.Context, closeCh chan struct{}) {
	defer close(closeCh)

	t := time.NewTicker(sendInterval)
	defer t.Stop()

	events := make([]axiom.Event, 0, batchSize)

	defer func() {
		flushCtx, flushCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer flushCancel()
		h.ingest(flushCtx, events)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if len(events) == 0 {
				continue
			}
		case event, ok := <-h.eventCh:
			if !ok {
				continue
			}

			events = append(events, event)

			if len(events) < batchSize {
				continue
			}
		}

		h.ingest(ctx, events)

		// Clear batch buffer.
		// TODO(lukasmalkmus): In the future we might want to implement some
		// kind of backoff and retry mechanism.
		events = make([]axiom.Event, 0, batchSize)
	}
}

func (h *Handler) ingest(ctx context.Context, events []axiom.Event) {
	if len(events) == 0 {
		return
	}

	res, err := h.client.Datasets.IngestEvents(ctx, h.datasetName, h.ingestOptions, events...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to ingest batch of %d events: %s\n", len(events), err)
	} else if res.Failed > 0 {
		// Best effort on notifying the user about the ingest failure.
		fmt.Fprintf(os.Stderr, "event at %s failed to ingest: %s\n",
			res.Failures[0].Timestamp, res.Failures[0].Error)
	}
}
