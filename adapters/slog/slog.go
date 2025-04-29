package slog

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var _ slog.Handler = (*Handler)(nil)

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

// SetLevel specifies the log level the handler is enabled for.
func SetLevel(level slog.Leveler) Option {
	return func(h *Handler) error {
		h.level = level
		return nil
	}
}

// SetAddSource specifies whether to add the source of the log message to the event.
func SetAddSource(addSource bool) Option {
	return func(h *Handler) error {
		h.addSource = addSource
		return nil
	}
}

type rootHandler struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions []ingest.Option

	eventCh   chan axiom.Event
	closeCh   chan struct{}
	closeOnce sync.Once
}

// Handler implements a [slog.Handler] used for shipping logs to Axiom.
type Handler struct {
	*rootHandler

	level     slog.Leveler
	attrs     []slog.Attr
	groups    []string
	addSource bool
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
	root := &rootHandler{
		eventCh: make(chan axiom.Event, defaultBatchSize),
		closeCh: make(chan struct{}),
	}

	handler := &Handler{
		rootHandler: root,
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
	if root.client == nil {
		var err error
		if root.client, err = axiom.NewClient(root.clientOptions...); err != nil {
			return nil, err
		}
	}

	// When the dataset name is not set, use "AXIOM_DATASET".
	if root.datasetName == "" {
		root.datasetName = os.Getenv("AXIOM_DATASET")
		if root.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	// Run background ingest.
	go func() {
		defer close(root.closeCh)

		logger := log.New(os.Stderr, "[AXIOM|SLOG]", 0)

		res, err := root.client.IngestChannel(context.Background(), root.datasetName, root.eventCh, root.ingestOptions...)
		if err != nil {
			logger.Printf("failed to ingest events: %s\n", err)
		} else if res.Failed > 0 {
			// Best effort on notifying the user about the ingest failure.
			logger.Printf("event at %s failed to ingest: %s\n",
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

// Enabled implements [slog.Handler].
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.level != nil {
		minLevel = h.level.Level()
	}
	return level >= minLevel
}

// Handle implements [slog.Handler].
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	event := axiom.Event{}

	// Set handler attributes first, record attributes second.
	for _, attr := range h.attrs {
		addAttrToEvent(event, attr)
	}
	r.Attrs(func(attr slog.Attr) bool {
		addAttrToEvent(event, attr)
		return true
	})

	// Nest attributes in handler groups as objects, if any.
	for i := len(h.groups) - 1; i >= 0; i-- {
		event = axiom.Event{h.groups[i]: event}
	}

	// Set timestamp, level and actual message. The zero time is ignored.
	if !r.Time.IsZero() {
		event[ingest.TimestampField] = r.Time.Format(time.RFC3339Nano)
	}
	event[slog.LevelKey] = r.Level.String()
	event[slog.MessageKey] = r.Message

	if h.addSource {
		event[slog.SourceKey] = source(r)
	}

	select {
	case <-h.closeCh:
		return errors.New("handler closed")
	default:
		h.eventCh <- event
		return nil
	}
}

// WithAttrs implements [slog.Handler].
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)
	return h2
}

// WithGroup implements [slog.Handler].
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

func (h *Handler) clone() *Handler {
	return &Handler{
		rootHandler: h.rootHandler,

		level:     h.level,
		attrs:     slices.Clip(h.attrs),
		groups:    slices.Clip(h.groups),
		addSource: h.addSource,
	}
}

func addAttrToEvent(event axiom.Event, attr slog.Attr) {
	if attr.Equal(slog.Attr{}) {
		return
	}

	// If we have a group, nest it as an object.
	v := attr.Value.Resolve()
	if v.Kind() == slog.KindGroup {
		group := axiom.Event{}
		for _, attr := range v.Group() {
			addAttrToEvent(group, attr)
		}
		if len(group) > 0 && attr.Key != "" {
			event[attr.Key] = group
		}
	} else {
		event[attr.Key] = v.Any()
	}
}

// source returns a Source for the log event.
// If the Record was created without the necessary information,
// or if the location is unavailable, it returns a non-nil *Source
// with zero fields.
func source(r slog.Record) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}
