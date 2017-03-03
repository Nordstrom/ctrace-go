package http

import (
	"net/http"
	"sync"

	"github.com/Nordstrom/ctrace-go/ext"
	log "github.com/Nordstrom/ctrace-go/log"
	"github.com/felixge/httpsnoop"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
)

type thOptions struct {
	opNameFunc func(r *http.Request) string
}

// THOption controls the behavior of the TracedHandler
type THOption func(*thOptions)

// OperationNameFunc returns a THOption that uses given function f to
// generate operation name for each server-side span.
func OperationNameFunc(f func(r *http.Request) string) THOption {
	return func(options *thOptions) {
		options.opNameFunc = f
	}
}

// OperationName returns a THOption that uses given opName as operation name
// for each server-side span.
func OperationName(opName string) THOption {
	return func(options *thOptions) {
		options.opNameFunc = func(r *http.Request) string {
			return opName
		}
	}
}

// TracedHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHandler(h http.Handler, options ...THOption) http.Handler {
	mux, muxFound := h.(*http.ServeMux)
	opts := thOptions{
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

			status        = http.StatusOK
			body          []byte
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
				Write: func(fn httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(bytes []byte) (int, error) {
						n, err := fn(bytes)
						lock.Lock()
						defer lock.Unlock()

						if body == nil {
							body = bytes
						} else {
							body = append(body, bytes...)
						}
						headerWritten = true
						return n, err
					}
				},
			}
			// ww := NewCapturingResponseWriter(w)
			wr = r.WithContext(ctx)
		)
		h.ServeHTTP(httpsnoop.Wrap(w, hooks), wr)
		span.SetTag(ext.HTTPStatusCodeKey, status)

		if status >= 400 {
			span.SetTag(ext.ErrorKey, true)
			span.LogFields(
				log.Event("error"),
				log.ErrorKind("http-server"),
				olog.String("http.response.body", string(body)),
			)
		}

		span.Finish()
	}

	return http.HandlerFunc(fn)
}

// TracedHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHandlerFunc(fn func(http.ResponseWriter, *http.Request), options ...THOption) http.HandlerFunc {
	return TracedHandler(http.HandlerFunc(fn), options...).ServeHTTP
}
