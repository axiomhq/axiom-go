package apex

import (
	"context"
	"errors"
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

// ErrMissingDatasetName is raised when a dataset name is not provided. Set it
// manually using the SetDataset option or export `AXIOM_DATASET`.
var ErrMissingDatasetName = errors.New("missing dataset name")

// An Option modifies the behaviour of the Axiom handler.
type Option func(*Handler) error

// SetClient specifies the Axiom client to use for ingesting the logs.
func SetClient(client *axiom.Client) Option {
	return func(h *Handler) error {
		h.client = client
		return nil
	}
}

// SetClientOptions specifies the Axiom client options to pass to
// `axiom.NewClient()`. `axiom.NewClient()` is only called if no client was
// specified by the `SetClient` option.
func SetClientOptions(options []axiom.Option) Option {
	return func(h *Handler) error {
		h.clientOptions = options
		return nil
	}
}

// SetDataset specifies the dataset to ingest the logs into. Can also be
// specified using the `AXIOM_DATASET` environment variable.
func SetDataset(datasetName string) Option {
	return func(h *Handler) error {
		h.datasetName = datasetName
		return nil
	}
}

// SetIngestOptions specifies the ingestion options to use for ingesting the
// logs.
func SetIngestOptions(opts axiom.IngestOptions) Option {
	return func(h *Handler) error {
		h.ingestOptions = opts
		return nil
	}
}

// Handler implements a `log.Handler` used for shipping logs to Axiom.
type Handler struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions axiom.IngestOptions

	eventCh   chan axiom.Event
	cancel    context.CancelFunc
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new `Handler` configured to ingest logs to the Axiom deployment
// and dataset as specified by the environment. Refer to `axiom.NewClient()` for
// more details on how configuring the Axiom deployment works or pass the
// `SetClient()` option to pass a custom client or `SetClientOptions()` to
// control the Axiom client creation. To specify the dataset set `AXIOM_DATASET`
// or use the `SetDataset()` option.
//
// An ingest token is sufficient enough. Additional options can be supplied to
// configure the `Handler`. A handler needs to be closed properly to make sure
// all logs are sent by calling `Close()`.
func New(options ...Option) (*Handler, error) {
	handler := &Handler{
		eventCh: make(chan axiom.Event, 1),
		closeCh: make(chan struct{}),
	}

	// Apply supplied options.
	for _, option := range options {
		if err := option(handler); err != nil {
			return nil, err
		}
	}

	// Create client, if not set.
	if handler.client == nil {
		var err error
		if handler.client, err = axiom.NewClient(handler.clientOptions...); err != nil {
			return nil, err
		}
	}

	// When the dataset name is not set, use `AXIOM_DATASET`.
	if handler.datasetName == "" {
		handler.datasetName = os.Getenv("AXIOM_DATASET")
		if handler.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	// Run background scheduler.
	var ctx context.Context
	ctx, handler.cancel = context.WithCancel(context.Background())
	go handler.run(ctx, handler.closeCh)

	return handler, nil
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
