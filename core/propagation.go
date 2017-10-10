package core

import (
	"net/url"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
)

const (
	// HTTPHeaders represents SpanContexts as HTTP header string pairs.
	//
	// Unlike TextMap, the HTTPHeaders format requires that the keys and values
	// be valid as HTTP headers as-is (i.e., character casing may be unstable
	// and special characters are disallowed in keys, values should be
	// URL-escaped, etc).
	//
	// For Tracer.Inject(): the carrier must be a `TextMapWriter`.
	//
	// For Tracer.Extract(): the carrier must be a `TextMapReader`.
	//
	// See HTTPHeaderCarrier for an implementation of both TextMapWriter
	// and TextMapReader that defers to an http.Header instance for storage.
	// For example, Inject():
	//
	//    carrier := ctrace.HTTPHeadersCarrier(httpReq.Header)
	//    err := span.Tracer().Inject(
	//        span, ctrace.HTTPHeaders, carrier)
	//
	// Or Extract():
	//
	//    carrier := ctrace.HTTPHeadersCarrier(httpReq.Header)
	//    span, err := tracer.Extract(
	//        ctrace.HTTPHeaders, carrier)
	//
	HTTPHeaders = opentracing.HTTPHeaders

	// TextMap represents SpanContexts as key:value string pairs.
	//
	// Unlike HTTPHeaders, the TextMap format does not restrict the key or
	// value character sets in any way.
	//
	// For Tracer.Inject(): the carrier must be a `TextMapWriter`.
	//
	// For Tracer.Extract(): the carrier must be a `TextMapReader`.
	TextMap = opentracing.TextMap
)

// Injector is responsible for injecting SpanContext instances in a manner suitable
// for propagation via a format-specific "carrier" object. Typically the
// injection will take place across an RPC boundary, but message queues and
// other IPC mechanisms are also reasonable places to use an Injector.
type Injector interface {
	// Inject takes `SpanContext` and injects it into `carrier`. The actual type
	// of `carrier` depends on the `format` passed to `Tracer.Inject()`.
	//
	// Implementations may return opentracing.ErrInvalidCarrier or any other
	// implementation-specific error if injection fails.
	Inject(ctx opentracing.SpanContext, carrier interface{})
}

// Extractor is responsible for extracting SpanContext instances from a
// format-specific "carrier" object. Typically the extraction will take place
// on the server side of an RPC boundary, but message queues and other IPC
// mechanisms are also reasonable places to use an Extractor.
type Extractor interface {
	// Extract decodes a SpanContext instance from the given `carrier`,
	// or (nil, opentracing.ErrSpanContextNotFound) if no context could
	// be found in the `carrier`.
	Extract(carrier interface{}) (opentracing.SpanContext, error)
}

// TextMapWriter is the Inject() carrier for the TextMap builtin format. With
// it, the caller can encode a SpanContext for propagation as entries in a map
// of unicode strings.
type TextMapWriter opentracing.TextMapWriter

// TextMapCarrier allows the use of regular map[string]string
// as both TextMapWriter and TextMapReader.
type TextMapCarrier opentracing.TextMapCarrier

// HTTPHeadersCarrier satisfies both TextMapWriter and TextMapReader.
//
// Example usage for server side:
//
//     carrier := opentracing.HttpHeadersCarrier(httpReq.Header)
//     spanContext, err := tracer.Extract(opentracing.HttpHeaders, carrier)
//
// Example usage for client side:
//
//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
//     err := tracer.Inject(
//         span.Context(),
//         opentracing.HttpHeaders,
//         carrier)
//
type HTTPHeadersCarrier opentracing.HTTPHeadersCarrier

// Set implements Set() of ctrace.TextMapWriter
func (c TextMapCarrier) Set(key, val string) {
	c[key] = val
}

// Set implements Set() of ctrace.TextMapWriter
func (c HTTPHeadersCarrier) Set(key, val string) {
	c[key] = []string{val}
}

type textMapPropagator struct {
	traceIDKey    string
	spanIDKey     string
	baggagePrefix string
	encodeKey     func(string) string
	decodeKey     func(string) string
	encodeValue   func(string) string
	decodeValue   func(string) string
}

func newTextMapPropagator() *textMapPropagator {
	var passthrough = func(s string) string {
		return s
	}

	return &textMapPropagator{
		traceIDKey:    "ct-trace-id",
		spanIDKey:     "ct-span-id",
		baggagePrefix: "ct-bag-",
		encodeKey:     passthrough,
		decodeKey:     passthrough,
		encodeValue:   passthrough,
		decodeValue:   passthrough,
	}
}

func newHTTPHeadersPropagator() *textMapPropagator {
	return &textMapPropagator{
		traceIDKey:    "ct-trace-id",
		spanIDKey:     "ct-span-id",
		baggagePrefix: "ct-bag-",
		encodeKey: func(key string) string {
			return url.QueryEscape(key)
		},
		decodeKey: func(key string) string {
			// ignore decoding errors, cannot do anything about them
			if k, err := url.QueryUnescape(key); err == nil {
				return strings.ToLower(k)
			}
			return strings.ToLower(key)
		},
		encodeValue: func(val string) string {
			return url.QueryEscape(val)
		},
		decodeValue: func(val string) string {
			// ignore decoding errors, cannot do anything about them
			if v, err := url.QueryUnescape(val); err == nil {
				return v
			}
			return val
		},
	}
}

func (p *textMapPropagator) Inject(
	ctx opentracing.SpanContext,
	opaqueCarrier interface{},
) error {
	sc, ok := ctx.(spanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	carrier, ok := opaqueCarrier.(TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}
	// TODO: At this point we don't need to encode trace and span id values
	// this may change
	carrier.Set(p.traceIDKey, sc.traceID)
	carrier.Set(p.spanIDKey, sc.spanID)

	for k, v := range sc.baggage {
		carrier.Set(p.baggagePrefix+p.encodeKey(k), p.encodeValue(v))
	}
	return nil
}

func (p *textMapPropagator) Extract(
	opaqueCarrier interface{},
) (opentracing.SpanContext, error) {
	carrier, ok := opaqueCarrier.(opentracing.TextMapReader)
	if !ok {
		return nil, opentracing.ErrInvalidCarrier
	}
	requiredFieldCount := 0
	var traceID, spanID string
	var err error

	decodedBaggage := make(map[string]string)
	err = carrier.ForeachKey(func(k, v string) error {
		k = p.decodeKey(k)
		switch k {
		case p.traceIDKey:
			traceID = v
		case p.spanIDKey:
			spanID = v
		default:
			if strings.HasPrefix(k, p.baggagePrefix) {
				key := strings.TrimPrefix(k, p.baggagePrefix)
				decodedBaggage[key] = v
			}
			// Balance off the requiredFieldCount++ just below...
			requiredFieldCount--
		}
		requiredFieldCount++
		return nil
	})
	if err != nil {
		return nil, err
	}
	if requiredFieldCount < 2 {
		if requiredFieldCount == 0 {
			return nil, opentracing.ErrSpanContextNotFound
		}
		return nil, opentracing.ErrSpanContextCorrupted
	}

	return spanContext{
		traceID: traceID,
		spanID:  spanID,
		baggage: decodedBaggage,
	}, nil
}
