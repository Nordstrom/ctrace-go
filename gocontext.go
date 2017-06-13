package ctrace

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
)

// ContextWithSpan returns a new `context.Context` that holds a reference to
// `span`'s SpanContext.
func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

// SpanFromContext returns the `Span` previously associated with `ctx`, or
// `nil` if no such `Span` could be found.
//
// NOTE: context.Context != SpanContext: the former is Go's intra-process
// context propagation mechanism, and the latter houses OpenTracing's per-Span
// identity and baggage information.
func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// StartSpanFromContext starts and returns a Span with `operationName`, using
// any Span found within `ctx` as a ChildOfRef. If no such parent could be
// found, StartSpanFromContext creates a root (parentless) Span.
//
// The second return value is a context.Context object built around the
// returned Span.
//
// Example usage:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp, ctx := opentracing.StartSpanFromContext(ctx, "SomeFunction")
//        defer sp.Finish()
//        ...
//    }
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, operationName, opts...)
}

func startSpanWithOptionsFromContext(ctx context.Context, operationName string, opts opentracing.StartSpanOptions) (opentracing.Span, context.Context) {
	var span opentracing.Span
	tracer := Global()
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		ref := ChildOf(parentSpan.Context())
		opts.References = append(opts.References, ref)
		span = tracer.StartSpanWithOptions(operationName, opts)
	} else {
		span = tracer.StartSpanWithOptions(operationName, opts)
	}
	return span, ContextWithSpan(ctx, span)
}
