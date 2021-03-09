package axiom

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
	"unicode"

	"github.com/axiomhq/axiom-go/axiom/query"
)

//go:generate ../bin/stringer -type=ContentType,ContentEncoding -linecomment -output=datasets_string.go

// TimestampField is the default field the server looks for a time to use as
// ingestion time. If not present, the server will set the ingestion time by
// itself.
const TimestampField = "_time"

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
	// NDJSON treats the data as newline delimited JSON objects. Preferred
	// format.
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
	// TypeHint is a hint regarding the datatype of the field.
	TypeHint string `json:"typeHint"`
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

// TrimResult is the result of a trim operation.
type TrimResult struct {
	// BlocksDeleted is the amount of blocks deleted by the trim operation.
	BlocksDeleted int `json:"numDeleted"`
}

// HistoryQuery represents a query stored inside the query history.
type HistoryQuery struct {
	// ID is the unique id of the starred query.
	ID string `json:"id"`
	// Kind of the starred query.
	Kind QueryKind `json:"kind"`
	// Dataset the starred query belongs to.
	Dataset string `json:"dataset"`
	// Owner is the ID of the starred queries owner. Can be a user or team ID.
	Owner string `json:"who"`
	// Query is the actual query.
	Query query.Query `json:"query"`
	// Created is the time the starred query was created at.
	Created time.Time `json:"created"`
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

// DatasetCreateRequest is a request used to create a dataset.
type DatasetCreateRequest struct {
	// Name of the dataset to create. Restricted to 128 bytes of [a-zA-Z0-9] and
	// special characters "-", "_" and ".". Special characters cannot be a
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

type datasetTrimRequest struct {
	// MaxDuration marks the oldest timestamp an event can have before getting
	// deleted.
	MaxDuration string `json:"maxDuration"`
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

// Trim the dataset identified by its id to a given length. The max duration
// given will mark the oldest timestamp an event can have. Older ones will be
// deleted from the dataset.
func (s *DatasetsService) Trim(ctx context.Context, id string, maxDuration time.Duration) (*TrimResult, error) {
	req := datasetTrimRequest{
		MaxDuration: maxDuration.String(),
	}

	path := s.basePath + "/" + id + "/trim"

	var res TrimResult
	if err := s.client.call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// History retrieves the query stored inside the query history dataset
// identified by its id.
func (s *DatasetsService) History(ctx context.Context, id string) (*HistoryQuery, error) {
	path := s.basePath + "/_history/" + id

	var res HistoryQuery
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Ingest data into the dataset identified by its id. If the dataset doesn't
// exist, it will be created. The given data will be flattened, thus there are
// some restrictions on the field names (JSON object keys):
//
// * Not more than 200 bytes
// * Valid UTF-8
// * "_time" must conform to a valid timestamp or not be present at all
// * If "_time" is not present, the server will assign a timestamp
// * The ingestion content type must be one of JSON, NDJSON or CSV and the input
//   must be formatted accordingly
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

// IngestEvents ingests events into the dataset identified by its id. If the
// dataset doesn't exist, it will be created. The given data will be flattened,
// thus there are some restrictions on the field names (JSON object keys):
//
// * Not more than 200 bytes
// * Valid UTF-8
// * "_time" must conform to a valid timestamp or not be present at all
// * If "_time" is not present, the server will assign a timestamp
// * The ingestion content type must be one of JSON, NDJSON or CSV and the input
//   must be formatted accordingly
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
func (s *DatasetsService) Query(ctx context.Context, id string, q query.Query, opts query.Options) (*query.Result, error) {
	path, err := addOptions(s.basePath+"/"+id+"/query", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, q)
	if err != nil {
		return nil, err
	}

	var res query.Result
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

// DetectContentType detects the content type of an io.Reader's data. The
// returned io.Reader must be used instead of the passed one. Compressed content
// is not detected.
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
		} else if !unicode.IsSpace(c) {
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
