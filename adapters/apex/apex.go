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
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var _ log.Handler = (*Handler)(nil)

const defaultBatchSize = 1024

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
func SetClientOptions(options ...axiom.Option) Option {
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
func SetIngestOptions(opts ...ingest.Option) Option {
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
	ingestOptions []ingest.Option

	eventCh   chan axiom.Event
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
// An API token with `ingest` permission is sufficient enough.
//
// Additional options can be supplied to configure the `Handler`.
//
// A handler needs to be closed properly to make sure all logs are sent by
// calling `Close()`.
func New(options ...Option) (*Handler, error) {
	handler := &Handler{
		eventCh: make(chan axiom.Event, defaultBatchSize),
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

	// Run background ingest.
	go func() {
		defer close(handler.closeCh)

		res, err := handler.client.Datasets.IngestChannel(context.Background(), handler.datasetName, handler.eventCh, handler.ingestOptions...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to ingest events: %s\n", err)
		} else if res.Failed > 0 {
			// Best effort on notifying the user about the ingest failure.
			fmt.Fprintf(os.Stderr, "event at %s failed to ingest: %s\n",
				res.Failures[0].Timestamp, res.Failures[0].Error)
		}
	}()

	return handler, nil
}

// Close the handler and make sure all events are flushed. Closing the handler
// renders it unusable for further use.
func (h *Handler) Close() {
	h.closeOnce.Do(func() {
		close(h.eventCh)
		<-h.closeCh
	})
}

// HandleLog implements `log.Handler`.
func (h *Handler) HandleLog(entry *log.Entry) error {
	event := axiom.Event{}

	// Set fields first.
	for k, v := range entry.Fields {
		event[k] = v
	}

	// Set timestamp, severity and actual message.
	event[ingest.TimestampField] = entry.Timestamp.Format(time.RFC3339Nano)
	event["severity"] = entry.Level.String()
	event["message"] = entry.Message

	h.eventCh <- event

	return nil
}
