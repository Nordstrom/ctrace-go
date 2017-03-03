package http

import (
	"net/http"

	"github.com/Nordstrom/ctrace-go/ext"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type interceptor struct {
	component   string
	transporter http.RoundTripper
}

// NewTransporter creates a new Transporter (http.RoundTripper) that intercepts
// and traces egress requests.
func NewTransporter(component string, t http.RoundTripper) http.RoundTripper {
	return interceptor{
		component:   component,
		transporter: t,
	}
}

func (i interceptor) RoundTrip(r *http.Request) (*http.Response, error) {

	span, _ := opentracing.StartSpanFromContext(
		r.Context(),
		r.Method+":"+r.URL.Path,
		ext.SpanKindClient(),
		ext.Component(i.component),
		ext.HTTPMethod(r.Method),
		ext.HTTPUrl(r.URL.String()),
	)

	tracer := opentracing.GlobalTracer()
	tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	res, err := i.transporter.RoundTrip(r)

	if err != nil {
		var errDetails = err.Error()
		span.LogFields(
			log.String("event", "client-transport-error"),
			log.String("error_details", errDetails))
		span.SetTag(ext.ErrorKey, true)
		span.Finish()
		return res, err
	}

	span.SetTag(ext.HTTPStatusCodeKey, res.StatusCode)
	if res.StatusCode >= 400 {
		span.SetTag(ext.ErrorKey, true)
		// errDetails, err := httputil.DumpResponse(res, true)
		// if err != nil {
		// 	errDetails = []byte("Cannot Parse Response")
		// }
		// span.SetTag(ctrace.ErrorDetailsKey, string(errDetails))
	}
	span.Finish()
	return res, nil
}
