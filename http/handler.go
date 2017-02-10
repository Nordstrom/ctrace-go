package http

import (
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
	opentracing "github.com/opentracing/opentracing-go"
)

type responseWriter struct {
	bytes []byte
}

func finishSpan(span opentracing.Span, w CapturingResponseWriter, r *http.Request) {
	status := w.StatusCode()
	span.SetTag(ctrace.HTTPStatusCodeKey, status)

	if status >= 400 {
		span.SetTag(ctrace.ErrorKey, true)
		span.SetTag(ctrace.ErrorDetailsKey, string(w.ResponseBody()))
	}

	span.Finish()
}

// TracedHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHandler(
	comp string,
	op string,
	handler http.Handler,
) http.Handler {
	return http.HandlerFunc(TracedHandlerFunc(comp, op, handler.ServeHTTP))
}

// TracedHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHandlerFunc(
	comp string,
	op string,
	fn func(http.ResponseWriter, *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, _ := opentracing.StartSpanFromContext(
			r.Context(),
			op,
			ctrace.SpanKindServer(),
			ctrace.Component(comp),
			ctrace.HTTPRemoteAddr(r.RemoteAddr),
			ctrace.HTTPMethod(r.Method),
			ctrace.HTTPUrl(r.URL.String()),
		)

		cw := NewCapturingResponseWriter(w)
		defer finishSpan(span, cw, r)
		fn(cw, r)
	}
}
