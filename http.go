package ctrace

import (
	"io"
	"net/http"
	"sync"

	"github.com/Nordstrom/ctrace-go/core"
	"github.com/Nordstrom/ctrace-go/ext"
	"github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
)

// *******************  Http Client  *******************

// TracedHTTPInterceptor is a defined function called during HTTP tracing
// to return custom OperationName and/or Tags for the given request
type TracedHTTPInterceptor func(r *http.Request) SpanConfig

type tracedHTTPClientTransport struct {
	component   string
	transport   http.RoundTripper
	interceptor []TracedHTTPInterceptor // There is 0 or 1 interceptor
}

// TracedHTTPClientTransport creates a new Transporter (http.RoundTripper) that intercepts
// and traces egress requests.
func TracedHTTPClientTransport(
	t http.RoundTripper,
	interceptor ...TracedHTTPInterceptor,
) http.RoundTripper {
	return &tracedHTTPClientTransport{
		component:   "ctrace.TracedHttpClientTransport",
		transport:   t,
		interceptor: interceptor,
	}
}

type closeTracker struct {
	io.ReadCloser
	sp opentracing.Span
}

func (c closeTracker) Close() error {
	debug("Closing Response Writer...")
	err := c.ReadCloser.Close()
	if err != nil {
		c.sp.SetTag(ext.ErrorKey, true)
		c.sp.LogFields(
			log.Event("error"),
			log.ErrorKind("http-client"),
			log.ErrorObject(err),
		)
	}
	c.sp.Finish()
	return err
}

func optioniallyInterceptHTTP(
	i []TracedHTTPInterceptor,
	r *http.Request,
) SpanConfig {
	for _, f := range i {
		return f(r)
	}
	return SpanConfig{}
}

func (t *tracedHTTPClientTransport) RoundTrip(
	r *http.Request,
) (*http.Response, error) {
	op := r.Method + ":" + r.URL.Path
	debug("Starting client RoundTrip: op=%s", op)

	config := optioniallyInterceptHTTP(t.interceptor, r)
	opts := []opentracing.StartSpanOption{
		ext.SpanKindClient(),
		ext.Component(t.component),
		ext.HTTPMethod(r.Method),
		ext.HTTPUrl(r.URL.String()),
	}

	if config.OperationName != "" {
		op = config.OperationName
	}
	if len(config.Tags) > 0 {
		opts = append(opts, config.Tags...)
	}

	tracer := Global()
	span, _ := StartSpanFromContext(r.Context(), op, opts...)
	tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	res, err := t.transport.RoundTrip(r)

	if err != nil {
		span.SetTag(ext.ErrorKey, true)
		span.LogFields(
			log.Event("error"),
			log.ErrorKind("http-client"),
			log.ErrorObject(err),
		)
		span.Finish()
		return res, err
	}

	span.SetTag(ext.HTTPStatusCodeKey, res.StatusCode)
	if res.StatusCode >= 400 {
		span.SetTag(ext.ErrorKey, true)
	}
	if r.Method == "HEAD" {
		span.Finish()
	} else {
		res.Body = closeTracker{res.Body, span}
	}
	return res, nil
}

// *******************  Http Server  *******************

func httpOperationName(
	mux *http.ServeMux,
	muxFound bool,
	r *http.Request,
) string {
	if muxFound {
		_, pattern := mux.Handler(r)
		return r.Method + ":" + pattern
	}
	return r.Method + ":" + r.URL.Path
}

type responseInterceptor struct {
	tracer        opentracing.Tracer
	span          opentracing.Span
	parentCtx     core.SpanContext
	ctx           core.SpanContext
	writer        http.ResponseWriter
	headerWritten bool
	traced        bool
	lock          sync.Mutex
}

func (i *responseInterceptor) trace(code int) {
	i.lock.Lock()
	defer i.lock.Unlock()
	if !i.headerWritten {
		i.span.SetTag(ext.HTTPStatusCodeKey, code)

		if code >= 400 {
			i.span.SetTag(ext.ErrorKey, true)
		}

		i.span.Finish()
		i.tracer.Inject(
			i.ctx,
			core.HTTPHeaders,
			core.HTTPHeadersCarrier(i.writer.Header()),
		)
		i.headerWritten = true
	}
}

func (i *responseInterceptor) WriteHeader(code int) {
	i.writer.WriteHeader(code)
	i.trace(code)
}

func (i *responseInterceptor) Write(b []byte) (int, error) {
	debug("Writing response body %s", string(b))
	r, e := i.writer.Write(b)

	if !i.headerWritten {
		i.trace(http.StatusOK)
	}
	return r, e
}

func (i *responseInterceptor) Header() http.Header {
	return i.writer.Header()
}

// TracedHTTPHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHTTPHandler(
	h http.Handler,
	interceptor ...TracedHTTPInterceptor,
) http.Handler {
	mux, muxFound := h.(*http.ServeMux)

	fn := func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		parentCtx, _ := tracer.Extract(core.HTTPHeaders, core.HTTPHeadersCarrier(r.Header))

		config := optioniallyInterceptHTTP(interceptor, r)

		var op string
		if config.OperationName != "" {
			op = config.OperationName
		} else {
			op = httpOperationName(mux, muxFound, r)
		}
		opts := []opentracing.StartSpanOption{
			ChildOf(parentCtx),
			ext.SpanKindServer(),
			ext.Component("ctrace.TracedHttpHandler"),
			ext.HTTPRemoteAddr(r.RemoteAddr),
			ext.HTTPMethod(r.Method),
			ext.HTTPUrl(r.URL.String()),
			ext.HTTPUserAgent(r.UserAgent()),
		}

		if config.OperationName != "" {
			op = config.OperationName
		}
		if len(config.Tags) > 0 {
			opts = append(opts, config.Tags...)
		}
		debug("TracedHttpHandler: StartSpan(%s)", op)

		span := tracer.StartSpan(op, opts...)
		ctx := span.Context()

		ri := responseInterceptor{
			tracer: tracer,
			span:   span,
			ctx:    ctx.(core.SpanContext),
			writer: w,
		}

		if parentCtx != nil {
			ri.parentCtx = parentCtx.(core.SpanContext)
		}

		debug("TracedHttpHandler: ServeHTTP(...)")
		h.ServeHTTP(&ri, r.WithContext(ContextWithSpan(r.Context(), span)))
	}

	return http.HandlerFunc(fn)
}

// TracedHTTPHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHTTPHandlerFunc(
	fn func(http.ResponseWriter, *http.Request),
	interceptor ...TracedHTTPInterceptor,
) http.HandlerFunc {
	return TracedHTTPHandler(http.HandlerFunc(fn), interceptor...).ServeHTTP
}
