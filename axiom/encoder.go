package axiom

import (
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// ContentEncoder is a function that wraps a given `io.Reader` with encoding
// functionality and returns that enhanced reader. The content type of the
// encoded content must obviously be accepted by the server.
//
// See `GzipEncoder` and `ZstdEncoder` for implementation reference.
type ContentEncoder func(io.Reader) (io.Reader, error)

// GzipEncoder is a `ContentEncoder` that gzip compresses the data it reads
// from the provided reader. The compression level defaults to `gzip.BestSpeed`.
func GzipEncoder(r io.Reader) (io.Reader, error) {
	return GzipEncoderWithLevel(gzip.BestSpeed)(r)
}

// GzipEncoderWithLevel returns a `ContentEncoder` that gzip compresses data
// using the specified compression level.
func GzipEncoderWithLevel(level int) ContentEncoder {
	return func(r io.Reader) (io.Reader, error) {
		pr, pw := io.Pipe()

		gzw, err := gzip.NewWriterLevel(pw, level)
		if err != nil {
			return nil, err
		}

		go func() {
			_, err := io.Copy(gzw, r)
			if closeErr := gzw.Close(); err == nil {
				// If we have no error from copying but from closing, capture
				// that one.
				err = closeErr
			}
			_ = pw.CloseWithError(err)
		}()

		return pr, nil
	}
}

// ZstdEncoder is a `ContentEncoder` that zstd compresses the data it reads
// from the provided reader.
func ZstdEncoder(r io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()

	zsw, err := zstd.NewWriter(pw)
	if err != nil {
		return nil, err
	}

	go func() {
		_, err := io.Copy(zsw, r)
		if closeErr := zsw.Close(); err == nil {
			// If we have no error from copying but from closing, capture that
			// one.
			err = closeErr
		}
		_ = pw.CloseWithError(err)
	}()

	return pr, nil
}
