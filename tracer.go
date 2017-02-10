package ctrace

import (
	"io"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
)

// Options allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.
type Options struct {
	Writer io.Writer
}

// New creates a default Tracer.
func New() opentracing.Tracer {
	return NewWithOptions(Options{})
}

// NewWithOptions creates a customized Tracer.
func NewWithOptions(opts Options) opentracing.Tracer {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	return &ctracer{
		options: opts,
	}
}

// CTrace Implements the `Tracer` interface.
type ctracer struct {
	options Options
}

func (t *ctracer) StartSpan(
	operationName string,
	opts ...opentracing.StartSpanOption,
) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return t.StartSpanWithOptions(operationName, sso)
}

func (t *ctracer) StartSpanWithOptions(
	operationName string,
	opts opentracing.StartSpanOptions,
) opentracing.Span {
	s := spanPool.Get().(*cspan)
	return s.start(operationName, t, opts)
}

func (t *ctracer) Inject(sc opentracing.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return injectText(sc, carrier)
	case opentracing.Binary:
		return opentracing.ErrUnsupportedFormat
	}
	return opentracing.ErrUnsupportedFormat
}

func (t *ctracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return extractText(carrier)
	case opentracing.Binary:
		return nil, opentracing.ErrUnsupportedFormat
	}
	return nil, opentracing.ErrUnsupportedFormat
}
