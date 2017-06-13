package ctrace

import (
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var (
	// ChildOf returns a StartSpanOption pointing to a dependent parent span.
	// If sc == nil, the option has no effect.
	//
	// See ChildOfRef, SpanReference
	ChildOf = opentracing.ChildOf
)

// Tracer is a simple, thin interface for Span creation and SpanContext
// propagation.
type Tracer interface {
	opentracing.Tracer
	StartSpanWithOptions(string, opentracing.StartSpanOptions) opentracing.Span
}

// Tracer Implements the `Tracer` interface.
type tracer struct {
	options TracerOptions
	SpanReporter
	spanPool *sync.Pool
	rng      *rand.Rand
	sync.Mutex
	textMapPropagator     *textMapPropagator
	httpHeadersPropagator *textMapPropagator
}

// TracerOptions allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.
type TracerOptions struct {
	// MultiEvent tells whether the tracer outputs in Single-Event or Multi-Event Mode.
	// See [Canonical Events](https://github.com/Nordstrom/ctrace/tree/new#canonical-events).
	// If MultiEvent=true, the tracer is using Multi-Event Mode which means Start-Span, Log,
	// and Finish-Span events are output with each containing a single log.
	// If MultiEvent=false (default), the tracer is using Single-Event Mode which
	// means only Finish-Span events are output with a collectionn of all logs for that Span.
	MultiEvent bool

	// Writer is used to write serialized trace events.  It defaults to os.Stdout.
	Writer io.Writer

	// ServiceName allows the configuration of the "service" tag for the entire Tracer.
	// If not specified here, it can also be specified using environment variable "CTRACE_SERVICE"
	ServiceName string

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

func init() {
	Init(TracerOptions{})
}

// Init initializes the global Tracer returned by Global().
func Init(opts TracerOptions) {
	opentracing.SetGlobalTracer(NewWithOptions(opts))
}

// Global returns the global Tracer
func Global() Tracer {
	return opentracing.GlobalTracer().(Tracer)
}

// New creates a default Tracer.
func New() Tracer {
	return NewWithOptions(TracerOptions{
		MultiEvent: true,
		Writer:     nil,
	})
}

// NewWithOptions creates a customized Tracer.
func NewWithOptions(opts TracerOptions) Tracer {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	if opts.ServiceName == "" {
		opts.ServiceName = os.Getenv("CTRACE_SERVICE_NAME")
	}

	return &tracer{
		options:               opts,
		SpanReporter:          NewSpanReporter(opts.Writer, NewSpanEncoder()),
		spanPool:              &sync.Pool{New: func() interface{} { return &span{} }},
		rng:                   rand.New(rand.NewSource(time.Now().UnixNano())),
		textMapPropagator:     newTextMapPropagator(),
		httpHeadersPropagator: newHTTPHeadersPropagator(),
	}
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
	sp := t.newSpan()

	// Start time.
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	sp.tracer = t
	sp.start = startTime
	sp.operation = operationName
	sp.tags = opts.Tags

	if t.options.ServiceName != "" {
		sp.setTag("service", t.options.ServiceName)
	}

	sp.logs = make([]opentracing.LogRecord, 0, 10)
	sp.logs = append(sp.logs, opentracing.LogRecord{
		Timestamp: startTime,
		Fields:    []log.Field{log.String("event", "Start-Span")},
	})

	if t.options.DebugAssertSingleGoroutine {
		sp.setTag(debugGoroutineIDTag, curGoroutineID())
	}

	// Look for a parent in the list of References.
	//
	// TODO: would be nice if basictracer did something with all
	// References, not just the first one.
	for _, ref := range opts.References {
		refCtx := ref.ReferencedContext.(spanContext)
		sp.context.traceID = refCtx.traceID
		sp.context.spanID = t.randomID()
		sp.parentID = refCtx.spanID

		if l := len(refCtx.baggage); l > 0 {
			sp.context.baggage = make(map[string]string, l)
			for k, v := range refCtx.baggage {
				sp.context.baggage[k] = v
			}
		}
		break
	}
	if sp.context.traceID == 0 {
		// No parent Span found; allocate new trace and span ids and determine
		// the Sampled status.
		sp.context.traceID = t.randomID()
		sp.context.spanID = sp.context.traceID
	}

	if t.options.MultiEvent {
		t.Report(sp)
	}
	return sp
}

func (t *tracer) Inject(sc opentracing.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case opentracing.TextMap:
		return t.textMapPropagator.Inject(sc, carrier)
	case opentracing.HTTPHeaders:
		return t.httpHeadersPropagator.Inject(sc, carrier)
	case opentracing.Binary:
		return opentracing.ErrUnsupportedFormat
	}
	return opentracing.ErrUnsupportedFormat
}

func (t *tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap:
		return t.textMapPropagator.Extract(carrier)
	case opentracing.HTTPHeaders:
		return t.httpHeadersPropagator.Extract(carrier)
	case opentracing.Binary:
		return nil, opentracing.ErrUnsupportedFormat
	}
	return nil, opentracing.ErrUnsupportedFormat
}

// newSpan retrieves an instance of a clean Span object.
func (t *tracer) newSpan() *span {
	sp := t.spanPool.Get().(*span)
	sp.context = spanContext{}
	sp.duration = -1
	sp.tracer = nil
	sp.tags = nil
	sp.logs = nil
	sp.prefix = nil
	return sp
}

func (t *tracer) freeSpan(sp *span) {
	t.spanPool.Put(sp)
}

func (t *tracer) randomNumber() uint64 {
	return uint64(t.rng.Int63())
}

// randomID generates a random trace/span ID, using tracer.random() generator.
// It never returns 0.
func (t *tracer) randomID() uint64 {
	t.Lock()
	defer t.Unlock()

	val := t.randomNumber()
	for val == 0 {
		val = t.randomNumber()
	}
	return val
}
