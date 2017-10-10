package core

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// jsonEncoder is a fast / lite json encoder with just enough functionality to
// support the span and log encoders.  It is the default (and at present only)
// encoding supported by ctrace.
type jsonEncoder struct{}

func (enc *jsonEncoder) encodeKeyString(bytes []byte, key, val string) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = append(bytes, '"')
	bytes = enc.encodeString(bytes, val)
	bytes = append(bytes, '"')
	return bytes
}

func (enc *jsonEncoder) encodeKeyValue(bytes []byte, k string, v interface{}) []byte {
	switch tval := v.(type) {
	case bool:
		bytes = enc.encodeKeyBool(bytes, k, tval)
	case string:
		bytes = enc.encodeKeyString(bytes, k, tval)
	case int:
		bytes = enc.encodeKeyInt(bytes, k, int64(tval))
	case int8:
		bytes = enc.encodeKeyInt(bytes, k, int64(tval))
	case int16:
		bytes = enc.encodeKeyInt(bytes, k, int64(tval))
	case int32:
		bytes = enc.encodeKeyInt(bytes, k, int64(tval))
	case int64:
		bytes = enc.encodeKeyInt(bytes, k, tval)
	case uint:
		bytes = enc.encodeKeyUint(bytes, k, uint64(tval))
	case uint8:
		bytes = enc.encodeKeyUint(bytes, k, uint64(tval))
	case uint16:
		bytes = enc.encodeKeyUint(bytes, k, uint64(tval))
	case uint32:
		bytes = enc.encodeKeyUint(bytes, k, uint64(tval))
	case uint64:
		bytes = enc.encodeKeyUint(bytes, k, tval)
	case float32:
		bytes = enc.encodeKeyFloat(bytes, k, float64(tval))
	case float64:
		bytes = enc.encodeKeyFloat(bytes, k, tval)
	default:
		bytes = enc.encodeKeyString(bytes, k, fmt.Sprint(tval))
	}

	return bytes
}

func (enc *jsonEncoder) encodeKeyID(bytes []byte, key string, id string) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = append(bytes, '"')
	bytes = enc.encodeString(bytes, id)
	bytes = append(bytes, '"')
	return bytes
}

func (enc *jsonEncoder) encodeKeyInt(bytes []byte, key string, i int64) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = strconv.AppendInt(bytes, i, 10)
	return bytes
}

func (enc *jsonEncoder) encodeKeyUint(bytes []byte, key string, u uint64) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = strconv.AppendUint(bytes, u, 10)
	return bytes
}

func (enc *jsonEncoder) encodeKeyFloat(bytes []byte, key string, f float64) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = strconv.AppendFloat(bytes, f, 'f', -1, 64)
	return bytes
}

func (enc *jsonEncoder) encodeKeyBool(bytes []byte, key string, b bool) []byte {
	bytes = enc.encodeKey(bytes, key)
	bytes = strconv.AppendBool(bytes, b)
	return bytes
}

func (enc *jsonEncoder) encodeKey(bytes []byte, key string) []byte {
	last := len(bytes) - 1
	// At some point, we'll also want to support arrays.
	if last >= 0 && bytes[last] != '{' {
		bytes = append(bytes, ',')
	}
	bytes = append(bytes, '"')
	bytes = enc.encodeString(bytes, key)
	bytes = append(bytes, '"', ':')

	return bytes
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's escaping function, it doesn't attempt to
// protect the user from browser vulnerabilities or JSONP-related problems.
func (enc *jsonEncoder) encodeString(bytes []byte, s string) []byte {
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
