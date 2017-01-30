package ctrace

import "io"

// Encoder is a format-agnostic interface for all log entry marshalers. Since
// log encoders don't need to support the same wide range of use cases as
// general-purpose marshalers, it's possible to make them much faster and
// lower-allocation.
//
// Implementations of the KeyValue interface's methods can, of course, freely
// modify the receiver. However, the Clone and WriteEntry methods will be
// called concurrently and shouldn't modify the receiver.
type Encoder interface {
	// Return the encoder to the appropriate sync.Pool. Unpooled encoder
	// implementations can no-op this method.
	Free()

	WriteStart(io.Writer, spanData) error
	WriteLog(io.Writer, spanData) error
	WriteFinish(io.Writer, spanData) error
}
