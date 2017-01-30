package ctrace

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/opentracing/opentracing-go/log"
)

const (
	// For JSON-escaping; see jsonEncoder.safeAddString below.
	_hex = "0123456789abcdef"
	// Initial buffer size for encoders.
	_initialBufSize = 1024
)

var (
	// errNilSink signals that Encoder.WriteEntry was called with a nil WriteSyncer.
	errNilSink = errors.New("can't write encoded message a nil WriteSyncer")

	jsonPool = sync.Pool{New: func() interface{} {
		return &jsonEncoder{
			// Pre-allocate a reasonably-sized buffer for each encoder.
			prefix: make([]byte, 0, _initialBufSize),
			bytes:  make([]byte, 0, _initialBufSize),
		}
	}}

	// bytesPool = sync.Pool{New: func() interface{} {
	// 	return make([]byte, 0, _initialBufSize)
	// }}
)

// jsonEncoder is an Encoder implementation that writes JSON.
type jsonEncoder struct {
	prefix []byte
	bytes  []byte
}

// NewJSONEncoder creates a fast, low-allocation JSON encoder. By default, JSON
// encoders put the log message under the "msg" key, the timestamp (as
// floating-point seconds since epoch) under the "ts" key, and the log level
// under the "level" key. The encoder appropriately escapes all field keys and
// values.
//
// Note that the encoder doesn't deduplicate keys, so it's possible to produce a
// message like
//   {"foo":"bar","foo":"baz"}
// This is permitted by the JSON specification, but not encouraged. Many
// libraries will ignore duplicate key-value pairs (typically keeping the last
// pair) when unmarshaling, but users should attempt to avoid adding duplicate
// keys.
func NewJSONEncoder() Encoder {
	enc := jsonPool.Get().(*jsonEncoder)
	enc.truncate()
	return enc
}

func (enc *jsonEncoder) EmitString(key, value string) {
	enc.bytes = addKeyString(enc.bytes, key, value)
}

func (enc *jsonEncoder) EmitBool(key string, value bool) {
	enc.bytes = addKeyBool(enc.bytes, key, value)
}

func (enc *jsonEncoder) EmitInt(key string, value int) {
	enc.bytes = addKeyInt(enc.bytes, key, int64(value))
}

func (enc *jsonEncoder) EmitInt32(key string, value int32) {
	enc.bytes = addKeyInt(enc.bytes, key, int64(value))
}

func (enc *jsonEncoder) EmitInt64(key string, value int64) {
	enc.bytes = addKeyInt(enc.bytes, key, value)
}

func (enc *jsonEncoder) EmitUint32(key string, value uint32) {
	enc.bytes = addKeyUint(enc.bytes, key, uint64(value))
}

func (enc *jsonEncoder) EmitUint64(key string, value uint64) {
	enc.bytes = addKeyUint(enc.bytes, key, value)
}

func (enc *jsonEncoder) EmitFloat32(key string, value float32) {
	enc.bytes = addKeyFloat(enc.bytes, key, float64(value))
}

func (enc *jsonEncoder) EmitFloat64(key string, value float64) {
	enc.bytes = addKeyFloat(enc.bytes, key, value)
}

func (enc *jsonEncoder) EmitObject(key string, value interface{}) {
	// Not implemented
}

func (enc *jsonEncoder) EmitLazyLogger(value log.LazyLogger) {
	// Not implemented
}

func (enc *jsonEncoder) Free() {
	jsonPool.Put(enc)
}

func (enc *jsonEncoder) WriteStart(sink io.Writer, d spanData) error {
	if sink == nil {
		return errNilSink
	}

	enc.addPrefix(d)

	enc.bytes = enc.bytes[:0]

	enc.bytes = append(enc.bytes, enc.prefix...)
	enc.bytes = addTags(enc.bytes, d.tags)
	enc.bytes = addKey(enc.bytes, "log")
	enc.bytes = append(enc.bytes, '{')
	enc.bytes = addKeyInt(enc.bytes, "timestamp", d.start.UnixNano()/1e3)
	enc.bytes = addKeyValue(enc.bytes, "event", "Start-Span")
	enc.bytes = append(enc.bytes, '}')
	enc.bytes = append(enc.bytes, '}', '\n')

	expectedBytes := len(enc.bytes)
	n, err := sink.Write(enc.bytes)

	if err != nil {
		return err
	}
	if n != expectedBytes {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, expectedBytes)
	}
	return nil
}

func (enc *jsonEncoder) WriteLog(sink io.Writer, d spanData) error {
	if sink == nil {
		return errNilSink
	}

	enc.bytes = enc.bytes[:0]

	enc.bytes = append(enc.bytes, enc.prefix...)
	enc.bytes = addTags(enc.bytes, d.tags)
	enc.bytes = addKey(enc.bytes, "log")
	enc.bytes = append(enc.bytes, '{')
	enc.bytes = addKeyInt(enc.bytes, "timestamp", d.log.Timestamp.UnixNano()/1e3)
	for _, f := range d.log.Fields {
		f.Marshal(enc)
	}
	enc.bytes = append(enc.bytes, '}')
	enc.bytes = append(enc.bytes, '}', '\n')

	expectedBytes := len(enc.bytes)
	n, err := sink.Write(enc.bytes)

	if err != nil {
		return err
	}
	if n != expectedBytes {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, expectedBytes)
	}
	return nil
}

func (enc *jsonEncoder) WriteFinish(sink io.Writer, data spanData) error {
	return nil
}

func (enc *jsonEncoder) truncate() {
	enc.prefix = enc.prefix[:0]
	enc.bytes = enc.bytes[:0]
}

// func (enc *jsonEncoder) addKey(key string) {
// 	last := len(enc.bytes) - 1
// 	// At some point, we'll also want to support arrays.
// 	if last >= 0 && enc.bytes[last] != '{' {
// 		enc.bytes = append(enc.bytes, ',')
// 	}
// 	enc.bytes = append(enc.bytes, '"')
// 	enc.safeAddString(key)
// 	enc.bytes = append(enc.bytes, '"', ':')
// }

func (enc *jsonEncoder) addPrefix(d spanData) {
	enc.prefix = append(enc.prefix, '{')
	enc.prefix = addKeyID(enc.prefix, "traceId", d.context.TraceID)
	enc.prefix = addKeyID(enc.prefix, "spanId", d.context.SpanID)

	if d.parentID > 0 {
		enc.prefix = addKeyID(enc.prefix, "parentId", d.parentID)
	}

	enc.prefix = addKeyValue(enc.prefix, "operation", d.operation)
	enc.prefix = addKeyInt(enc.prefix, "start", d.start.UnixNano()/1e3)
}

func addTags(bytes []byte, tags map[string]interface{}) []byte {
	if len(tags) <= 0 {
		return bytes
	}
	bytes = addKey(bytes, "tags")
	bytes = append(bytes, '{')

	for k, v := range tags {
		bytes = addKeyValue(bytes, k, v)
	}

	bytes = append(bytes, '}')
	return bytes
}

func addKeyString(bytes []byte, key, val string) []byte {
	bytes = addKey(bytes, key)
	bytes = append(bytes, '"')
	bytes = addSafeString(bytes, val)
	bytes = append(bytes, '"')
	return bytes
}

func addKeyValue(bytes []byte, k string, v interface{}) []byte {
	switch tval := v.(type) {
	case bool:
		bytes = addKeyBool(bytes, k, tval)
	case string:
		bytes = addKeyString(bytes, k, tval)
	case int:
		bytes = addKeyInt(bytes, k, int64(tval))
	case int8:
		bytes = addKeyInt(bytes, k, int64(tval))
	case int16:
		bytes = addKeyInt(bytes, k, int64(tval))
	case int32:
		bytes = addKeyInt(bytes, k, int64(tval))
	case int64:
		bytes = addKeyInt(bytes, k, tval)
	case uint:
		bytes = addKeyUint(bytes, k, uint64(tval))
	case uint8:
		bytes = addKeyUint(bytes, k, uint64(tval))
	case uint16:
		bytes = addKeyUint(bytes, k, uint64(tval))
	case uint32:
		bytes = addKeyUint(bytes, k, uint64(tval))
	case uint64:
		bytes = addKeyUint(bytes, k, tval)
	case float32:
		bytes = addKeyFloat(bytes, k, float64(tval))
	case float64:
		bytes = addKeyFloat(bytes, k, tval)
	default:
		bytes = addKeyString(bytes, k, fmt.Sprint(tval))
	}

	return bytes
}

func addKeyID(bytes []byte, key string, id uint64) []byte {
	bytes = addKey(bytes, key)
	bytes = append(bytes, '"')
	bytes = addSafeString(bytes, fmt.Sprintf("%016x", id))
	bytes = append(bytes, '"')
	return bytes
}

func addKeyInt(bytes []byte, key string, i int64) []byte {
	bytes = addKey(bytes, key)
	bytes = strconv.AppendInt(bytes, i, 10)
	return bytes
}

func addKeyUint(bytes []byte, key string, u uint64) []byte {
	bytes = addKey(bytes, key)
	bytes = strconv.AppendUint(bytes, u, 10)
	return bytes
}

func addKeyFloat(bytes []byte, key string, f float64) []byte {
	bytes = addKey(bytes, key)
	bytes = strconv.AppendFloat(bytes, f, 'f', -1, 64)
	return bytes
}

func addKeyBool(bytes []byte, key string, b bool) []byte {
	bytes = addKey(bytes, key)
	bytes = strconv.AppendBool(bytes, b)
	return bytes
}

func addKey(bytes []byte, key string) []byte {
	last := len(bytes) - 1
	// At some point, we'll also want to support arrays.
	if last >= 0 && bytes[last] != '{' {
		bytes = append(bytes, ',')
	}
	bytes = append(bytes, '"')
	bytes = addSafeString(bytes, key)
	bytes = append(bytes, '"', ':')

	return bytes
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's escaping function, it doesn't attempt to
// protect the user from browser vulnerabilities or JSONP-related problems.
func addSafeString(bytes []byte, s string) []byte {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '\\' && b != '"' {
				bytes = append(bytes, b)
				continue
			}
			switch b {
			case '\\', '"':
				bytes = append(bytes, '\\', b)
			case '\n':
				bytes = append(bytes, '\\', 'n')
			case '\r':
				bytes = append(bytes, '\\', 'r')
			case '\t':
				bytes = append(bytes, '\\', 't')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				bytes = append(bytes, `\u00`...)
				bytes = append(bytes, _hex[b>>4], _hex[b&0xF])
			}
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			bytes = append(bytes, `\ufffd`...)
			i++
			continue
		}
		bytes = append(bytes, s[i:i+size]...)
		i += size
	}
	return bytes
}
