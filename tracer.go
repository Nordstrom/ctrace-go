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

// New creates a customized Tracer.
func New(opts Options) opentracing.Tracer {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	return &tracer{
		options: opts,
	}
}

// Implements the `Tracer` interface.
type tracer struct {
	options Options
}

func (t *tracer) StartSpan(
	operationName string,
	opts ...opentracing.StartSpanOption,
) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return t.StartSpanWithOptions(operationName, sso)
}

func (t *tracer) StartSpanWithOptions(
	operationName string,
	opts opentracing.StartSpanOptions,
) opentracing.Span {
	s := spanPool.Get().(*span)
	return s.start(operationName, t, opts)
}

func (t *tracer) Inject(sc opentracing.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return injectText(sc, carrier)
	case opentracing.Binary:
		return opentracing.ErrUnsupportedFormat
	}
	return opentracing.ErrUnsupportedFormat
}

func (t *tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return extractText(carrier)
	case opentracing.Binary:
		return nil, opentracing.ErrUnsupportedFormat
	}
	return nil, opentracing.ErrUnsupportedFormat
}
