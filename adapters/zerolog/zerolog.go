package zerolog

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/rs/zerolog"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

var (
	_ = io.Writer(new(Writer))

	// ErrMissingDataset is raised when a dataset name is not provided. Set it
	// manually using the [SetDataset] option or export "AXIOM_DATASET".
	ErrMissingDataset = errors.New("missing dataset name")

	logger     = log.New(os.Stderr, "[AXIOM|ZEROLOG]", 0)
	loggerName = []byte(`"zerolog"`)
)

const (
	defaultBatchSize         = 10_000
	defaultMaxBufferCapacity = 1 << 20 // 1MB
	flushInterval            = time.Second
)

// Writer is a axiom events writer with std io.Writer interface.
type Writer struct {
	client  *axiom.Client
	dataset string

	clientOptions     []axiom.Option
	ingestOptions     []ingest.Option
	levels            map[zerolog.Level]struct{}
	maxBufferCapacity int

	byteCh    chan []byte
	closeOnce sync.Once
	closeCh   chan struct{}
}

// Write must not modify the slice data, even temporarily.
func (w *Writer) Write(data []byte) (int, error) {
	select {
	case <-w.closeCh:
	default:
		b := make([]byte, len(data))
		copy(b, data)
		w.byteCh <- b
	}

	return len(data), nil
}

func (w *Writer) Close() {
	w.closeOnce.Do(func() {
		close(w.byteCh)
		<-w.closeCh
	})
}

// Option configures axiom events writer.
type Option func(*Writer) error

// SetClient configures a custom axiom client.
func SetClient(client *axiom.Client) Option {
	return func(cfg *Writer) error {
		cfg.client = client
		return nil
	}
}

// SetLevels configures zerolog levels that have to be sent to Axiom.
func SetLevels(levels []zerolog.Level) Option {
	return func(cfg *Writer) error {
		for _, level := range levels {
			cfg.levels[level] = struct{}{}
		}
		return nil
	}
}

// SetDataset configures the axiom dataset name.
func SetDataset(dataset string) Option {
	return func(cfg *Writer) error {
		cfg.dataset = dataset
		return nil
	}
}

// SetClientOptions specifies the Axiom client options to pass to
// [axiom.NewClient] which is only called if no [axiom.Client] was specified by
// the [SetClient] option.
func SetClientOptions(options ...axiom.Option) Option {
	return func(cfg *Writer) error {
		cfg.clientOptions = options
		return nil
	}
}

// SetIngestOptions configures the axiom ingest options.
func SetIngestOptions(ingestOptions []ingest.Option) Option {
	return func(cfg *Writer) error {
		cfg.ingestOptions = ingestOptions
		return nil
	}
}

// SetMaxBufferCapacity configures the maximum buffer capacity in bytes. Buffers
// exceeding this capacity are released after flushing to prevent memory bloat
// from traffic spikes. Set to 0 to always release the buffer. Defaults to 1MB.
func SetMaxBufferCapacity(size int) Option {
	return func(cfg *Writer) error {
		if size < 0 {
			return errors.New("max buffer capacity cannot be negative")
		}
		cfg.maxBufferCapacity = size
		return nil
	}
}

// New creates a new Writer that ingests logs into Axiom. It automatically takes
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
// An API token with "ingest" permission is sufficient enough.
//
// A Writer needs to be closed properly to make sure all logs are sent by calling
// [Writer.Close].
func New(opts ...Option) (*Writer, error) {
	w := &Writer{
		levels:            make(map[zerolog.Level]struct{}),
		ingestOptions:     []ingest.Option{ingest.SetTimestampField(zerolog.TimestampFieldName), ingest.SetTimestampFormat(zerolog.TimeFieldFormat)},
		clientOptions:     []axiom.Option{},
		maxBufferCapacity: defaultMaxBufferCapacity,
		byteCh:            make(chan []byte, defaultBatchSize),
		closeCh:           make(chan struct{}),
	}

	// Apply supplied options.
	for _, option := range opts {
		if option == nil {
			continue
		}
		if err := option(w); err != nil {
			return nil, err
		}
	}

	if len(w.levels) == 0 {
		for _, level := range []zerolog.Level{zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel} {
			w.levels[level] = struct{}{}
		}
	}

	// Create client, if not set.
	if w.client == nil {
		var err error
		if w.client, err = axiom.NewClient(w.clientOptions...); err != nil {
			return nil, err
		}
	}

	// When the dataset name is not set, use "AXIOM_DATASET".
	if w.dataset == "" {
		w.dataset = os.Getenv("AXIOM_DATASET")
		if w.dataset == "" {
			return nil, ErrMissingDataset
		}
	}

	go w.runBackgroundJob()
	return w, nil
}

func (w *Writer) runBackgroundJob() {
	var (
		counter = 0
		buffer  = &bytes.Buffer{}
		encoder = axiom.ZstdEncoder()
	)

	flush := func() error {
		defer func() {
			counter = 0
			if buffer.Cap() > w.maxBufferCapacity {
				buffer = &bytes.Buffer{}
				return
			}
			buffer.Reset()
		}()

		if buffer.Len() == 0 {
			return nil
		}

		r, err := encoder(buffer)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()

		res, err := w.client.Ingest(ctx, w.dataset, r, axiom.NDJSON, axiom.Zstd, w.ingestOptions...)
		if err != nil {
			return err
		}
		if res.Failed > 0 {
			logger.Printf("event(s) [%v] at %s failed to ingest: %s\n", res.Failed, res.Failures[0].Timestamp, res.Failures[0].Error)
		}
		return nil
	}

	defer close(w.closeCh)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-w.byteCh:
			if !ok {
				if err := flush(); err != nil {
					logger.Printf("failed to ingest events: %s\n", err)
				}
				return
			}
			if len(data) == 0 {
				continue
			}

			lvlStr, err := jsonparser.GetUnsafeString(data, zerolog.LevelFieldName)
			if err != nil {
				logger.Printf("failed to retrieve level field name from data: %s\n", err)
				continue
			}

			lvl, err := zerolog.ParseLevel(lvlStr)
			if err != nil {
				logger.Printf("failed to parse level: %s\n", err)
				continue
			}

			if _, enabled := w.levels[lvl]; !enabled {
				continue
			}

			counter++

			data, _ = jsonparser.Set(data, loggerName, "logger")

			buffer.Write(data)

			if counter >= defaultBatchSize {
				if err := flush(); err != nil {
					logger.Printf("failed to ingest events: %s\n", err)
				}
			}
		case <-ticker.C:
			if err := flush(); err != nil {
				logger.Printf("failed to ingest events: %s\n", err)
			}
		}
	}
}
