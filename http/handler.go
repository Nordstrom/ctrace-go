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

func finishSpan(span opentracing.Span, r *http.Request) {
	fmt.Print(r)
	ctrace.HTTPStatusCode(r.Response.StatusCode).Set(span)

	if r.Response.StatusCode >= 400 {
		ctrace.Error(true).Set(span)

		var b []byte
		n, err := r.Response.Body.Read(b)
		if err != nil && n > 0 {
			ctrace.ErrorDetails(string(b)).Set(span)
		}
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

		defer finishSpan(span, r)
		fn(w, r)
	}
}
