package ctrace

import (
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
)

func injectText(
	ctx opentracing.SpanContext,
	opaqueCarrier interface{},
) error {
	sc, ok := ctx.(spanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	carrier, ok := opaqueCarrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}
	carrier.Set("X-Correlation-Id", strconv.FormatUint(sc.traceID, 16))
	carrier.Set("X-Request-Id", strconv.FormatUint(sc.spanID, 16))

	for k, v := range sc.baggage {
		carrier.Set("X-Baggage-"+k, v)
	}
	return nil
}

func extractText(
	opaqueCarrier interface{},
) (opentracing.SpanContext, error) {
	carrier, ok := opaqueCarrier.(opentracing.TextMapReader)
	if !ok {
		return nil, opentracing.ErrInvalidCarrier
	}
	requiredFieldCount := 0
	var traceID, spanID uint64
	var err error
	decodedBaggage := make(map[string]string)
	err = carrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case "x-correlation-id":
			traceID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return opentracing.ErrSpanContextCorrupted
			}
		case "x-request-id":
			spanID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return opentracing.ErrSpanContextCorrupted
			}
		default:
			lowercaseK := strings.ToLower(k)
			if strings.HasPrefix(lowercaseK, "x-baggage-") {
				decodedBaggage[strings.TrimPrefix(lowercaseK, "x-baggage-")] = v
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
