package iochain

import (
	"errors"
	"io"
	"sync"
)

// ResettableWriter is an io.Writer that can be reset to wrap another writer.
type ResettableWriter interface {
	io.Writer
	Reset(w io.Writer)
}

// Flusher is implemented by writers that support flushing their internal buffer.
type Flusher interface {
	Flush() error
}

// StackWriter manages a stack of writers, each one writing to the previous.
type StackWriter struct {
	mu      sync.Mutex
	base    io.Writer
	writers []io.Writer // from base to top
}

// NewStackWriter creates a StackWriter starting with the base writer.
func NewStackWriter(base io.Writer) (*StackWriter, error) {
	if base == nil {
		return nil, errors.New("base writer cannot be nil")
	}
	return &StackWriter{
		base:    base,
		writers: []io.Writer{base},
	}, nil
}

// AddWriter wraps the current top writer with a new ResettableWriteCloser.
func (m *StackWriter) AddWriter(w ResettableWriter) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.writers[len(m.writers)-1]
	w.Reset(prev)

	m.writers = append(m.writers, w)
	return nil
}

// Write writes to the top-most writer in the stack.
func (m *StackWriter) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.writers) == 0 {
		return 0, io.ErrClosedPipe
	}
	return m.writers[len(m.writers)-1].Write(p)
}

// Flush calls Flush() on all writers from top to base if they implement Flusher.
func (m *StackWriter) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for i := len(m.writers) - 1; i >= 0; i-- {
		if flusher, ok := m.writers[i].(Flusher); ok {
			if err := flusher.Flush(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Close closes all writers from top to base.
func (m *StackWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for i := len(m.writers) - 1; i >= 0; i-- {
		if closer, ok := m.writers[i].(io.Closer); ok {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	m.writers = nil
	return firstErr
}

// FlushAndClose flushes all writers (if supported) and then closes them.
func (m *StackWriter) FlushAndClose() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error

	// Flush from top to base
	for i := len(m.writers) - 1; i >= 0; i-- {
		if flusher, ok := m.writers[i].(Flusher); ok {
			if err := flusher.Flush(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close from top to base
	for i := len(m.writers) - 1; i >= 0; i-- {
		if closer, ok := m.writers[i].(io.Closer); ok {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	m.writers = nil
	return firstErr
}
