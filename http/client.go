package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Nordstrom/ctrace-go/ext"
	"github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
)

type tracedTransport struct {
	component string
	transport http.RoundTripper
	options   httpOptions
}

// NewTracedTransport creates a new Transporter (http.RoundTripper) that intercepts
// and traces egress requests.
func NewTracedTransport(t http.RoundTripper, options ...Option) http.RoundTripper {
	opts := httpOptions{
		opNameFunc: func(r *http.Request) string {
			return r.Method + ":" + r.URL.Path
		},
	}

	for _, opt := range options {
		opt(&opts)
	}
	return &tracedTransport{
		component: "ctrace.TracedTransport",
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

func (t *tracedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println("RoundTrip")
	fmt.Println(req.Context())
	span, _ := opentracing.StartSpanFromContext(
		req.Context(),
		t.options.opNameFunc(req),
		ext.SpanKindClient(),
		ext.Component(t.component),
		ext.HTTPMethod(req.Method),
		ext.HTTPUrl(req.URL.String()),
	)

	tracer := opentracing.GlobalTracer()
	tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))

	res, err := t.transport.RoundTrip(req)

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
	if req.Method == "HEAD" {
		span.Finish()
	} else {
		res.Body = closeTracker{res.Body, span}
	}
	return res, nil
}
