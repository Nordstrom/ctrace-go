package ctrace

import (
	"io"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
)

// Options allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.
type Options struct {
	// Writer is used to write serialized trace events.  It defaults to os.Stdout.
	Writer io.Writer
	// DebugAssertSingleGoroutine internally records the ID of the goroutine
	// creating each Span and verifies that no operation is carried out on
	// it on a different goroutine.
	// Provided strictly for development purposes.
	// Passing Spans between goroutine without proper synchronization often
	// results in use-after-Finish() errors. For a simple example, consider the
	// following pseudocode:
	//
	//  func (s *Server) Handle(req http.Request) error {
	//    sp := s.StartSpan("server")
	//    defer sp.Finish()
	//    wait := s.queueProcessing(opentracing.ContextWithSpan(context.Background(), sp), req)
	//    select {
	//    case resp := <-wait:
	//      return resp.Error
	//    case <-time.After(10*time.Second):
	//      sp.LogEvent("timed out waiting for processing")
	//      return ErrTimedOut
	//    }
	//  }
	//
	// This looks reasonable at first, but a request which spends more than ten
	// seconds in the queue is abandoned by the main goroutine and its trace
	// finished, leading to use-after-finish when the request is finally
	// processed. Note also that even joining on to a finished Span via
	// StartSpanWithOptions constitutes an illegal operation.
	//
	// Code bases which do not require (or decide they do not want) Spans to
	// be passed across goroutine boundaries can run with this flag enabled in
	// tests to increase their chances of spotting wrong-doers.
	DebugAssertSingleGoroutine bool
	// DebugAssertUseAfterFinish is provided strictly for development purposes.
	// When set, it attempts to exacerbate issues emanating from use of Spans
	// after calling Finish by running additional assertions.
	DebugAssertUseAfterFinish bool
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
