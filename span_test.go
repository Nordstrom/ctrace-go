package ctrace

import (
	"bytes"
	"strings"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

// func TestSpan_Baggage(t *testing.T) {
// 	var buf bytes.Buffer
// 	tracer := New(Options{Writer: &buf})
// 	span := tracer.StartSpan("x")
// 	span.SetBaggageItem("x", "y")
// 	assert.Equal(t, "y", span.BaggageItem("x"))
// 	span.Finish()
// 	s := buf.String()
// 	assert.Equal(t, "hello", s)
//
// 	span = tracer.StartSpan("x")
// 	span.SetBaggageItem("x", "y")
// 	baggage := make(map[string]string)
// 	span.Context().ForeachBaggageItem(func(k, v string) bool {
// 		baggage[k] = v
// 		return true
// 	})
// 	assert.Equal(t, map[string]string{"x": "y"}, baggage)
//
// 	span.SetBaggageItem("a", "b")
// 	baggage = make(map[string]string)
// 	span.Context().ForeachBaggageItem(func(k, v string) bool {
// 		baggage[k] = v
// 		return false // exit early
// 	})
// 	assert.Equal(t, 1, len(baggage))
// 	span.Finish()
// 	// spans = recorder.GetSpans()
// 	// assert.Equal(t, 1, len(spans))
// 	// assert.Equal(t, 2, len(spans[0].Context.Baggage))
// }

func TestSpan_SingleLoggedTaggedSpan(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(Options{Writer: &buf})
	span := tracer.StartSpan("x", opentracing.Tag{Key: "stag", Value: "sval"})
	// span.LogEventWithPayload("event", "payload")
	span.LogFields(log.String("key_str", "value"), log.Uint32("32bit", 4294967295))
	span.SetTag("ftag", "fval")
	span.Finish()
	lines := strings.Split(buf.String(), "\n")
	assert.Regexp(t, "\\{\"traceId\":\"[0-9a-f]{16}\",\"spanId\":\"[0-9a-f]{16}\",\"operation\":\"x\",\"start\":\\d+,\"tags\":\\{\"stag\":\"sval\"\\},\"log\":\\{\"timestamp\":\\d+,\"event\":\"Start-Span\"}\\}", lines[0])
	// assert.Regexp(t, "\\{\"traceId\":\"[0-9a-f]{16}\",\"spanId\":\"[0-9a-f]{16}\",\"operation\":\"x\",\"start\":\\d+,\"tags\":\\{\"stag\":\"sval\"\\},\"log\":\\{\"timestamp\":\\d+,\"event\":\"event\",\"payload\":\"payload\"}\\}", lines[1])
	assert.Regexp(t, "\\{\"traceId\":\"[0-9a-f]{16}\",\"spanId\":\"[0-9a-f]{16}\",\"operation\":\"x\",\"start\":\\d+,\"tags\":\\{\"stag\":\"sval\"\\},\"log\":\\{\"timestamp\":\\d+,\"key_str\":\"value\",\"32bit\":4294967295}\\}", lines[1])
}
