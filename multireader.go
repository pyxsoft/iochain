package iochain

import (
	"errors"
	"io"
	"sync"
)

// ResettableReader is an io.Reader that can be reset to read from another reader.
type ResettableReader interface {
	io.Reader
	Reset(r io.Reader) error
}

// MultiReader manages a stack of readers, each reading from the previous one.
type MultiReader struct {
	mu      sync.Mutex
	readers []io.Reader // from base to top
}

// NewReader creates a new MultiReader with a base reader.
func NewReader(base io.Reader) (*MultiReader, error) {
	if base == nil {
		return nil, errors.New("base reader cannot be nil")
	}
	return &MultiReader{
		readers: []io.Reader{base},
	}, nil
}

// AddReader wraps the current top reader with a new ResettableReader.
func (m *MultiReader) AddReader(r ResettableReader) error {
	if r == nil {
		return errors.New("reader cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.readers[len(m.readers)-1]
	if err := r.Reset(prev); err != nil {
		return err
	}

	m.readers = append(m.readers, r)
	return nil
}

// Read reads from the top-most reader in the chain.
func (m *MultiReader) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.readers) == 0 {
		return 0, io.EOF
	}
	return m.readers[len(m.readers)-1].Read(p)
}

// Close calls Close() on each reader from top to base if it implements io.Closer.
func (m *MultiReader) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for i := len(m.readers) - 1; i >= 0; i-- {
		if closer, ok := m.readers[i].(io.Closer); ok {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	m.readers = nil
	return firstErr
}
