package logrus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var _ logrus.Hook = (*Hook)(nil)

const (
	batchSize    = 1024
	sendInterval = time.Second
)

// ErrMissingDatasetName is raised when a dataset name is not provided. Set it
// manually using the SetDataset option or export `AXIOM_DATASET`.
var ErrMissingDatasetName = errors.New("missing dataset name")

// An Option modifies the behaviour of the Axiom hook.
type Option func(*Hook) error

// SetClient specifies the Axiom client to use for ingesting the logs.
func SetClient(client *axiom.Client) Option {
	return func(h *Hook) error {
		h.client = client
		return nil
	}
}

// SetClientOptions specifies the Axiom client options to pass to
// `axiom.NewClient()`. `axiom.NewClient()` is only called if no client was
// specified by the `SetClient` option.
func SetClientOptions(options ...axiom.Option) Option {
	return func(h *Hook) error {
		h.clientOptions = options
		return nil
	}
}

// SetDataset specifies the dataset to ingest the logs into. Can also be
// specified using the `AXIOM_DATASET` environment variable.
func SetDataset(datasetName string) Option {
	return func(h *Hook) error {
		h.datasetName = datasetName
		return nil
	}
}

// SetIngestOptions specifies the ingestion options to use for ingesting the
// logs.
func SetIngestOptions(opts ...ingest.Option) Option {
	return func(h *Hook) error {
		h.ingestOptions = opts
		return nil
	}
}

// SetLevels sets the logrus levels that the Axiom hook will create log entries
// for.
func SetLevels(levels ...logrus.Level) Option {
	return func(h *Hook) error {
		h.levels = levels
		return nil
	}
}

// Hook implements a `logrus.Hook` used for shipping logs to Axiom.
type Hook struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions []ingest.Option
	levels        []logrus.Level

	eventCh   chan axiom.Event
	cancel    context.CancelFunc
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new `Hook` configured to ingest logs to the Axiom deployment
// and dataset as specified by the environment. Refer to `axiom.NewClient()` for
// more details on how configuring the Axiom deployment works or pass the
// `SetClient()` option to pass a custom client or `SetClientOptions()` to
// control the Axiom client creation. To specify the dataset set `AXIOM_DATASET`
// or use the `SetDataset()` option.
//
// An API token with `ingest` permission is sufficient enough.
//
// Additional options can be supplied to configure the `Hook`.
//
// A hook needs to be closed properly to make sure all logs are sent by calling
// `Close()`.
func New(options ...Option) (*Hook, error) {
	hook := &Hook{
		levels: logrus.AllLevels,

		eventCh: make(chan axiom.Event, 1),
		closeCh: make(chan struct{}),
	}

	// Apply supplied options.
	for _, option := range options {
		if err := option(hook); err != nil {
			return nil, err
		}
	}

	// Create client, if not set.
	if hook.client == nil {
		var err error
		if hook.client, err = axiom.NewClient(hook.clientOptions...); err != nil {
			return nil, err
		}
	}

	// When the dataset name is not set, use `AXIOM_DATASET`.
	if hook.datasetName == "" {
		hook.datasetName = os.Getenv("AXIOM_DATASET")
		if hook.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	// Run background scheduler.
	var ctx context.Context
	ctx, hook.cancel = context.WithCancel(context.Background())
	go hook.run(ctx, hook.closeCh)

	return hook, nil
}

// Close the hook and make sure all events are flushed. This should be
// registered with `logrus.RegisterExitHandler(h.Close)`. Closing the hook
// renders it unusable for further use.
func (h *Hook) Close() {
	h.closeOnce.Do(func() {
		close(h.eventCh)
		h.cancel()
		<-h.closeCh
	})
}

// Levels implements `logrus.Hook`.
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

// Fire implements `logrus.Hook`.
func (h *Hook) Fire(entry *logrus.Entry) error {
	event := axiom.Event{}

	// Set fields first.
	for k, v := range entry.Data {
		event[k] = v
	}

	// Set timestamp, severity and actual message.
	event[ingest.TimestampField] = entry.Time.Format(time.RFC3339Nano)
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
		events = make([]axiom.Event, 0, batchSize)
	}
}

func (h *Hook) ingest(ctx context.Context, events []axiom.Event) {
	if len(events) == 0 {
		return
	}

	res, err := h.client.Datasets.IngestEvents(ctx, h.datasetName, events, h.ingestOptions...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to ingest batch of %d events: %s\n", len(events), err)
	} else if res.Failed > 0 {
		// Best effort on notifying the user about the ingest failure.
		fmt.Fprintf(os.Stderr, "event at %s failed to ingest: %s\n",
			res.Failures[0].Timestamp, res.Failures[0].Error)
	}
}
