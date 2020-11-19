package axiom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	// ErrUnknownContentType is raised when the given content type is not valid.
	ErrUnknownContentType = errors.New("unknown content type")
	// ErrUnknownContentEncoding is raised when the given content encoding is
	// not valid.
	ErrUnknownContentEncoding = errors.New("unknown content encoding")
)

// ContentType describes the content type of the data to ingest.
type ContentType string

const (
	// JSON treats the data as JSON array.
	JSON ContentType = "application/json"
	// NDJSON treats the data as newline delimited JSON objects. Preferred as it
	// is faster than JSON array based ingestion.
	// TODO(lukasmalkmus): Is this still true?
	NDJSON ContentType = "application/x-ndjson"
	// CSV treats the data as CSV content.
	CSV ContentType = "text/csv"
)

// ContentEncoding describes the content encoding of the data to ingest.
type ContentEncoding string

const (
	// Identity marks the data as not being encoded.
	Identity ContentEncoding = ""
	// GZIP marks the data as being gzip encoded.
	GZIP ContentEncoding = "gzip"
)

// Dataset represents an Axiom dataset.
type Dataset struct {
	// ID is the unique ID of the dataset.
	ID string `json:"id"`
	// Name is the unique name of the dataset.
	Name string `json:"name"`
	// Description is the description of the dataset.
	Description string `json:"description"`
	// Create is the time the dataset was created at.
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
	// DisplayName is the human displayable name of a dataset.
	DisplayName string `json:"displayName"`
	// NumBlocks is the number of blocks of a dataset.
	NumBlocks uint64 `json:"numBlocks"`
	// NumEvents is the number of events of a dataset.
	NumEvents uint64 `json:"numEvents"`
	// NumFields is the number of fields of a dataset.
	NumFields uint32 `json:"numFields"`
	// InputBytes is the amount of data stored in a dataset.
	InputBytes uint64 `json:"inputBytes"`
	// InputBytesHuman is the amount of data stored in a dataset formatted in a
	// human readable format.
	InputBytesHuman string `json:"inputBytesHuman"`
	// CompressedBytes is the amount of compressed data stored in a dataset.
	CompressedBytes uint64 `json:"compressedBytes"`
	// CompressedBytesHuman is the amount of compressed data stored in a dataset
	// formatted in a human readable format.
	CompressedBytesHuman string `json:"compressedBytesHuman"`
	// MinTime is the time of the oldest event stored in a dataset.
	MinTime time.Time `json:"minTime"`
	// MaxTime is the time of the newest event stored in a dataset.
	MaxTime time.Time `json:"maxTime"`
	// Fields are the fields of a dataset.
	Fields []Field `json:"fields"`
}

// CreateDatasetRequest is the request body used to create a dataset.
type CreateDatasetRequest struct {
	// Name of the dataset to create.
	Name string `json:"name"`
	// Description of the dataset to create.
	Description string `json:"description"`
}

// IngestOptions are the request query url parameters for event ingestion.
type IngestOptions struct {
	// Define a custom field for the timestamps, defaults to `_time`.
	TimestampField string `url:"timestamp-field"`
	// TimestampFormat defines a custom format for the timestamps.
	// The reference time is `Mon Jan 2 15:04:05 -0700 MST 2006`, as specified
	// in https://pkg.go.dev/time/?tab=doc#Parse
	TimestampFormat string `url:"timestamp-format"`
}

// IngestFailure describes the ingestion failure of a single event.
type IngestFailure struct {
	// Timestamp of the event that failed to ingest.
	Timestamp time.Time `json:"timestamp"`
	// Error that made the event fail to ingest.
	Error string `json:"error"`
}

// IngestResponse is the response returned after event ingestion.
type IngestResponse struct {
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

// DatasetsService bundles all the operations on Axiom datasets.
type DatasetsService interface {
	// List all available datasets.
	List(ctx context.Context) ([]*Dataset, error)
	// Get a dataset by id.
	Get(ctx context.Context, datasetID string) (*Dataset, error)
	// Info retrieves the information of a dataset identified by its id.
	Info(ctx context.Context, datasetID string) (*DatasetInfo, error)
	// Create a dataset with the given properties.
	Create(ctx context.Context, opts CreateDatasetRequest) (*Dataset, error)
	// Delete the dataset identified by its id.
	Delete(ctx context.Context, datasetID string) error
	// Ingest data into the dataset identified by its id. If the dataset doesn't
	// exist, it will be created. The given data will be flattened, thus there
	// are some restrictions on the field names (JSON object keys):
	//
	// * Not more than 200 bytes (not characters!)
	// * UTF-8 compatible
	// * "_time" and "_source" are reserved
	// * The ingestion content type must be one of JSON, NDJSON or CSV and the
	//   input must be formatted accordingly.
	// TODO(lukasmalkmus): Review the restrictions.
	Ingest(ctx context.Context, datasetID string, r io.Reader, typ ContentType, enc ContentEncoding, opts IngestOptions) (*IngestResponse, error)
}

var _ DatasetsService = (*datasetsService)(nil)

type datasetsService struct {
	client *Client
}

// List all available datasets.
func (s *datasetsService) List(ctx context.Context) ([]*Dataset, error) {
	path := "/api/v1/datasets"

	var res []*Dataset
	if _, err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a dataset by id.
func (s *datasetsService) Get(ctx context.Context, id string) (*Dataset, error) {
	path := "/api/v1/datasets/" + id

	var res Dataset
	if _, err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Info retrieves the information of a dataset identified by its id.
func (s *datasetsService) Info(ctx context.Context, id string) (*DatasetInfo, error) {
	path := "/api/v1/datasets/" + id + "/info"

	var res DatasetInfo
	if _, err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a dataset with the given id.
func (s *datasetsService) Create(ctx context.Context, req CreateDatasetRequest) (*Dataset, error) {
	path := "/api/v1/datasets"

	var res Dataset
	if _, err := s.client.call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the dataset identified by its id.
func (s *datasetsService) Delete(ctx context.Context, id string) error {
	path := "/api/v1/datasets/" + id

	if _, err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}

// Ingest data into the dataset identified by its id. If the dataset doesn't
// exist, it will be created. The given data will be flattened, thus there are
// some restrictions on the field names (JSON object keys):
//
// * Not more than 200 bytes (not characters!)
// * UTF-8 compatible
// * "_time" and "_source" are reserved
// * The ingestion content type must be one of JSON, NDJSON or CSV and the input
//   must be formatted accordingly.
// TODO(lukasmalkmus): Review the restrictions.
func (s *datasetsService) Ingest(ctx context.Context, datasetID string, r io.Reader, typ ContentType, enc ContentEncoding, opts IngestOptions) (*IngestResponse, error) {
	path, err := addOptions("/api/v1/datasets/"+datasetID+"/ingest", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, r)
	if err != nil {
		return nil, err
	}

	switch typ {
	case JSON, NDJSON, CSV:
		req.Header.Set("Content-Type", string(typ))
	default:
		return nil, ErrUnknownContentType
	}

	switch enc {
	case Identity:
	case GZIP:
		req.Header.Set("Content-Encoding", string(enc))
	default:
		return nil, ErrUnknownContentEncoding
	}

	var res IngestResponse
	if _, err = s.client.do(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
