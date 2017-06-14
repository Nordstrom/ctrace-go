package http

import (
	"net/http"
	"os"
	"sync"

	"github.com/Nordstrom/ctrace-go/ext"
	"github.com/felixge/httpsnoop"
	opentracing "github.com/opentracing/opentracing-go"
)

// TracedHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHandler(h http.Handler, options ...Option) http.Handler {
	mux, muxFound := h.(*http.ServeMux)
	opts := httpOptions{
		opNameFunc: func(r *http.Request) string {
			if muxFound {
				_, pattern := mux.Handler(r)
				return r.Method + ":" + pattern
			}
			return r.Method + ":" + r.URL.Path
		},
	}

	for _, opt := range options {
		opt(&opts)
	}
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			tracer       = opentracing.GlobalTracer()
			parentCtx, _ = tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header))

			span, ctx = opentracing.StartSpanFromContext(
				r.Context(),
				opts.opNameFunc(r),
				opentracing.ChildOf(parentCtx),
				ext.SpanKindServer(),
				ext.Component("ctrace.TracedHandler"),
				ext.HTTPRemoteAddr(r.RemoteAddr),
				ext.HTTPMethod(r.Method),
				ext.HTTPUrl(r.URL.String()),
				ext.HTTPUserAgent(r.UserAgent()),
			)

			status = http.StatusOK
			// body          []byte
			headerWritten = false
			lock          sync.Mutex
			hooks         = httpsnoop.Hooks{
				WriteHeader: func(fn httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						fn(code)
						lock.Lock()
						defer lock.Unlock()
						if !headerWritten {
							status = code
							headerWritten = true
						}
					}
				},
				// TODO: Disable Tracing Response Body for Now.  Needs more research.
				// Write: func(fn httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				// 	return func(bytes []byte) (int, error) {
				// 		n, err := fn(bytes)
				// 		lock.Lock()
				// 		defer lock.Unlock()
				//
				// 		if body == nil {
				// 			body = bytes
				// 		} else {
				// 			body = append(body, bytes...)
				// 		}
				// 		headerWritten = true
				// 		return n, err
				// 	}
				// },
			}
			// ww := NewCapturingResponseWriter(w)
			wr = r.WithContext(ctx)
		)
		h.ServeHTTP(httpsnoop.Wrap(w, hooks), wr)
		span.SetTag(ext.HTTPStatusCodeKey, status)

		serviceName := os.Getenv("CTRACE_SERVICE_NAME")
		span.SetTag("service", serviceName)

		if status >= 400 {
			span.SetTag(ext.ErrorKey, true)
		}

		span.Finish()
	}

	return http.HandlerFunc(fn)
}

// TracedHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHandlerFunc(fn func(http.ResponseWriter, *http.Request), options ...Option) http.HandlerFunc {
	return TracedHandler(http.HandlerFunc(fn), options...).ServeHTTP
}
