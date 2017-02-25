package ctrace

// SpanContext holds the basic Span metadata.
type spanContext struct {
	// A probabilistically unique identifier for a [multi-span] trace.
	traceID uint64

	// A probabilistically unique identifier for a span.
	spanID uint64

	// The span's associated baggage.
	baggage map[string]string // initialized on first use
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c spanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem returns an entirely new basictracer SpanContext with the
// given key:value baggage pair set.
func (c spanContext) WithBaggageItem(key, val string) spanContext {
	var newBaggage map[string]string
	if c.baggage == nil {
		newBaggage = map[string]string{key: val}
	} else {
		newBaggage = make(map[string]string, len(c.baggage)+1)
		for k, v := range c.baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = val
	}
	// Use positional parameters so the compiler will help catch new fields.
	return spanContext{c.traceID, c.spanID, newBaggage}
}
