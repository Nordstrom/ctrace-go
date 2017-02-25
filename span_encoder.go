package ctrace

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

const (
	// For JSON-escaping; see spanEncoder.safeAddString below.
	_hex = "0123456789abcdef"
)

// SpanEncoder is a format-agnostic interface for encoding span events.
type SpanEncoder interface {
	//Close()

	// Return the encoder to the appropriate sync.Pool. Unpooled encoder
	// implementations can no-op this method.
	Encode(opentracing.Span) []byte
}

// spanEncoder is an Encoder implementation that writes JSON.
type spanEncoder struct {
	jsonEncoder
}

// NewSpanEncoder creates a fast, low-allocation JSON encoder.
func NewSpanEncoder() SpanEncoder {
	return &spanEncoder{jsonEncoder: jsonEncoder{}}
}

func (enc *spanEncoder) Encode(osp opentracing.Span) []byte {
	sp := osp.(*span)
	bytes := make([]byte, 0, 1024)

	if sp.prefix == nil || len(sp.prefix) <= 0 {
		enc.encodePrefix(sp)
	}

	bytes = append(bytes, sp.prefix...)
	if sp.duration >= 0 {
		bytes = enc.encodeKeyInt(bytes, "duration", sp.duration.Nanoseconds()/1e3)
	}
	bytes = enc.encodeTags(bytes, sp.tags)
	bytes = enc.encodeBaggage(bytes, sp.context.baggage)
	bytes = enc.encodeLog(bytes, sp.log)
	bytes = append(bytes, '}', '\n')

	return bytes
}

func (enc *spanEncoder) encodePrefix(sp *span) {
	if sp.prefix == nil {
		sp.prefix = make([]byte, 0, 512)
	}
	sp.prefix = append(sp.prefix, '{')
	sp.prefix = enc.encodeKeyID(sp.prefix, "traceId", sp.context.traceID)
	sp.prefix = enc.encodeKeyID(sp.prefix, "spanId", sp.context.spanID)

	if sp.parentID > 0 {
		sp.prefix = enc.encodeKeyID(sp.prefix, "parentId", sp.parentID)
	}

	sp.prefix = enc.encodeKeyValue(sp.prefix, "operation", sp.operation)
	if sp.start.IsZero() {
		sp.start = time.Now()
	}
	sp.prefix = enc.encodeKeyInt(sp.prefix, "start", sp.start.UnixNano()/1e3)
}

func (enc *spanEncoder) encodeTags(bytes []byte, tags map[string]interface{}) []byte {
	if len(tags) <= 0 {
		return bytes
	}
	// sort.Strings(tags)
	bytes = enc.encodeKey(bytes, "tags")
	bytes = append(bytes, '{')

	for k, v := range tags {
		bytes = enc.encodeKeyValue(bytes, k, v)
	}

	bytes = append(bytes, '}')
	return bytes
}

func (enc *spanEncoder) encodeBaggage(
	bytes []byte,
	baggage map[string]string,
) []byte {
	if len(baggage) <= 0 {
		return bytes
	}
	// sort.Strings(tags)
	bytes = enc.encodeKey(bytes, "baggage")
	bytes = append(bytes, '{')

	for k, v := range baggage {
		bytes = enc.encodeKeyString(bytes, k, v)
	}

	bytes = append(bytes, '}')
	return bytes
}

func (enc *spanEncoder) encodeLog(bytes []byte, log opentracing.LogRecord) []byte {
	if log.Timestamp.IsZero() {
		return bytes
	}
	bytes = enc.encodeKey(bytes, "log")
	bytes = append(bytes, '{')
	bytes = enc.encodeKeyInt(bytes, "timestamp", log.Timestamp.UnixNano()/1e3)
	for _, f := range log.Fields {
		bytes = enc.encodeKeyValue(bytes, f.Key(), f.Value())
	}
	bytes = append(bytes, '}')
	return bytes
}
