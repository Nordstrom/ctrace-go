package ctrace

import (
	"encoding/json"

	"github.com/Nordstrom/ctrace-go/ext"
  "github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
  "github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	opentracing "github.com/opentracing/opentracing-go"
)

type LambdaFunction func(evt *apigatewayproxyevt.Event, ctx *runtime.Context) (interface{}, error)

type LambdaInterceptorFunction func(evt *apigatewayproxyevt.Event, ctx *runtime.Context) SpanConfig

type tracedLambdaOptions struct {
	interceptor LambdaInterceptorFunction
}

// Option controls the behavior of the ctrace http middleware
type TracedLambdaOption func(*tracedLambdaOptions)

// LambdaInterceptor returns a Option that uses given function f to
// generate operation name for each span.
func LambdaInterceptor(f LambdaInterceptorFunction) TracedLambdaOption {
	return func(options *tracedLambdaOptions) {
		options.interceptor = f
	}
}

func TracedApiGwLambdaProxy(fn LambdaFunction, options ...TracedLambdaOption) {
	opts := tracedLambdaOptions{}

	for _, opt := range options {
		opt(&opts)
	}
	fn := func(evt *apigatewayproxyevt.Event, ctx *runtime.Context) (interface{}, error) {
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

		if status >= 400 {
			span.SetTag(ext.ErrorKey, true)
		}

		span.Finish()
	}

	return http.HandlerFunc(fn)
}
