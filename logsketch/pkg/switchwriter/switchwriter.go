// Package switchwriter implements managed io.Writers that provide a destination stream
// which can be changed dynamically.  The destination stream can be disabled,
// (see SwitchTo(nil)), in which case all Write()s to it are silently discarded.
//
// All writers are initially created disabled.
//
// Writers contain a mutex, and so must not be copied.

package switchwriter

import (
	"io"
	"sync"
)

// TODO  ? Make unexpeorted to avoid accidental copies ?
//
type Writer struct {
	sync.Mutex
	dest io.Writer // The current destination stream.
}

// A compile-time check that Writer implements io.Writer.
//
var _ io.Writer = New()

// New creates a new, initially disabled, SwitchWriter.
//
func New() *Writer {
	return new(Writer)
}

// SwitchTo(w) switches the destination of future Write()s.
// If w is nil future Write()s will be silently discarded.
//
func (sw *Writer) SwitchTo(w io.Writer) {
	sw.Lock()
	defer sw.Unlock()

	sw.dest = w // Change the destination for future Write()s.
}

// Write(buf) writes to the current destination unless
// the current destination is nil. In that case Write() writes nothing
// but pretends that the write was done.
//
func (sw *Writer) Write(buf []byte) (n int, err error) {
	sw.Lock()
	defer sw.Unlock()

	switch sw.dest {
	case nil:
		return len(buf), nil // Don't write, but pretend all of buf was written.
	default:
		return sw.dest.Write(buf) // Otherwise do the write to the current destination.
	}
}
