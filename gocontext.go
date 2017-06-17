package ctrace

import (
	"context"

	clog "github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// ChildOf returns a StartSpanOption pointing to a dependent parent span.
// If sc == nil, the option has no effect.
//
// See ChildOfRef, SpanReference
var ChildOf = opentracing.ChildOf

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

// LogInfo allows the logging of an Info Event based on the
// current context.Context.  If a running span does not exist on the current
// context, nothing is logged.
func LogInfo(ctx context.Context, event string, fields ...log.Field) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	f := []log.Field{
		clog.Event(event),
	}
	f = append(f, fields...)
	span.LogFields(f...)
}

// LogErrorMessage allows the logging of an Error with a Message based on the
// current context.Context.  If a running span does not exist on the current
// context, nothing is logged.
func LogErrorMessage(ctx context.Context, message string, fields ...log.Field) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	f := []log.Field{
		clog.Event("error"),
		clog.ErrorKind("message"),
		clog.Message(message),
	}
	f = append(f, fields...)
	span.LogFields(f...)
}

// LogErrorObject allows the logging of an Error Object based on the
// current context.Context.  If a running span does not exist on the current
// context, nothing is logged.
func LogErrorObject(ctx context.Context, e error, fields ...log.Field) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	f := []log.Field{
		clog.Event("error"),
		clog.ErrorKind("object"),
		clog.ErrorObject(e),
	}
	f = append(f, fields...)
	span.LogFields(f...)
}
