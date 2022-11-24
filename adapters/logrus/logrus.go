package logrus

import (
	"context"
	"errors"
	stdlog "log"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var _ logrus.Hook = (*Hook)(nil)

const defaultBatchSize = 1024

// ErrMissingDatasetName is raised when a dataset name is not provided. Set it
// manually using the [SetDataset] option or export "AXIOM_DATASET".
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
// [axiom.NewClient] which is only called if no [axiom.Client] was specified by
// the [SetClient] option.
func SetClientOptions(options ...axiom.Option) Option {
	return func(h *Hook) error {
		h.clientOptions = options
		return nil
	}
}

// SetDataset specifies the dataset to ingest the logs into. Can also be
// specified using the "AXIOM_DATASET" environment variable.
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

// SetLevels sets the logrus levels that the Axiom [Hook] will create log
// entries for.
func SetLevels(levels ...logrus.Level) Option {
	return func(h *Hook) error {
		h.levels = levels
		return nil
	}
}

// Hook implements a [logrus.Hook] used for shipping logs to Axiom.
type Hook struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions []ingest.Option
	levels        []logrus.Level

	eventCh   chan axiom.Event
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new hook that ingests logs into Axiom. It automatically takes
// its configuration from the environment. To connect, export the following
// environment variables:
//
//   - AXIOM_TOKEN
//   - AXIOM_ORG_ID (only when using a personal token)
//   - AXIOM_DATASET
//
// The configuration can be set manually using options which are prefixed with
// "Set".
//
// An api token with "ingest" permission is sufficient enough.
//
// A hook needs to be closed properly to make sure all logs are sent by calling
// [Hook.Close].
func New(options ...Option) (*Hook, error) {
	hook := &Hook{
		levels: logrus.AllLevels,

		eventCh: make(chan axiom.Event, defaultBatchSize),
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

	// When the dataset name is not set, use "AXIOM_DATASET".
	if hook.datasetName == "" {
		hook.datasetName = os.Getenv("AXIOM_DATASET")
		if hook.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	// Run background ingest.
	go func() {
		defer close(hook.closeCh)

		logger := stdlog.New(os.Stderr, "[AXIOM|LOGRUS]", 0)

		res, err := hook.client.IngestChannel(context.Background(), hook.datasetName, hook.eventCh, hook.ingestOptions...)
		if err != nil {
			logger.Printf("failed to ingest events: %s\n", err)
		} else if res.Failed > 0 {
			// Best effort on notifying the user about the ingest failure.
			logger.Printf("event at %s failed to ingest: %s\n",
				res.Failures[0].Timestamp, res.Failures[0].Error)
		}
	}()

	return hook, nil
}

// Close the hook and make sure all events are flushed. This should be
// registered with [logrus.RegisterExitHandler]. Closing the hook renders it
// unusable for further use.
func (h *Hook) Close() {
	h.closeOnce.Do(func() {
		close(h.eventCh)
		<-h.closeCh
	})
}

// Levels implements [logrus.Hook].
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

// Fire implements [logrus.Hook].
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
