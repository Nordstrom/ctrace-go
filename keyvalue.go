package ctrace

// KeyValue is an encoding-agnostic interface to add structured data to the
// logging context. Like maps, KeyValues aren't safe for concurrent use (though
// typical use shouldn't require locks).
//
// See Marshaler for an example.
type KeyValue interface {
	AddBool(key string, value bool)
	AddFloat64(key string, value float64)
	AddInt(key string, value int)
	AddInt64(key string, value int64)
	AddUint(key string, value uint)
	AddUint64(key string, value uint64)
	AddUintptr(key string, value uintptr)
	// AddMarshaler(key string, marshaler LogMarshaler) error
	// AddObject uses reflection to serialize arbitrary objects, so it's slow and
	// allocation-heavy. Consider implementing the LogMarshaler interface instead.
	AddObject(key string, value interface{}) error
	AddString(key, value string)
}
