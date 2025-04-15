package axiom

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/klauspost/compress/zstd"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=ContentType,ContentEncoding -linecomment -output=datasets_string.go

// ErrUnknownContentType is raised when the given [ContentType] is not valid.
var ErrUnknownContentType = errors.New("unknown content type")

// ErrUnknownContentEncoding is raised when the given [ContentEncoding] is not
// valid.
var ErrUnknownContentEncoding = errors.New("unknown content encoding")

// ContentType describes the content type of the data to ingest.
type ContentType uint8

const (
	// JSON treats the data as JSON array.
	JSON ContentType = iota + 1 // application/json
	// NDJSON treats the data as newline delimited JSON objects. Preferred
	// data format.
	NDJSON // application/x-ndjson
	// CSV treats the data as CSV content.
	CSV // text/csv
)

// ContentEncoding describes the content encoding of the data to ingest.
type ContentEncoding uint8

const (
	// Identity marks the data as not being encoded.
	Identity ContentEncoding = iota + 1 //
	// Gzip marks the data as being gzip encoded. A [GzipEncoder] can be used to
	// encode the data.
	Gzip // gzip
	// Zstd marks the data as being zstd encoded. A [ZstdEncoder] can be used to
	// encode the data.
	Zstd // zstd
)

// An Event is a map of key-value pairs.
//
// If you want to set a timestamp for the event you can either use
// [ingest.TimestampField] as map key for the timestamp or specify any other
// field that carries the timestamp by passing [ingest.SetTimestampField] to
// [DatasetsService.Ingest], [DatasetsService.IngestEvents] or
// [DatasetsService.IngestChannel] as an [Option].
type Event map[string]any

// Dataset represents an Axiom dataset.
type Dataset struct {
	// ID of the dataset.
	ID string `json:"id"`
	// Name is the unique name of the dataset.
	Name string `json:"name"`
	// Description of the dataset.
	Description string `json:"description"`
	// CreatedBy is the ID of the user who created the dataset.
	CreatedBy string `json:"who"`
	// CreatedAt is the time the dataset was created at.
	CreatedAt time.Time `json:"created"`
	// CanWrite sets whether writeable access is granted.
	CanWrite bool `json:"canWrite"`
	// UseRetentionPeriod sets whether the dataset uses a retention period.
	UseRetentionPeriod bool `json:"useRetentionPeriod"`
	// RetentionDays is the number of days events are kept in the dataset.
	RetentionDays int `json:"retentionDays"`
}

// DatasetCreateRequest is a request used to create a dataset.
type DatasetCreateRequest struct {
	// Name of the dataset to create. Restricted to 80 characters of [a-zA-Z0-9]
	// and special characters "-", "_" and ".". Special characters cannot be a
	// prefix or suffix. The prefix cannot be "axiom-".
	Name string `json:"name"`
	// Description of the dataset to create.
	Description string `json:"description"`
}

// DatasetUpdateRequest is a request used to update a dataset.
type DatasetUpdateRequest struct {
	// Description of the dataset to update.
	Description string `json:"description"`
}

type wrappedDataset struct {
	*Dataset

	// HINT(lukasmalkmus) This is some future stuff we don't yet support in this
	// package so we just ignore it for now.
	IntegrationConfigs any `json:"integrationConfigs,omitempty"`
	IntegrationFilters any `json:"integrationFilters,omitempty"`
	QuickQueries       any `json:"quickQueries,omitempty"`
}

type datasetTrimRequest struct {
	// MaxDuration marks the oldest timestamp an event can have before getting
	// deleted.
	MaxDuration string `json:"maxDuration"`
}

type aplQueryRequest struct {
	query.Options

	// APL is the APL query string.
	APL string `json:"apl"`
}

type aplQueryResponse struct {
	query.Result

	Format        any `json:"format"`
	Request       any `json:"request"`
	DatasetNames  any `json:"datasetNames"`
	FieldsMetaMap any `json:"fieldsMetaMap"`
}

// DatasetsService handles communication with the dataset related operations of
// the Axiom API.
//
// Axiom API Reference: /v2/datasets
type DatasetsService service

// List all available datasets.
func (s *DatasetsService) List(ctx context.Context) ([]*Dataset, error) {
	ctx, span := s.client.trace(ctx, "Datasets.List")
	defer span.End()

	var res []*wrappedDataset
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	datasets := make([]*Dataset, len(res))
	for i, r := range res {
		datasets[i] = r.Dataset
	}

	return datasets, nil
}

// Get a dataset by id.
func (s *DatasetsService) Get(ctx context.Context, id string) (*Dataset, error) {
	ctx, span := s.client.trace(ctx, "Datasets.Get", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res wrappedDataset
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res.Dataset, nil
}

// Create a dataset with the given properties.
func (s *DatasetsService) Create(ctx context.Context, req DatasetCreateRequest) (*Dataset, error) {
	ctx, span := s.client.trace(ctx, "Datasets.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
		attribute.String("axiom.param.description", req.Description),
	))
	defer span.End()

	var res wrappedDataset
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res.Dataset, nil
}

// Update the dataset identified by the given id with the given properties.
func (s *DatasetsService) Update(ctx context.Context, id string, req DatasetUpdateRequest) (*Dataset, error) {
	ctx, span := s.client.trace(ctx, "Datasets.Update", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
		attribute.String("axiom.param.description", req.Description),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res wrappedDataset
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res.Dataset, nil
}

// Delete the dataset identified by the given id.
func (s *DatasetsService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Datasets.Delete", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}

// Trim the dataset identified by its id to a given length. The max duration
// given will mark the oldest timestamp an event can have. Older ones will be
// deleted from the dataset.
func (s *DatasetsService) Trim(ctx context.Context, id string, maxDuration time.Duration) error {
	ctx, span := s.client.trace(ctx, "Datasets.Trim", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
		attribute.String("axiom.param.max_duration", maxDuration.String()),
	))
	defer span.End()

	req := datasetTrimRequest{
		MaxDuration: maxDuration.String(),
	}

	path, err := url.JoinPath(s.basePath, id, "trim")
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodPost, path, req, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}

// Ingest data into the dataset identified by its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// The reader is streamed to the server until EOF is reached on a single
// connection. Keep that in mind when dealing with slow readers.
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (s *DatasetsService) Ingest(ctx context.Context, id string, r io.Reader, typ ContentType, enc ContentEncoding, options ...ingest.Option) (*ingest.Status, error) {
	ctx, span := s.client.trace(ctx, "Datasets.Ingest", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
		attribute.String("axiom.param.content_type", typ.String()),
		attribute.String("axiom.param.content_encoding", enc.String()),
	))
	defer span.End()

	// Apply supplied options.
	var opts ingest.Options
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	// TODO(lukasmalkmus): Use 's.basePath' once ingest v2 is available.
	path, err := url.JoinPath("/v1/datasets", id, "ingest")
	if err != nil {
		return nil, spanError(span, err)
	} else if path, err = AddURLOptions(path, opts); err != nil {
		return nil, spanError(span, err)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, r)
	if err != nil {
		return nil, spanError(span, err)
	}

	if err = setOptionHeaders(req, opts, typ); err != nil {
		return nil, spanError(span, err)
	}

	switch typ {
	case JSON, NDJSON, CSV:
		req.Header.Set("Content-Type", typ.String())
	default:
		err = ErrUnknownContentType
		return nil, spanError(span, err)
	}

	switch enc {
	case Identity:
	case Gzip, Zstd:
		req.Header.Set("Content-Encoding", enc.String())
	default:
		err = ErrUnknownContentEncoding
		return nil, spanError(span, err)
	}

	var (
		res  ingest.Status
		resp *Response
	)
	if resp, err = s.client.Do(req, &res); err != nil {
		return nil, spanError(span, err)
	}
	res.TraceID = resp.TraceID()

	setIngestStatusOnSpan(span, res)

	return &res, nil
}

// IngestEvents ingests events into the dataset identified by its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// For ingesting large amounts of data, consider using the
// [DatasetsService.Ingest] or [DatasetsService.IngestChannel] method.
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (s *DatasetsService) IngestEvents(ctx context.Context, id string, events []Event, options ...ingest.Option) (*ingest.Status, error) {
	ctx, span := s.client.trace(ctx, "Datasets.IngestEvents", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
		attribute.Int("axiom.events_to_ingest", len(events)),
	))
	defer span.End()

	if len(events) == 0 {
		return &ingest.Status{}, nil
	}

	// Apply supplied options.
	var opts ingest.Options
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	// TODO(lukasmalkmus): Use 's.basePath' once ingest v2 is available.
	path, err := url.JoinPath("/v1/datasets", id, "ingest")
	if err != nil {
		return nil, spanError(span, err)
	} else if path, err = AddURLOptions(path, opts); err != nil {
		return nil, spanError(span, err)
	}

	getBody := func() (io.ReadCloser, error) {
		pr, pw := io.Pipe()

		zsw, wErr := zstd.NewWriter(pw)
		if wErr != nil {
			_ = pr.Close()
			_ = pw.Close()
			return nil, wErr
		}

		go func() {
			var (
				enc    = json.NewEncoder(zsw)
				encErr error
			)
			for _, event := range events {
				if encErr = enc.Encode(event); encErr != nil {
					break
				}
			}

			if closeErr := zsw.Close(); encErr == nil {
				// If we have no error from encoding but from closing, capture
				// that one.
				encErr = closeErr
			}
			_ = pw.CloseWithError(encErr)
		}()

		return pr, nil
	}

	r, err := getBody()
	if err != nil {
		return nil, spanError(span, err)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, r)
	if err != nil {
		return nil, spanError(span, err)
	}
	req.GetBody = getBody

	if err = setOptionHeaders(req, opts, NDJSON); err != nil {
		return nil, spanError(span, err)
	}

	req.Header.Set("Content-Type", NDJSON.String())
	req.Header.Set("Content-Encoding", Zstd.String())

	var (
		res  ingest.Status
		resp *Response
	)
	if resp, err = s.client.Do(req, &res); err != nil {
		return nil, spanError(span, err)
	}
	res.TraceID = resp.TraceID()

	setIngestStatusOnSpan(span, res)

	return &res, nil
}

// IngestChannel ingests events from a channel into the dataset identified by
// its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// Events are ingested in batches. A batch is either 10000 events for unbuffered
// channels or the capacity of the channel for buffered channels. The maximum
// batch size is 10000. A batch is sent to the server as soon as it is full,
// after one second or when the channel is closed.
//
// The method returns with an error when the context is marked as done or an
// error occurs when sending the events to the server. A partial ingestion is
// possible and the returned ingest status is valid to use. When the context is
// marked as done, no attempt is made to send the buffered events to the server.
//
// The method returns without an error if the channel is closed and the buffered
// events are successfully sent to the server.
//
// The returned ingest status does not contain a trace ID as the underlying
// implementation possibly sends multiple requests to the server thus generating
// multiple trace IDs.
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (s *DatasetsService) IngestChannel(ctx context.Context, id string, events <-chan Event, options ...ingest.Option) (*ingest.Status, error) {
	ctx, span := s.client.trace(ctx, "Datasets.IngestChannel", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
		attribute.Int("axiom.channel.capacity", cap(events)),
	))
	defer span.End()

	// Batch is either 10000 events for unbuffered channels or the capacity of
	// the channel for buffered channels. The maximum batch size is 10000.
	batchSize := 10_000
	if cap(events) > 0 && cap(events) <= batchSize {
		batchSize = cap(events)
	}
	batch := make([]Event, 0, batchSize)

	// Flush on a per second basis.
	const flushInterval = time.Second
	t := time.NewTicker(flushInterval)
	defer t.Stop()

	var ingestStatus ingest.Status
	defer func() {
		setIngestStatusOnSpan(span, ingestStatus)
	}()

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		res, err := s.IngestEvents(ctx, id, batch, options...)
		if err != nil {
			return fmt.Errorf("failed to ingest events: %w", err)
		}
		ingestStatus.Add(res)
		t.Reset(flushInterval) // Reset the ticker.
		batch = batch[:0]      // Clear the batch.

		return nil
	}

	for {
		select {
		case <-ctx.Done():

			return &ingestStatus, spanError(span, context.Cause(ctx))
		case event, ok := <-events:
			if !ok {
				// Channel is closed.
				err := flush()
				return &ingestStatus, spanError(span, err)
			}
			batch = append(batch, event)

			if len(batch) >= batchSize {
				if err := flush(); err != nil {
					return &ingestStatus, spanError(span, err)
				}
			}
		case <-t.C:
			if err := flush(); err != nil {
				return &ingestStatus, spanError(span, err)
			}
		}
	}
}

// Query executes the given query specified using the Axiom Processing
// Language (APL).
//
// To learn more about APL, please refer to [our documentation].
//
// [our documentation]: https://www.axiom.co/docs/apl/introduction
func (s *DatasetsService) Query(ctx context.Context, apl string, options ...query.Option) (*query.Result, error) {
	// Apply supplied options.
	var opts query.Options
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	ctx, span := s.client.trace(ctx, "Datasets.Query", trace.WithAttributes(
		attribute.String("axiom.param.apl", apl),
		attribute.String("axiom.param.start_time", opts.StartTime.String()),
		attribute.String("axiom.param.end_time", opts.EndTime.String()),
		attribute.String("axiom.param.cursor", opts.Cursor),
	))
	defer span.End()

	// The only query parameters supported can be hardcoded as they are not
	// configurable as of now.
	queryParams := struct {
		Format string `url:"format"`
	}{
		Format: "tabular", // Hardcode tabular result format for now.
	}

	path, err := url.JoinPath(s.basePath, "query")
	if err != nil {
		return nil, spanError(span, err)
	} else if path, err = AddURLOptions(path, queryParams); err != nil {
		return nil, spanError(span, err)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, aplQueryRequest{
		Options: opts,

		APL: apl,
	})
	if err != nil {
		return nil, spanError(span, err)
	}

	var (
		res  aplQueryResponse
		resp *Response
	)
	if resp, err = s.client.Do(req, &res); err != nil {
		return nil, spanError(span, err)
	}
	res.TraceID = resp.TraceID()

	setQueryStatusOnSpan(span, res.Status)
	span.SetAttributes(attribute.String("axiom.trace_id", res.TraceID))

	return &res.Result, nil
}

// QueryLegacy executes the given legacy query on the dataset identified by its
// id.
//
// Deprecated: Legacy queries will be replaced by queries specified using the
// Axiom Processing Language (APL) and the legacy query API will be removed in
// the future. Use [DatasetsService.Query] instead.
func (s *DatasetsService) QueryLegacy(ctx context.Context, id string, q querylegacy.Query, opts querylegacy.Options) (*querylegacy.Result, error) {
	ctx, span := s.client.trace(ctx, "Datasets.QueryLegacy", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
	))
	defer span.End()

	if opts.SaveKind == querylegacy.APL {
		err := fmt.Errorf("invalid query kind %q: must be %q or %q",
			opts.SaveKind, querylegacy.Analytics, querylegacy.Stream)
		return nil, spanError(span, err)
	}

	path, err := url.JoinPath("/v1/datasets", id, "query")
	if err != nil {
		return nil, spanError(span, err)
	} else if path, err = AddURLOptions(path, opts); err != nil {
		return nil, spanError(span, err)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, q)
	if err != nil {
		return nil, spanError(span, err)
	}

	var (
		res struct {
			querylegacy.Result

			// HINT(lukasmalkmus): Ignore these fields as they are not relevant
			// for the user.
			FieldsMeta any `json:"fieldsMeta"`
			Format     any `json:"format"`
		}
		resp *Response
	)
	if resp, err = s.client.Do(req, &res); err != nil {
		return nil, spanError(span, err)
	}
	res.SavedQueryID = resp.Header.Get("X-Axiom-History-Query-Id")
	res.TraceID = resp.TraceID()

	setLegacyQueryStatusOnSpan(span, res.Status)
	span.SetAttributes(attribute.String("axiom.trace_id", res.TraceID))

	return &res.Result, nil
}

// DetectContentType detects the content type of a readers data. The returned
// reader must be used instead of the passed one. Compressed content is not
// detected.
func DetectContentType(r io.Reader) (io.Reader, ContentType, error) {
	var (
		br  = bufio.NewReader(r)
		typ ContentType
	)
	for {
		var (
			c   rune
			err error
		)
		if c, _, err = br.ReadRune(); err == io.EOF {
			return nil, 0, errors.New("couldn't find beginning of supported ingestion format")
		} else if err != nil {
			return nil, 0, err
		} else if c == '[' {
			typ = JSON
		} else if c == '{' {
			typ = NDJSON
		} else if unicode.IsLetter(c) || c == '"' { // We assume a CSV table starts with a letter or a quote.
			typ = CSV
		} else if unicode.IsSpace(c) {
			continue
		} else {
			return nil, 0, errors.New("cannot determine content type")
		}

		if err = br.UnreadRune(); err != nil {
			return nil, 0, err
		}
		break
	}

	// Create a new reader and prepend what we have already consumed in order to
	// figure out the content type.
	buf, err := br.Peek(br.Buffered())
	if err != nil {
		return nil, 0, err
	}
	alreadyRead := bytes.NewReader(buf)
	r = io.MultiReader(alreadyRead, r)

	return r, typ, nil
}

func setIngestStatusOnSpan(span trace.Span, status ingest.Status) {
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("axiom.trace_id", status.TraceID),
		attribute.Int64("axiom.events.ingested", int64(status.Ingested)),              //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.events.failed", int64(status.Failed)),                  //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.events.processed_bytes", int64(status.ProcessedBytes)), //nolint:gosec // Fine for this use case.
	)
}

func setQueryStatusOnSpan(span trace.Span, status query.Status) {
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("axiom.query.min_cursor", status.MinCursor),
		attribute.String("axiom.query.max_cursor", status.MaxCursor),
		attribute.String("axiom.query.elapsed_time", status.ElapsedTime.String()),
		attribute.Int64("axiom.query.rows_examined", int64(status.RowsExamined)), //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.query.rows_matched", int64(status.RowsMatched)),   //nolint:gosec // Fine for this use case.
	)
}

func setLegacyQueryStatusOnSpan(span trace.Span, status querylegacy.Status) {
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("axiom.querylegacy.elapsed_time", status.ElapsedTime.String()),
		attribute.Int64("axiom.querylegacy.blocks_examined", int64(status.BlocksExamined)), //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.querylegacy.rows_examined", int64(status.RowsExamined)),     //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.querylegacy.rows_matched", int64(status.RowsMatched)),       //nolint:gosec // Fine for this use case.
		attribute.Int64("axiom.querylegacy.num_groups", int64(status.NumGroups)),
		attribute.Bool("axiom.querylegacy.is_partial", status.IsPartial),
		attribute.Bool("axiom.querylegacy.is_estimate", status.IsEstimate),
		attribute.String("axiom.querylegacy.min_block_time", status.MinBlockTime.String()),
		attribute.String("axiom.querylegacy.max_block_time", status.MaxBlockTime.String()),
		attribute.String("axiom.querylegacy.min_cursor", status.MinCursor),
		attribute.String("axiom.querylegacy.max_cursor", status.MaxCursor),
	)
}

func setOptionHeaders(req *http.Request, opts ingest.Options, typ ContentType) error {
	// Set event labels.
	if err := setEventLabels(req, opts.EventLabels); err != nil {
		return err
	}

	// Set object/csv fields. The former is only valid for JSON and NDJSON as
	// the latter is obviously only valid for CSV. Both are equally optional.
	if typ == CSV {
		req.Header.Set("X-Axiom-CSV-Fields", strings.Join(opts.CSVFields, ","))
	}

	return nil
}

func setEventLabels(req *http.Request, labels map[string]any) error {
	if len(labels) == 0 {
		return nil
	}

	b, err := json.Marshal(labels)
	if err != nil {
		return err
	}

	req.Header.Set(headerEventLabels, string(b))

	return nil
}
