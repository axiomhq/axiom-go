package apex

import (
	"context"
	"errors"
	stdlog "log"
	"maps"
	"os"
	"sync"
	"time"

	"github.com/apex/log"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var _ log.Handler = (*Handler)(nil)

const defaultBatchSize = 10_000

// ErrMissingDatasetName is raised when a dataset name is not provided. Set it
// manually using the [SetDataset] option or export "AXIOM_DATASET".
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
// [axiom.NewClient] which is only called if no [axiom.Client] was specified by
// the [SetClient] option.
func SetClientOptions(options ...axiom.Option) Option {
	return func(h *Handler) error {
		h.clientOptions = options
		return nil
	}
}

// SetDataset specifies the dataset to ingest the logs into. Can also be
// specified using the "AXIOM_DATASET" environment variable.
func SetDataset(datasetName string) Option {
	return func(h *Handler) error {
		h.datasetName = datasetName
		return nil
	}
}

// SetIngestOptions specifies the ingestion options to use for ingesting the
// logs.
func SetIngestOptions(opts ...ingest.Option) Option {
	return func(h *Handler) error {
		h.ingestOptions = opts
		return nil
	}
}

// Handler implements a [log.Handler] used for shipping logs to Axiom.
type Handler struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions []ingest.Option

	eventCh   chan axiom.Event
	stopCh    chan struct{}
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New creates a new handler that ingests logs into Axiom. It automatically
// takes its configuration from the environment. To connect, export the
// following environment variables:
//
//   - AXIOM_TOKEN
//   - AXIOM_ORG_ID (only when using a personal token)
//   - AXIOM_DATASET
//
// The configuration can be set manually using options which are prefixed with
// "Set".
//
// An API token with "ingest" permission is sufficient enough.
//
// A handler needs to be closed properly to make sure all logs are sent by
// calling [Handler.Close].
func New(options ...Option) (*Handler, error) {
	handler := &Handler{
		eventCh: make(chan axiom.Event, defaultBatchSize),
		stopCh:  make(chan struct{}),
		closeCh: make(chan struct{}),
	}

	// Apply supplied options.
	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(handler); err != nil {
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

	// When the dataset name is not set, use "AXIOM_DATASET".
	if handler.datasetName == "" {
		handler.datasetName = os.Getenv("AXIOM_DATASET")
		if handler.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	// Run background ingest.
	go func() {
		defer close(handler.closeCh)

		logger := stdlog.New(os.Stderr, "[AXIOM|APEX]", 0)

		for {
			if res, err := handler.client.IngestChannel(context.Background(), handler.datasetName, handler.eventCh, handler.ingestOptions...); err != nil {
				logger.Printf("failed to ingest events: %s\n", err)
			} else if res.Failed > 0 {
				// Best effort on notifying the user about the ingest failure.
				logger.Printf("event at %s failed to ingest: %s\n",
					res.Failures[0].Timestamp, res.Failures[0].Error)
			}

			select {
			case <-handler.stopCh:
				return
			case <-time.After(time.Second):
			}
		}
	}()

	return handler, nil
}

// Close the handler and make sure all events are flushed. Closing the handler
// renders it unusable for further use.
func (h *Handler) Close() {
	h.closeOnce.Do(func() {
		close(h.stopCh)
		close(h.eventCh)
		<-h.closeCh
	})
}

// HandleLog implements [log.Handler].
func (h *Handler) HandleLog(entry *log.Entry) error {
	event := axiom.Event{}

	// Set fields first.
	maps.Copy(event, entry.Fields)

	// Set timestamp, severity and actual message.
	event[ingest.TimestampField] = entry.Timestamp.Format(time.RFC3339Nano)
	event["severity"] = entry.Level.String()
	event["message"] = entry.Message

	select {
	case <-h.closeCh:
		return errors.New("handler closed")
	default:
		h.eventCh <- event
		return nil
	}
}
