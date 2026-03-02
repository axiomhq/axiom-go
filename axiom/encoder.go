package axiom

import (
	"fmt"
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// ContentEncoder is a function that wraps a given reader with encoding
// functionality and returns that enhanced reader. The content type of the
// encoded content must obviously be accepted by the server.
//
// The built-in encoders returned by [GzipEncoder], [GzipEncoderWithLevel],
// [ZstdEncoder] and [ZstdEncoderWithLevel] pool compression writers internally
// to amortize allocation
// costs. Use [NewPooledEncoder] to create a pooled encoder for a custom
// compression writer.
//
// See [GzipEncoder] and [ZstdEncoder] for implementation reference.
type ContentEncoder func(io.Reader) (io.Reader, error)

// WriteResetter is implemented by compression writers that support pooled
// reuse. Both [zstd.Encoder] and [gzip.Writer] satisfy this interface.
// Reset must discard all buffered state and configure the writer to write
// compressed output to w. Close must flush any remaining compressed data.
type WriteResetter interface {
	io.WriteCloser
	Reset(w io.Writer)
}

// encoderPool is a type-safe pool for compression writers.
type encoderPool[T WriteResetter] struct {
	pool sync.Pool
}

func newEncoderPool[T WriteResetter](newFunc func() T) *encoderPool[T] {
	return &encoderPool[T]{
		pool: sync.Pool{
			New: func() any { return newFunc() },
		},
	}
}

func (p *encoderPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *encoderPool[T]) Put(v T) {
	p.pool.Put(v)
}

// zstdPools holds per-level pools for zstd writers. The array is indexed by
// zstdPoolIndex, covering all valid zstd levels from SpeedFastest (1) through
// SpeedBestCompression (4).
var zstdPools [zstd.SpeedBestCompression - zstd.SpeedFastest + 1]*encoderPool[*zstd.Encoder]

func init() {
	for level := zstd.SpeedFastest; level <= zstd.SpeedBestCompression; level++ {
		l := level
		zstdPools[zstdPoolIndex(l)] = newEncoderPool(func() *zstd.Encoder {
			w, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(l))
			if err != nil {
				panic("zstd: failed to create writer: " + err.Error())
			}
			return w
		})
	}
}

func zstdPoolIndex(level zstd.EncoderLevel) int {
	return int(level - zstd.SpeedFastest)
}

// gzipPools holds per-level pools for gzip writers. The array is indexed by
// gzipPoolIndex, covering all valid gzip levels from HuffmanOnly (-2) through
// BestCompression (9).
var gzipPools [gzip.BestCompression - gzip.HuffmanOnly + 1]*encoderPool[*gzip.Writer]

func init() {
	for level := gzip.HuffmanOnly; level <= gzip.BestCompression; level++ {
		l := level
		gzipPools[gzipPoolIndex(l)] = newEncoderPool(func() *gzip.Writer {
			w, err := gzip.NewWriterLevel(nil, l)
			if err != nil {
				panic("gzip: failed to create writer: " + err.Error())
			}
			return w
		})
	}
}

func gzipPoolIndex(level int) int {
	return level - gzip.HuffmanOnly
}

// NewPooledEncoder creates a [ContentEncoder] that pools compression writers
// to amortize allocation costs. Call it once and reuse the returned encoder.
//
// The newWriter function creates fresh instances when the pool is empty.
// Each writer is Reset before reuse and returned to the pool after Close.
func NewPooledEncoder[T WriteResetter](newWriter func() T) ContentEncoder {
	return pooledContentEncoder(newEncoderPool(newWriter))
}

func pooledContentEncoder[T WriteResetter](pool *encoderPool[T]) ContentEncoder {
	return func(r io.Reader) (io.Reader, error) {
		pr, pw := io.Pipe()
		w := pool.Get()
		w.Reset(pw)
		go func() {
			_, err := io.Copy(w, r)
			if closeErr := w.Close(); closeErr != nil {
				if err == nil {
					err = closeErr
				}
			} else {
				pool.Put(w)
			}
			_ = pw.CloseWithError(err)
		}()
		return pr, nil
	}
}

// GzipEncoder returns a content encoder that gzip compresses the data it reads
// from the provided reader. The compression level defaults to [gzip.BestSpeed].
// Writers are pooled internally to avoid repeated allocations.
func GzipEncoder() ContentEncoder {
	return GzipEncoderWithLevel(gzip.BestSpeed)
}

// GzipEncoderWithLevel returns a content encoder that gzip compresses data
// using the specified compression level. Writers are pooled internally per
// compression level.
func GzipEncoderWithLevel(level int) ContentEncoder {
	idx := gzipPoolIndex(level)
	if idx < 0 || idx >= len(gzipPools) || gzipPools[idx] == nil {
		return func(_ io.Reader) (io.Reader, error) {
			return nil, fmt.Errorf("unsupported gzip compression level: %d", level)
		}
	}
	return pooledContentEncoder(gzipPools[idx])
}

// ZstdEncoder returns a content encoder that zstd compresses the data it reads
// from the provided reader. The compression level defaults to
// [zstd.SpeedDefault]. Writers are pooled internally to avoid the ~4 MB
// allocation cost per [zstd.NewWriter] call.
func ZstdEncoder() ContentEncoder {
	return ZstdEncoderWithLevel(zstd.SpeedDefault)
}

// ZstdEncoderWithLevel returns a content encoder that zstd compresses data
// using the specified compression level. Writers are pooled internally per
// compression level.
func ZstdEncoderWithLevel(level zstd.EncoderLevel) ContentEncoder {
	idx := zstdPoolIndex(level)
	if idx < 0 || idx >= len(zstdPools) || zstdPools[idx] == nil {
		return func(_ io.Reader) (io.Reader, error) {
			return nil, fmt.Errorf("unsupported zstd compression level: %d", level)
		}
	}
	return pooledContentEncoder(zstdPools[idx])
}
