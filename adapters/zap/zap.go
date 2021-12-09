package zap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/axiomhq/axiom-go/axiom"
)

var _ zapcore.WriteSyncer = (*WriteSyncer)(nil)

// Matches https://github.com/uber-go/zap/blob/master/config.go#L98 but modifies
// the timestamp field to be Axiom compatible.
var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        axiom.TimestampField, // Modified
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.RFC3339NanoTimeEncoder, // Modified
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// ErrMissingDatasetName is raised when a dataset name is not provided. Set it
// manually using the SetDataset option or export `AXIOM_DATASET`.
var ErrMissingDatasetName = errors.New("missing dataset name")

// An Option modifies the behaviour of the Axiom WriteSyncer.
type Option func(*WriteSyncer) error

// SetClient specifies the Axiom client to use for ingesting the logs.
func SetClient(client *axiom.Client) Option {
	return func(ws *WriteSyncer) error {
		ws.client = client
		return nil
	}
}

// SetClientOptions specifies the Axiom client options to pass to
// `axiom.NewClient()`. `axiom.NewClient()` is only called if no client was
// specified by the `SetClient` option.
func SetClientOptions(options ...axiom.Option) Option {
	return func(ws *WriteSyncer) error {
		ws.clientOptions = options
		return nil
	}
}

// SetDataset specifies the dataset to ingest the logs into. Can also be
// specified using the `AXIOM_DATASET` environment variable.
func SetDataset(datasetName string) Option {
	return func(ws *WriteSyncer) error {
		ws.datasetName = datasetName
		return nil
	}
}

// SetIngestOptions specifies the ingestion options to use for ingesting the
// logs.
func SetIngestOptions(opts axiom.IngestOptions) Option {
	return func(ws *WriteSyncer) error {
		ws.ingestOptions = opts
		return nil
	}
}

// SetLevelEnabler sets the level enabler that the Axiom WriteSyncer will us to
// determine if logs will be shipped to Axiom.
func SetLevelEnabler(levelEnabler zapcore.LevelEnabler) Option {
	return func(ws *WriteSyncer) error {
		ws.levelEnabler = levelEnabler
		return nil
	}
}

// WriteSyncer implements a `zapcore.WriteSyncer` used for shipping logs to
// Axiom.
type WriteSyncer struct {
	client      *axiom.Client
	datasetName string

	clientOptions []axiom.Option
	ingestOptions axiom.IngestOptions
	levelEnabler  zapcore.LevelEnabler

	buf    bytes.Buffer
	bufMtx sync.Mutex
}

// New creates a new `zapcore.Core` configured to ingest logs to the Axiom
// deployment and dataset as specified by the environment. Refer to
// `axiom.NewClient()` for more details on how configuring the Axiom deployment
// works or pass the `SetClient()` option to pass a custom client or
// `SetClientOptions()` to control the Axiom client creation. To specify the
// dataset set `AXIOM_DATASET` or use the `SetDataset()` option.
//
// An ingest token is sufficient enough. Additional options can be supplied to
// configure the `zapcore.Core`.
func New(options ...Option) (zapcore.Core, error) {
	ws := &WriteSyncer{
		levelEnabler: zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return true
		}),
	}

	// Apply supplied options.
	for _, option := range options {
		if err := option(ws); err != nil {
			return nil, err
		}
	}

	// Create client, if not set.
	if ws.client == nil {
		var err error
		if ws.client, err = axiom.NewClient(ws.clientOptions...); err != nil {
			return nil, err
		}
	}

	// When the dataset name is not set, use `AXIOM_DATASET`.
	if ws.datasetName == "" {
		if ws.datasetName = os.Getenv("AXIOM_DATASET"); ws.datasetName == "" {
			return nil, ErrMissingDatasetName
		}
	}

	enc := zapcore.NewJSONEncoder(encoderConfig)

	return zapcore.NewCore(enc, ws, ws.levelEnabler), nil
}

// Write implements zapcore.WriteSyncer.
func (ws *WriteSyncer) Write(p []byte) (n int, err error) {
	ws.bufMtx.Lock()
	defer ws.bufMtx.Unlock()

	return ws.buf.Write(p)
}

// Sync implements zapcore.WriteSyncer.
func (ws *WriteSyncer) Sync() error {
	// Best effort context timeout. A `Sync()` should never take that long.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ws.bufMtx.Lock()
	defer ws.bufMtx.Unlock()

	if ws.buf.Len() == 0 {
		return nil
	}

	// Make sure to reset the buffer.
	defer ws.buf.Reset()

	r, err := axiom.GzipEncoder(&ws.buf)
	if err != nil {
		return err
	}

	res, err := ws.client.Datasets.Ingest(ctx, ws.datasetName, r, axiom.NDJSON, axiom.Gzip, ws.ingestOptions)
	if err != nil {
		return err
	} else if res.Failed > 0 {
		// Best effort on notifying the user about the ingest failure.
		return fmt.Errorf("event at %s failed to ingest: %s",
			res.Failures[0].Timestamp, res.Failures[0].Error)
	}

	return nil
}
