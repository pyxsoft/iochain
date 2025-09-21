package iochain

import "io"

// ReaderToWriter is a type that links an io.Reader source to an io.Writer target for data streaming and copying.
// It and writes data to the target every time a read is performed.
// The target writer can optionally implement io.Closer for resource cleanup upon closure.
type ReaderToWriter struct {
	src    io.Reader
	target io.Writer
}

// NewReaderToWriter creates a new ReaderToWriter instance with the specified io.Writer as the target destination.
func NewReaderToWriter(w io.Writer) *ReaderToWriter {
	return &ReaderToWriter{target: w}
}

func (r *ReaderToWriter) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	if n > 0 {
		_, _ = r.target.Write(p[:n]) // write copy, ignore write error
	}
	return n, err
}

func (r *ReaderToWriter) Reset(src io.Reader) error {
	r.src = src
	return nil
}

// Close closes the underlying writer if it implements io.Closer.
func (r *ReaderToWriter) Close() error {
	if closer, ok := r.target.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
