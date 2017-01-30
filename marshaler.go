package ctrace

// LogMarshaler allows user-defined types to efficiently add themselves to the
// logging context, and to selectively omit information which shouldn't be
// included in logs (e.g., passwords).
type LogMarshaler interface {
	MarshalLog(KeyValue) error
}

// LogMarshalerFunc is a type adapter that allows using a function as a
// LogMarshaler.
type LogMarshalerFunc func(KeyValue) error

// MarshalLog calls the underlying function.
func (f LogMarshalerFunc) MarshalLog(kv KeyValue) error {
	return f(kv)
}
