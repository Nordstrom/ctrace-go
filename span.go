package ctrace

import (
	"io"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type spanData struct {
	context SpanContext

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
}

// Implements the `Span` interface. Created via tracerImpl (see
// `basictracer.New()`).
type span struct {
	tracer *tracer
	io.Writer
	Encoder

	sync.Mutex // protects the fields below

	data spanData
}

var spanPool = &sync.Pool{New: func() interface{} {
	return &span{}
}}

func (s *span) start(
	operation string,
	t *tracer,
	opts opentracing.StartSpanOptions,
) opentracing.Span {
	s.tracer = t
	s.Writer = t.options.Writer
	s.Encoder = NewJSONEncoder()

	// Start time.
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	s.data = spanData{
		start:     startTime,
		operation: operation,
		context:   SpanContext{},
		tags:      opts.Tags,
		duration:  -1,
	}

	// Look for a parent in the list of References.
	//
	// TODO: would be nice if basictracer did something with all
	// References, not just the first one.
	for _, ref := range opts.References {
		refCtx := ref.ReferencedContext.(SpanContext)
		s.data.context.TraceID = refCtx.TraceID
		s.data.context.SpanID = randomID()
		s.data.parentID = refCtx.SpanID

		if l := len(refCtx.Baggage); l > 0 {
			s.data.context.Baggage = make(map[string]string, l)
			for k, v := range refCtx.Baggage {
				s.data.context.Baggage[k] = v
			}
		}
		break
	}
	if s.data.context.TraceID == 0 {
		// No parent Span found; allocate new trace and span ids and determine
		// the Sampled status.
		s.data.context.TraceID, s.data.context.SpanID = randomID2()
	}

	s.WriteStart(s, s.data)
	return s
}

func (s *span) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.data.operation = operationName
	return s
}

func (s *span) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	if s.data.tags == nil {
		s.data.tags = make(map[string]interface{})
	}
	s.data.tags[key] = value
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
	s.writeLog(l)
}

func (s *span) writeLog(l opentracing.LogRecord) {
	s.Lock()
	defer s.Unlock()
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	s.data.log = l
	s.WriteLog(s, s.data)
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
	s.writeLog(ld.ToLogRecord())
}

func (s *span) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

func (s *span) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	duration := finishTime.Sub(s.data.start)

	s.Lock()
	defer s.Unlock()

	for _, lr := range opts.LogRecords {
		s.writeLog(lr)
	}
	for _, ld := range opts.BulkLogData {
		s.writeLog(ld.ToLogRecord())
	}

	s.data.finish = finishTime
	s.data.duration = duration

	s.WriteFinish(s, s.data)

	spanPool.Put(s)
}

func (s *span) Context() opentracing.SpanContext {
	return s.data.context
}

func (s *span) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *span) SetBaggageItem(key, val string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.data.context = s.data.context.WithBaggageItem(key, val)
	return s
}

func (s *span) BaggageItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.data.context.Baggage[key]
}
