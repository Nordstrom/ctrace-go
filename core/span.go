package core

import (
	"fmt"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// SpanContext represents Span state that must propagate to descendant Spans and across process
// boundaries (e.g., a <trace_id, span_id, sampled> tuple).
type SpanContext interface {
	opentracing.SpanContext
	TraceID() string
	SpanID() string
	BaggageItem(key string) string
}

// spanContext holds the basic Span metadata.
type spanContext struct {
	// A probabilistically unique identifier for a [multi-span] trace.
	traceID uint64

	// A probabilistically unique identifier for a span.
	spanID uint64

	// The span's associated baggage.
	baggage map[string]string // initialized on first use
}

// NewSpanContext creates a new SpanContext
func NewSpanContext(
	traceID uint64,
	spanID uint64,
	baggage map[string]string,
) SpanContext {
	return spanContext{
		traceID: traceID,
		spanID:  spanID,
		baggage: baggage,
	}
}

func (c spanContext) TraceID() string {
	return fmt.Sprintf("%016x", c.traceID)
}

func (c spanContext) SpanID() string {
	return fmt.Sprintf("%016x", c.spanID)
}

func (c spanContext) BaggageItem(key string) string {
	if c.baggage == nil {
		return ""
	}
	return c.baggage[key]
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c spanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem returns an entirely new basictracer SpanContext with the
// given key:value baggage pair set.
func (c spanContext) WithBaggageItem(key, val string) spanContext {
	if c.baggage == nil {
		return spanContext{c.traceID, c.spanID, map[string]string{key: val}}
	}
	var newBaggage = make(map[string]string, len(c.baggage)+1)
	for k, v := range c.baggage {
		newBaggage[k] = v
	}
	newBaggage[key] = val

	// Use positional parameters so the compiler will help catch new fields.
	return spanContext{c.traceID, c.spanID, newBaggage}
}

// Span represents an active, un-finished span in the OpenTracing system.
//
// Spans are created by the Tracer interface.
type Span interface {
	opentracing.Span
	RawContext() SpanContext
	RawTracer() Tracer
}

type span struct {
	tracer     *tracer
	sync.Mutex // protects the fields below

	context spanContext

	// The SpanID of this SpanContext's first intra-trace reference (i.e.,
	// "parent"), or 0 if there is no parent.
	parentID uint64

	// The name of the "operation" this span is an instance of. (Called a "span
	// name" in some implementations)
	operation string

	// We store <start, duration> rather than <start, end> so that only
	// one of the timestamps has global clock uncertainty issues.
	start    time.Time
	finish   time.Time
	duration time.Duration

	// Essentially an extension mechanism. Can be used for many purposes,
	// not to be enumerated here.
	tags map[string]interface{}

	logs []opentracing.LogRecord

	prefix []byte
}

func (s *span) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.operation = operationName
	return s
}

func (s *span) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	return s.setTag(key, value)
}

func (s *span) setTag(key string, value interface{}) opentracing.Span {
	if s.tags == nil {
		s.tags = make(map[string]interface{})
	}
	s.tags[key] = value
	return s
}

func (s *span) LogKV(keyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(keyValues...)
	if err != nil {
		s.LogFields(log.Error(err), log.String("function", "LogKV"))
		return
	}
	s.LogFields(fields...)
}

func (s *span) LogFields(fields ...log.Field) {
	l := opentracing.LogRecord{
		Fields: fields,
	}
	s.reportLog(l)
}

func (s *span) reportLog(l opentracing.LogRecord) {
	s.Lock()
	defer s.Unlock()
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	if s.tracer.options.MultiEvent {
		s.logs[0] = l
		s.tracer.Report(s)
	} else {
		s.logs = append(s.logs, l)
	}
}

func (s *span) LogEvent(event string) {
	s.Log(opentracing.LogData{
		Event: event,
	})
}

func (s *span) LogEventWithPayload(event string, payload interface{}) {
	s.Log(opentracing.LogData{
		Event:   event,
		Payload: payload,
	})
}

func (s *span) Log(ld opentracing.LogData) {
	s.reportLog(ld.ToLogRecord())
}

func (s *span) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

func (s *span) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	duration := finishTime.Sub(s.start)

	s.Lock()
	defer s.Unlock()

	for _, lr := range opts.LogRecords {
		s.reportLog(lr)
	}
	for _, ld := range opts.BulkLogData {
		s.reportLog(ld.ToLogRecord())
	}

	s.finish = finishTime
	s.duration = duration

	log := opentracing.LogRecord{
		Timestamp: finishTime,
		Fields:    []log.Field{log.String("event", "Finish-Span")},
	}

	if s.tracer.options.MultiEvent {
		s.logs[0] = log
	} else {
		s.logs = append(s.logs, log)
	}

	s.tracer.Report(s)
	t := s.tracer
	if s.tracer.options.DebugAssertUseAfterFinish {
		// This makes it much more likely to catch a panic on any subsequent
		// operation since s.tracer is accessed on every call to `Lock`.
		// We don't call `reset()` here to preserve the logs in the Span
		// which are printed when the assertion triggers.
		s.tracer = nil
	}
	t.freeSpan(s)
}

func (s *span) Context() opentracing.SpanContext {
	return s.context
}

func (s *span) RawContext() SpanContext {
	return s.context
}

func (s *span) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *span) RawTracer() Tracer {
	return s.tracer
}

func (s *span) SetBaggageItem(key, val string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.context = s.context.WithBaggageItem(key, val)
	return s
}

func (s *span) BaggageItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.context.baggage[key]
}
