package axiom

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

//go:generate ../bin/stringer -type=ContentType,ContentEncoding -linecomment -output=datasets_string.go

var (
	// ErrUnknownContentType is raised when the given content type is not valid.
	ErrUnknownContentType = errors.New("unknown content type")
	// ErrUnknownContentEncoding is raised when the given content encoding is
	// not valid.
	ErrUnknownContentEncoding = errors.New("unknown content encoding")
)

// ContentType describes the content type of the data to ingest.
type ContentType uint8

const (
	// JSON treats the data as JSON array.
	JSON ContentType = iota + 1 // application/json
	// NDJSON treats the data as newline delimited JSON objects. Preferred as it
	// is faster than JSON array based ingestion.
	// TODO(lukasmalkmus): Is this still true?
	NDJSON // application/x-ndjson
	// CSV treats the data as CSV content.
	CSV // text/csv
)

// ContentEncoding describes the content encoding of the data to ingest.
type ContentEncoding uint8

const (
	// Identity marks the data as not being encoded.
	Identity ContentEncoding = iota + 1 //
	// GZIP marks the data as being gzip encoded.
	GZIP // gzip
)

// A FilterOp can be applied on queries to filter based on different conditions.
type FilterOp string

// All available query filter operations.
const (
	OpAnd FilterOp = "and"
	OpOr  FilterOp = "or"
	OpNot FilterOp = "not"

	// Works for strings and numbers.
	OpEqual     FilterOp = "=="
	OpNotEqual  FilterOp = "!="
	OpExists    FilterOp = "exists"
	OpNotExists FilterOp = "not-exists"

	// Only works for numbers.
	OpGreaterThan      FilterOp = ">"
	OpGreaterThanEqual FilterOp = ">="
	OpLessThan         FilterOp = "<"
	OpLessThanEqual    FilterOp = "<="

	// Only works for strings.
	OpStartsWith    FilterOp = "starts-with"
	OpNotStartsWith FilterOp = "not-starts-with"
	OpEndsWith      FilterOp = "ends-with"
	OpNotEndsWith   FilterOp = "not-ends-with"
	OpRegexp        FilterOp = "regexp"     // Uses `regexp.Compile` internally.
	OpNotRegexp     FilterOp = "not-regexp" // Uses `regexp.Compile` internally.

	// Works for strings and arrays.
	OpContains    FilterOp = "contains"
	OpNotContains FilterOp = "not-contains"
)

// An AggregationOp can be applied on queries to aggrgate based on different
// conditions.
type AggregationOp string

// All available query aggregation operations.
const (
	// Works with all types, field should be `*`.
	OpCount         AggregationOp = "count"
	OpCountDistinct AggregationOp = "distinct"

	// Only works for numbers.
	OpSum         AggregationOp = "sum"
	OpAvg         AggregationOp = "avg"
	OpMin         AggregationOp = "min"
	OpMax         AggregationOp = "max"
	OpTopk        AggregationOp = "topk"
	OpPercentiles AggregationOp = "percentiles"
	OpHistogram   AggregationOp = "histogram"
)

// An Event is a map of key-value pairs.
type Event map[string]interface{}

// Dataset represents an Axiom dataset.
type Dataset struct {
	// ID of the dataset.
	ID string `json:"id"`
	// Name is the unique name of the dataset.
	Name string `json:"name"`
	// Description of the dataset.
	Description string `json:"description"`
	// Created is the time the dataset was created at.
	Created time.Time `json:"created"`
}

// Field represents a field of an Axiom dataset.
type Field struct {
	// Name is the unique name of the field.
	Name string `json:"name"`
	// Type is the datatype of the field.
	Type string `json:"type"`
}

// DatasetInfo represents the details of the information stored inside an Axiom
// dataset.
type DatasetInfo struct {
	// Name is the unique name of the dataset.
	Name string `json:"name"`
	// NumBlocks is the number of blocks of the dataset.
	NumBlocks uint64 `json:"numBlocks"`
	// NumEvents is the number of events of the dataset.
	NumEvents uint64 `json:"numEvents"`
	// NumFields is the number of fields of the dataset.
	NumFields uint32 `json:"numFields"`
	// InputBytes is the amount of data stored in the dataset.
	InputBytes uint64 `json:"inputBytes"`
	// InputBytesHuman is the amount of data stored in the dataset formatted in
	// a human readable format.
	InputBytesHuman string `json:"inputBytesHuman"`
	// CompressedBytes is the amount of compressed data stored in the dataset.
	CompressedBytes uint64 `json:"compressedBytes"`
	// CompressedBytesHuman is the amount of compressed data stored in the
	// dataset formatted in a human readable format.
	CompressedBytesHuman string `json:"compressedBytesHuman"`
	// MinTime is the time of the oldest event stored in the dataset.
	MinTime time.Time `json:"minTime"`
	// MaxTime is the time of the newest event stored in the dataset.
	MaxTime time.Time `json:"maxTime"`
	// Fields are the fields of the dataset.
	Fields []Field `json:"fields"`
	// Created is the time the dataset was created at.
	Created time.Time `json:"created"`
}

// DatasetStats are the stats of
type DatasetStats struct {
	Datasets []*DatasetInfo `json:"datasets"`
	// NumBlocks is the number of blocks of the dataset.
	NumBlocks uint64 `json:"numBlocks"`
	// NumEvents is the number of events of the dataset.
	NumEvents uint64 `json:"numEvents"`
	// InputBytes is the amount of data stored in the dataset.
	InputBytes uint64 `json:"inputBytes"`
	// InputBytesHuman is the amount of data stored in the dataset formatted in
	// a human readable format.
	InputBytesHuman string `json:"inputBytesHuman"`
	// CompressedBytes is the amount of compressed data stored in the dataset.
	CompressedBytes uint64 `json:"compressedBytes"`
	// CompressedBytesHuman is the amount of compressed data stored in the
	// dataset formatted in a human readable format.
	CompressedBytesHuman string `json:"compressedBytesHuman"`
}

// IngestStatus is the status after an event ingestion operation.
type IngestStatus struct {
	// Ingested is the amount of events that have been ingested.
	Ingested uint64 `json:"ingested"`
	// Failed is the amount of events that failed to ingest.
	Failed uint64 `json:"failed"`
	// Failures are the ingestion failures, if any.
	Failures []*IngestFailure `json:"failures"`
	// ProcessedBytes is the number of bytes processed.
	ProcessedBytes uint64 `json:"processedBytes"`
	// BlocksCreated is the amount of blocks created.
	BlocksCreated uint32 `json:"blocksCreated"`
	// WALLength is the length of the Write-Ahead Log.
	WALLength uint32 `json:"walLength"`
}

// IngestFailure describes the ingestion failure of a single event.
type IngestFailure struct {
	// Timestamp of the event that failed to ingest.
	Timestamp time.Time `json:"timestamp"`
	// Error that made the event fail to ingest.
	Error string `json:"error"`
}

// Query represents a query that gets executed on a dataset.
type Query struct {
	// StartTime of the query. Required.
	StartTime time.Time `json:"startTime"`
	// EndTime of the query. Required.
	EndTime time.Time `json:"endTime"`
	// Resolution of the queries graph. Valid values are the queries time
	// range / 100 at maximum and / 1000 at minimum. Use zero value for
	// serve-side auto-detection.
	Resolution time.Duration `json:"resolution"`
	// Aggregations performed as part of the query.
	Aggregations []Aggregation `json:"aggregations"`
	// Filter applied on the queried results.
	// TODO(lukasmalkmus): Be more accurate! When is a filter applied in the
	// query lifecycle?
	Filter Filter `json:"filter"`
	// GroupBy specifies a list of field names to group the query result by.
	GroupBy []string `json:"groupBy"`
	// Order specifies a list of order rules that specify the order of the query
	// result.
	Order []Order `json:"order"`
	// Limit the amount of results returned from the query.
	Limit uint32 `json:"limit"`
	// VirtualColumns specifies a list of virtual fields that can be referenced
	// by aggregations, filters and orders.
	VirtualColumns []VirtualColumn `json:"virtualFields"`
	// Cursor is the query cursor.
	// TODO(lukasmalkmus): What is a query cursor?
	Cursor string `json:"cursor"`
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// Resolution to its string representation because that's what the server
// expects.
func (q Query) MarshalJSON() ([]byte, error) {
	type LocalQuery Query
	localQuery := struct {
		LocalQuery

		Resolution string `json:"resolution"`
	}{
		LocalQuery: LocalQuery(q),

		Resolution: q.Resolution.String(),
	}

	// If the resolution is not specified, it is set to auto for resolution
	// auto-detection on the server side.
	if q.Resolution == 0 {
		localQuery.Resolution = "auto"
	}

	return json.Marshal(localQuery)
}

// Aggregation performed as part of a query.
type Aggregation struct {
	// Op is the operation of the aggregation.
	Op AggregationOp `json:"op"`
	// Field the aggregation operation is performed on.
	Field string `json:"field"`
	// Argument to the aggregation.
	// TODO(lukasmalkmus): What exactly is an argument to an aggregation?
	Argument interface{} `json:"argument"`
}

// Filter applied as part of a query.
type Filter struct {
	// Op is the operation of the filter.
	Op FilterOp `json:"op"`
	// Field the filter operation is performed on.
	Field string `json:"field"`
	// Value to perform the filter operation against.
	Value interface{} `json:"value"`
	// CaseInsensitive specifies if the filter is case insensitive or not. Only
	// valid for OpStartsWith, OpNotStartsWith, OpEndsWith, OpNotEndsWith,
	// OpContains and OpNotContains.
	// TODO(lukasmalkmus): Why not "==" and "!="?
	CaseInsensitive bool `json:"caseInsensitive"`
	// Children specifies child filters for the filter. Only valid for OpAnd,
	// OpOr and OpNot.
	Children []Filter `json:"children"`
}

// Order specifies the order a queries result will be in.
type Order struct {
	// Field to order on.
	Field string `json:"field"`
	// Desc specifies if the field is ordered ascending or descending.
	Desc bool `json:"desc"`
}

// A VirtualColumn is not part of a dataset and its value is derived from an
// expression. Aggregations, filters and orders can reference this field like
// any other field.
// TODO(lukasmalkmus): Why is this not called VirtualField (apart from the name
// clashing with the type in vfields.go).
type VirtualColumn struct {
	// Alias the virtual field is referenced by.
	Alias string `json:"alias"`
	// Expression which specifies the virtual fields value.
	Expression string `json:"expr"`
}

// QueryResult is the result of a query.
type QueryResult struct {
	// Status of the query result.
	Status QueryStatus `json:"status"`
	// Matches are the events that matched the query.
	Matches []Entry `json:"matches"`
	// Buckets are the time series buckets.
	Buckets Timeseries `json:"buckets"`
}

// QueryStatus is the status of a query result.
type QueryStatus struct {
	// ElapsedTime is the duration it took the query to execute.
	ElapsedTime time.Duration `json:"elapsedTime"`
	// BlocksExamined is the amount of blocks that have been examined by the
	// query.
	BlocksExamined uint64 `json:"blocksExamined"`
	// RowsExamined is the amount of rows that have been examined by the query.
	RowsExamined uint64 `json:"rowsExamined"`
	// RowsMatched is the amount of rows that matched the query.
	RowsMatched uint64 `json:"rowsMatched"`
	// NumGroups is the amount of groups returned by the query.
	NumGroups uint32 `json:"numGroups"`
	// IsPartial describes if the query result is a partial result.
	IsPartial bool `json:"isPartial"`
	// IsEstimate describes if the query result is estimated.
	IsEstimate bool `json:"isEstimate"`
	// CacheStatus describes the effects the query had on cache.
	// TODO(lukasmalkmus): Better docs: What do the returned values mean?
	CacheStatus uint8 `json:"cacheStatus"`
	// MinBlockTime is the timestamp of the oldest block examined.
	MinBlockTime time.Time `json:"minBlockTime"`
	// MaxBlockTime is the timestamp of the newest block examined.
	MaxBlockTime time.Time `json:"maxBlockTime"`
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// ElapsedTime into a proper time.Duration value because the server returns it
// in microseconds.
func (qs *QueryStatus) UnmarshalJSON(b []byte) error {
	type localQueryStatus *QueryStatus

	if err := json.Unmarshal(b, localQueryStatus(qs)); err != nil {
		return err
	}

	// Set to a proper time.Duration value interpreting the server response
	// value in microseconds.
	qs.ElapsedTime = qs.ElapsedTime * time.Microsecond

	return nil
}

// Entry is an event that matched a query and is thus part of the result set.
type Entry struct {
	// Time is the time the event occurred. Matches SysTime if not specified
	// during ingestion.
	Time time.Time `json:"_time"`
	// SysTime is the time the event was recorded on the server.
	SysTime time.Time `json:"_sysTime"`
	// RowID is the unique ID of the event row.
	RowID string `json:"_rowId"`
	// Data contains the raw data of the event (with filters and aggregations
	// applied).
	Data map[string]interface{} `json:"data"`
}

// Timeseries are queried time series.
type Timeseries struct {
	// Series are the intervals that build a time series.
	Series []Interval `json:"series"`
	// Totals of the time series.
	Totals []EntryGroup `json:"totals"`
}

// Interval is the interval of queried time series.
type Interval struct {
	// StartTime of the interval.
	StartTime time.Time `json:"startTime"`
	// EndTime of the interval.
	EndTime time.Time `json:"endTime"`
	// Groups of the interval.
	Groups []EntryGroup `json:"groups"`
}

// EntryGroup is a group of queried event.
type EntryGroup struct {
	// ID of the group.
	ID uint64 `json:"id"`
	// Group ...
	// TODO(lukasmalkmus): What is this?
	Group map[string]interface{} `json:"group"`
	// Aggregations of the group.
	Aggregations []EntryGroupAgg `json:"aggregations"`
}

// EntryGroupAgg is an aggregation which is part of a group of queried events.
type EntryGroupAgg struct {
	// Op is the aggregations operation.
	Op AggregationOp `json:"op"`
	// Value is the result value of the aggregation.
	Value interface{} `json:"value"`
}

// DatasetCreateRequest is a request used to create a dataset.
type DatasetCreateRequest struct {
	// Name of the dataset to create. Restricted to 128 bytes and can not
	// contain the "axiom-" prefix.
	// TODO(lukasmalkmus): Clarify naming constraints.
	Name string `json:"name"`
	// Description of the dataset to create.
	Description string `json:"description"`
}

// DatasetUpdateRequest is a request used to update a dataset.
type DatasetUpdateRequest struct {
	// Description of the dataset to update.
	Description string `json:"description"`
}

// IngestOptions specifies the parameters for the Ingest and IngestEvents method
// of the Datasets service.
type IngestOptions struct {
	// TimestampField defines a custom field to extract the ingestion timestamp
	// from. Defaults to `_time`.
	TimestampField string `url:"timestamp-field,omitempty"`
	// TimestampFormat defines a custom format for the TimestampField.
	// The reference time is `Mon Jan 2 15:04:05 -0700 MST 2006`, as specified
	// in https://pkg.go.dev/time/?tab=doc#Parse.
	TimestampFormat string `url:"timestamp-format,omitempty"`
}

// QueryOptions specifies the parameters for the Query method of the Datasets
// service.
type QueryOptions struct {
	// StreamingDuration of a query.
	StreamingDuration time.Duration `url:"streaming-duration,omitempty"`
	// NoCache omits the query cache.
	NoCache bool `url:"no-cache,omitempty"`
}

// DatasetsService handles communication with the dataset related operations of
// the Axiom API.
//
// Axiom API Reference: /api/v1/datasets
type DatasetsService service

// Stats returns detailed statistics about all available datasets. This
// operation is more expenssive and listing the datasets and then getting the
// information of a specific dataset is preferred, when no aggregated
// statistics across all datasets are needed.
func (s *DatasetsService) Stats(ctx context.Context) (*DatasetStats, error) {
	path := s.basePath + "/_stats"

	var res *DatasetStats
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// List all available datasets.
func (s *DatasetsService) List(ctx context.Context) ([]*Dataset, error) {
	var res []*Dataset
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a dataset by id.
func (s *DatasetsService) Get(ctx context.Context, id string) (*Dataset, error) {
	path := s.basePath + "/" + id

	var res Dataset
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a dataset with the given properties.
func (s *DatasetsService) Create(ctx context.Context, req DatasetCreateRequest) (*Dataset, error) {
	var res Dataset
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the dataset identified by the given id with the given properties.
func (s *DatasetsService) Update(ctx context.Context, id string, req DatasetUpdateRequest) (*Dataset, error) {
	path := s.basePath + "/" + id

	var res Dataset
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the dataset identified by the given id.
func (s *DatasetsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}

// Info retrieves the information of the dataset identified by its id.
func (s *DatasetsService) Info(ctx context.Context, id string) (*DatasetInfo, error) {
	path := s.basePath + "/" + id + "/info"

	var res DatasetInfo
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Ingest data into the dataset identified by its id. If the dataset doesn't
// exist, it will be created. The given data will be flattened, thus there are
// some restrictions on the field names (JSON object keys):
//
// * Not more than 200 bytes (not characters!)
// * UTF-8 compatible
// * "_time" and "_source" are reserved
// * The ingestion content type must be one of JSON, NDJSON or CSV and the input
//   must be formatted accordingly
// TODO(lukasmalkmus): Review the restrictions.
func (s *DatasetsService) Ingest(ctx context.Context, id string, r io.Reader, typ ContentType, enc ContentEncoding, opts IngestOptions) (*IngestStatus, error) {
	path, err := addOptions(s.basePath+"/"+id+"/ingest", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, r)
	if err != nil {
		return nil, err
	}

	switch typ {
	case JSON, NDJSON, CSV:
		req.Header.Set("content-type", typ.String())
	default:
		return nil, ErrUnknownContentType
	}

	switch enc {
	case Identity:
	case GZIP:
		req.Header.Set("content-encoding", enc.String())
	default:
		return nil, ErrUnknownContentEncoding
	}

	var res IngestStatus
	if err = s.client.do(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Ingest events into the dataset identified by its id. If the dataset doesn't
// exist, it will be created. The given data will be flattened, thus there are
// some restrictions on the field names (JSON object keys):
//
// * Not more than 200 bytes (not characters!)
// * UTF-8 compatible
// * "_time" and "_source" are reserved
// * The ingestion content type must be one of JSON, NDJSON or CSV and the input
//   must be formatted accordingly
// TODO(lukasmalkmus): Review the restrictions.
func (s *DatasetsService) IngestEvents(ctx context.Context, id string, opts IngestOptions, events ...Event) (*IngestStatus, error) {
	if len(events) == 0 {
		return &IngestStatus{}, nil
	}

	path, err := addOptions(s.basePath+"/"+id+"/ingest", opts)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		// Does not fail with a valid compression level.
		gzw, _ := gzip.NewWriterLevel(pw, gzip.BestSpeed)

		var (
			enc    = json.NewEncoder(gzw)
			encErr error
		)
		for _, event := range events {
			if encErr = enc.Encode(event); encErr != nil {
				break
			}
		}

		_ = gzw.Close()
		_ = pw.CloseWithError(encErr)
	}()

	req, err := s.client.newRequest(ctx, http.MethodPost, path, pr)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", NDJSON.String())
	req.Header.Set("content-encoding", GZIP.String())

	var res IngestStatus
	if err = s.client.do(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Query executes the given query on the dataset identified by its id.
func (s *DatasetsService) Query(ctx context.Context, id string, query Query, opts QueryOptions) (*QueryResult, error) {
	path, err := addOptions(s.basePath+"/"+id+"/query", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, query)
	if err != nil {
		return nil, err
	}

	var res QueryResult
	if err = s.client.do(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GZIPStreamer returns an io.Reader that gzip compresses the data in reads from
// the provided reader using the specified compression level.
func GZIPStreamer(r io.Reader, level int) (io.Reader, error) {
	pr, pw := io.Pipe()

	gzw, err := gzip.NewWriterLevel(pw, level)
	if err != nil {
		return nil, err
	}

	go func() {
		_, err := io.Copy(gzw, r)
		_ = gzw.Close()
		_ = pw.CloseWithError(err)
	}()

	return pr, nil
}
