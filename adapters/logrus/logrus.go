package logrus

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomhq/axiom-go/axiom"
)

var _ logrus.Hook = (*Hook)(nil)

const (
	batchSize    = 1024
	sendInterval = time.Second
)

// An Option modifies the behaviour of the Axiom hook.
type Option func(*Hook) error

// Levels sets the logrus levels that the Axiom hook will create log entries
// for.
func Levels(levels ...logrus.Level) Option {
	return func(h *Hook) error {
		h.levels = levels
		return nil
	}
}

// IngestOptions specifies the ingestion options to use for ingesting the logs.
func IngestOptions(opts axiom.IngestOptions) Option {
	return func(h *Hook) error {
		h.ingestOptions = opts
		return nil
	}
}

// Hook implements an Axiom logging hook for use with logrus.
type Hook struct {
	client      *axiom.Client
	datasetName string

	levels        []logrus.Level
	ingestOptions axiom.IngestOptions

	eventCh   chan axiom.Event
	cancel    context.CancelFunc
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new Hook configured to talk to the specified Axiom deployment
// and authenticating with the given access token. An ingest token is
// sufficient enough. The logs will be ingested into the specified dataset.
// Additional options can be supplied to configure the Hook. A Hook needs to be
// closed properly to make sure all logs are sent by calling Close()
func New(baseURL, accessToken, datasetName string, options ...Option) (*Hook, error) {
	client, err := axiom.NewClient(baseURL, accessToken)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName)
}

// NewCloud is like New() but configures the Hook to talk to Axiom Cloud.
func NewCloud(accessToken, datasetName string, options ...Option) (*Hook, error) {
	client, err := axiom.NewCloudClient(accessToken)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName)
}

// NewWithClient behaves like New() but utilizes an already configured
// axiom.Client to talk to a deployment.
func NewWithClient(client *axiom.Client, datasetName string, options ...Option) (*Hook, error) {
	hook := &Hook{
		client:      client,
		datasetName: datasetName,

		levels: logrus.AllLevels,

		eventCh: make(chan axiom.Event, 1),
		closeCh: make(chan struct{}),
	}

	// Apply supplied options.
	if err := hook.Options(options...); err != nil {
		return nil, err
	}

	// Run background scheduler.
	var ctx context.Context
	ctx, hook.cancel = context.WithCancel(context.Background())
	go hook.run(ctx, hook.closeCh)

	return hook, nil
}

// Options applies Options to the Hook.
func (h *Hook) Options(options ...Option) error {
	for _, option := range options {
		if err := option(h); err != nil {
			return err
		}
	}
	return nil
}

// Close the hook and make sure all events are flushed. This should be
// registered with `logrus.RegisterExitHandler(h.Close)`. Closing the hooks
// renders it unusable for further use.
func (h *Hook) Close() {
	h.closeOnce.Do(func() {
		close(h.eventCh)
		h.cancel()
		<-h.closeCh
	})
}

// Levels implements logrus.Hook.
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

// Fire implements logrus.Hook.
func (h *Hook) Fire(entry *logrus.Entry) error {
	event := axiom.Event{}

	// Set fields first.
	for k, v := range entry.Data {
		event[k] = v
	}

	// Set timestamp, severity and actual message.
	event[axiom.TimestampField] = entry.Time.Format(time.RFC3339Nano)
	event["severity"] = entry.Level.String()
	event["message"] = entry.Message

	h.eventCh <- event

	return nil
}

func (h *Hook) run(ctx context.Context, closeCh chan struct{}) {
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

func (h *Hook) ingest(ctx context.Context, events []axiom.Event) {
	res, err := h.client.Datasets.IngestEvents(ctx, h.datasetName, h.ingestOptions, events...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to ingest batch of %d events: %s\n", len(events), err)
	} else if res.Failed > 0 {
		// Best effort on notifying the user about the ingest failure.
		fmt.Fprintf(os.Stderr, "event at %s failed to ingest: %s\n",
			res.Failures[0].Timestamp, res.Failures[0].Error)
	}
}
