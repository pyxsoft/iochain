package iochain

import "io"

// PassthroughReader wraps another io.Reader and just passes reads through.
type PassthroughReader struct {
	src io.Reader
}

// Read just delegates the call to the underlying reader.
func (r *PassthroughReader) Read(p []byte) (int, error) {
	return r.src.Read(p)
}

func (r *PassthroughReader) Close() error {
	if closer, ok := r.src.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func NewPasstroughReader(src io.Reader) *PassthroughReader {
	return &PassthroughReader{src: src}
}
