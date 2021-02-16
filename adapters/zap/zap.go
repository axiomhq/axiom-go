package zap

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
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

// An Option modifies the behaviour of the Axiom WriteSyncer.
type Option func(*WriteSyncer) error

// LevelEnabler sets the level enabler that the Axiom WriteSyncer will us to
// determine if logs will be shipped to Axiom.
func LevelEnabler(levelEnabler zapcore.LevelEnabler) Option {
	return func(c *WriteSyncer) error {
		c.levelEnabler = levelEnabler
		return nil
	}
}

// IngestOptions specifies the ingestion options to use for ingesting the logs.
func IngestOptions(opts axiom.IngestOptions) Option {
	return func(c *WriteSyncer) error {
		c.ingestOptions = opts
		return nil
	}
}

// WriteSyncer implements a zapcore.WriteSyncer used for shipping logs to Axiom.
type WriteSyncer struct {
	client      *axiom.Client
	datasetName string

	levelEnabler  zapcore.LevelEnabler
	ingestOptions axiom.IngestOptions

	buf    bytes.Buffer
	bufMtx sync.Mutex
}

// New creates a new zapcore.Core configured to talk to the specified Axiom
// deployment and authenticating with the given access token. An ingest token is
// sufficient enough. The logs will be ingested into the specified dataset.
// Additional options can be supplied to configure the core.
func New(baseURL, accessToken, orgID, datasetName string, options ...Option) (zapcore.Core, error) {
	client, err := axiom.NewClient(baseURL, accessToken, orgID)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName)
}

// NewCloud is like New() but configures the core to talk to Axiom Cloud.
func NewCloud(accessToken, datasetName string, options ...Option) (zapcore.Core, error) {
	client, err := axiom.NewCloudClient(accessToken)
	if err != nil {
		return nil, err
	}

	return NewWithClient(client, datasetName)
}

// NewWithClient behaves like New() but utilizes an already configured
// axiom.Client to talk to a deployment.
func NewWithClient(client *axiom.Client, datasetName string, options ...Option) (zapcore.Core, error) {
	ws := &WriteSyncer{
		client:      client,
		datasetName: datasetName,

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

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	return zapcore.NewCore(encoder, ws, ws.levelEnabler), nil
}

// Write implements zapcore.WriteSyncer.
func (ws *WriteSyncer) Write(p []byte) (n int, err error) {
	ws.bufMtx.Lock()
	defer ws.bufMtx.Unlock()

	return ws.buf.Write(p)
}

// Sync implements zapcore.WriteSyncer.
func (ws *WriteSyncer) Sync() error {
	// Best effort context timeout. A Sync() should never take that long.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ws.bufMtx.Lock()
	defer ws.bufMtx.Unlock()

	if ws.buf.Len() == 0 {
		return nil
	}

	// Make sure to reset the buffer.
	defer ws.buf.Reset()

	r, err := axiom.GZIPStreamer(&ws.buf, gzip.BestSpeed)
	if err != nil {
		return err
	}

	res, err := ws.client.Datasets.Ingest(ctx, ws.datasetName, r, axiom.NDJSON, axiom.GZIP, ws.ingestOptions)
	if err != nil {
		return err
	} else if res.Failed > 0 {
		// Best effort on notifying the user about the ingest failure.
		return fmt.Errorf("event at %s failed to ingest: %s",
			res.Failures[0].Timestamp, res.Failures[0].Error)
	}

	return nil
}
