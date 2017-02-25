package ctrace

import (
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

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

	log opentracing.LogRecord

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
	s.log = l
	s.tracer.Report(s)
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

	s.log = opentracing.LogRecord{
		Timestamp: finishTime,
		Fields:    []log.Field{log.String("event", "Finish-Span")},
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

func (s *span) Tracer() opentracing.Tracer {
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
