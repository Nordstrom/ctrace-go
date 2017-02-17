package http

import (
	"fmt"
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
	opentracing "github.com/opentracing/opentracing-go"
)

type responseWriter struct {
	bytes []byte
}

func finishSpan(span opentracing.Span, w CapturingResponseWriter, r *http.Request) {
	status := w.StatusCode()
	fmt.Printf("status=%d\n", status)
	span.SetTag(ctrace.HTTPStatusCodeKey, status)

	if status >= 400 {
		span.SetTag(ctrace.ErrorKey, true)
		span.SetTag(ctrace.ErrorDetailsKey, string(w.ResponseBody()))
	}

	span.Finish()
}

// TracedHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(TracedHandlerFunc(handler.ServeHTTP))
}

// TracedHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHandlerFunc(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		parentCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))

		span, ctx := opentracing.StartSpanFromContext(
			r.Context(),
			r.Method+":"+r.URL.Path,
			opentracing.ChildOf(parentCtx),
			ctrace.SpanKindServer(),
			ctrace.Component("http-handler"),
			ctrace.HTTPRemoteAddr(r.RemoteAddr),
			ctrace.HTTPMethod(r.Method),
			ctrace.HTTPUrl(r.URL.String()),
		)

		ww := NewCapturingResponseWriter(w)
		wr := r.WithContext(ctx)
		defer finishSpan(span, ww, wr)
		fn(ww, wr)
	}
}
