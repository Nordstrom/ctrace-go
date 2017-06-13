package ctrace

import (
	"io"
	"net/http"
	"sync"

	"github.com/Nordstrom/ctrace-go/ext"
	"github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
)

type SpanConfig struct {
	OperationName string
	Tags          opentracing.Tags
}

type httpOptions struct {
	interceptor func(r *http.Request) SpanConfig
}

// Option controls the behavior of the ctrace http middleware
type HttpOption func(*httpOptions)

// OperationNameFunc returns a Option that uses given function f to
// generate operation name for each span.
func HttpInterceptor(f func(r *http.Request) SpanConfig) HttpOption {
	return func(options *httpOptions) {
		options.interceptor = f
	}
}

type tracedHttpClientTransport struct {
	component string
	transport http.RoundTripper
	options   httpOptions
}

func loadHttpOptions(options ...HttpOption) httpOptions {
	opts := httpOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// TracedHttpClientTransport creates a new Transporter (http.RoundTripper) that intercepts
// and traces egress requests.
func TracedHttpClientTransport(t http.RoundTripper, options ...HttpOption) http.RoundTripper {
	opts := loadHttpOptions(options...)
	return &tracedHttpClientTransport{
		component: "ctrace.tracedHttpClientTransport",
		transport: t,
		options:   opts,
	}
}

type closeTracker struct {
	io.ReadCloser
	sp opentracing.Span
}

func (c closeTracker) Close() error {
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

func optioniallyIntercept(opts httpOptions, r *http.Request) SpanConfig {
	var config SpanConfig

	if opts.interceptor != nil {
		config = opts.interceptor(r)
	}
	return config
}

func (t *tracedHttpClientTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	config := optioniallyIntercept(t.options, r)
	op := r.Method + ":" + r.URL.Path
	tags := map[string]interface{}{}
	tags[ext.SpanKindKey] = ext.SpanKindClientValue
	tags[ext.ComponentKey] = t.component
	tags[ext.HTTPMethodKey] = r.Method
	tags[ext.HTTPUrlKey] = r.URL.String()

	if config.OperationName != "" {
		op = config.OperationName
	}
	for k, v := range config.Tags {
		tags[k] = v
	}
	opts := opentracing.StartSpanOptions{Tags: tags}
	span, _ := startSpanWithOptionsFromContext(r.Context(), op, opts)

	tracer := opentracing.GlobalTracer()
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

func httpOperationName(mux *http.ServeMux, muxFound bool, r *http.Request) string {
	if muxFound {
		_, pattern := mux.Handler(r)
		return r.Method + ":" + pattern
	}
	return r.Method + ":" + r.URL.Path
}

type responseInterceptor struct {
	tracer        Tracer
	parentCtx     SpanContext
	ctx           SpanContext
	writer        http.ResponseWriter
	headerWritten bool
	traced        bool
	status        int
	lock          sync.Mutex
}

func (i *responseInterceptor) trace(code int) {
	i.lock.Lock()
	defer i.lock.Unlock()
	if !i.headerWritten {
		i.status = code
		i.tracer.Inject(i.ctx, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(i.writer.Header()))
		i.headerWritten = true
	}
}

func (i *responseInterceptor) WriteHeader(code int) {
	i.writer.WriteHeader(code)
	i.trace(code)
}

func (i *responseInterceptor) Write(b []byte) (int, error) {
	if !i.headerWritten {
		r, e := i.writer.Write(b)
		i.trace(http.StatusOK)
		return r, e
	}
	return 0, nil
}

func (i *responseInterceptor) Header() http.Header {
	return i.writer.Header()
}

// TracedHandler returns a http.Handler that is traced as an opentracing.Span
func TracedHttpHandler(h http.Handler, options ...HttpOption) http.Handler {
	mux, muxFound := h.(*http.ServeMux)
	opts := loadHttpOptions(options...)

	fn := func(w http.ResponseWriter, r *http.Request) {
		tracer := Global()
		parentCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))

		config := optioniallyIntercept(opts, r)

		var op string
		if config.OperationName != "" {
			op = config.OperationName
		} else {
			op = httpOperationName(mux, muxFound, r)
		}
		tags := map[string]interface{}{}
		tags[ext.SpanKindKey] = ext.SpanKindServerValue
		tags[ext.ComponentKey] = "ctrace.TracedHttpHandler"
		tags[ext.HTTPRemoteAddrKey] = r.RemoteAddr
		tags[ext.HTTPMethodKey] = r.Method
		tags[ext.HTTPUrlKey] = r.URL.String()
		tags[ext.HTTPUserAgentKey] = r.UserAgent()

		if config.OperationName != "" {
			op = config.OperationName
		}
		for k, v := range config.Tags {
			tags[k] = v
		}
		so := opentracing.StartSpanOptions{
			References: []opentracing.SpanReference{ChildOf(parentCtx)},
			Tags:       tags,
		}
		span, ctx := startSpanWithOptionsFromContext(r.Context(), op, so)

		ri := responseInterceptor{
			tracer:    tracer,
			parentCtx: parentCtx.(SpanContext),
			ctx:       ctx.(SpanContext),
			writer:    w,
		}

		h.ServeHTTP(ri, r.WithContext(ctx))
		span.SetTag(ext.HTTPStatusCodeKey, status)

		if status >= 400 {
			span.SetTag(ext.ErrorKey, true)
		}

		span.Finish()
	}

	return http.HandlerFunc(fn)
}

// TracedHandlerFunc returns a http.HandlerFunc that is traced as an opentracing.Span
func TracedHttpHandlerFunc(fn func(http.ResponseWriter, *http.Request), options ...HttpOption) http.HandlerFunc {
	return TracedHttpHandler(http.HandlerFunc(fn), options...).ServeHTTP
}
