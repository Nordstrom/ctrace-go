package ctrace

import (
	"bytes"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func lines(buf bytes.Buffer) []string {
	return strings.Split(buf.String(), "\n")
}

var _ = Describe("Span", func() {

	var (
		buf    bytes.Buffer
		tracer opentracing.Tracer
		span   opentracing.Span
		out    map[string]interface{}
	)

	BeforeEach(func() {
		buf.Reset()
		tracer = NewWithOptions(Options{Writer: &buf})
		span = tracer.StartSpan("x")
		out = make(map[string]interface{})
	})

	Describe("LogFields", func() {
		Context("without parent", func() {
			It("outputs string value", func() {
				span.LogFields(log.String("key_str", "value"))
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"log":\{"timestamp":\d+,"key_str":"value"}\}`))
			})

			It("outputs uint32 value", func() {
				span.LogFields(log.Uint32("32bit", 4294967295))
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"log":\{"timestamp":\d+,` +
						`"32bit":4294967295}\}`))
			})
		})

		Context("with parent", func() {
			JustBeforeEach(func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
				}
				buf.Reset()
				span = tracer.StartSpan("x", opentracing.ChildOf(sc))
			})
			It("outputs string value", func() {
				span.LogFields(log.String("key_str", "value"))
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
						`"operation":"x","start":\d+,"log":\{"timestamp":\d+,"key_str":"value"}\}`))
			})
		})

		Context("with parent and Baggage", func() {
			JustBeforeEach(func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
					Baggage: map[string]string{
						"btag1": "bval1",
						"btag2": "bval2",
					},
				}
				buf.Reset()
				span = tracer.StartSpan("x", opentracing.ChildOf(sc))
			})
			It("outputs string value", func() {
				span.LogFields(log.String("key_str", "value"))
				if err := json.Unmarshal([]byte(lines(buf)[1]), &out); err != nil {
					Fail("Cannot unmarshal JSON")
				}
				bag := out["baggage"].(map[string]interface{})
				Ω(bag["btag1"]).Should(Equal("bval1"))
				Ω(bag["btag2"]).Should(Equal("bval2"))
			})
		})
	})

	Describe("LogKV", func() {
		It("outputs log record", func() {
			span.LogKV("lkey1", "lval1", "lkey2", "lval2")
			Ω(lines(buf)[1]).Should(MatchRegexp(
				`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
					`"start":\d+,"log":\{"timestamp":\d+,"lkey1":"lval1","lkey2":"lval2"}\}`))
		})
	})

	Describe("LogEvent", func() {
		It("outputs log record", func() {
			span.LogEvent("evt1")
			Ω(lines(buf)[1]).Should(MatchRegexp(
				`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
					`"start":\d+,"log":\{"timestamp":\d+,"event":"evt1"}\}`))
		})
	})

	Describe("SetTag", func() {
		It("outputs on Finish", func() {
			span.SetTag("ftag", "fval")
			span.Finish()
			Ω(lines(buf)[1]).Should(MatchRegexp(
				`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
					`"start":\d+,"duration":\d+,"tags":\{"ftag":"fval"},` +
					`"log":\{"timestamp":\d+,"event":"Finish-Span"}\}`))
		})
	})

	Describe("SetOperationName", func() {
		It("sets data.operation", func() {
			span.SetOperationName("newname")
			cs := span.(*cspan)
			Ω(cs.data.operation).Should(Equal("newname"))
		})
	})

	Describe("Tracer", func() {
		It("gets tracer pointer", func() {
			Ω(span.Tracer()).Should(Equal(tracer))
		})
	})

	Describe("Finish", func() {
		Context("without parent", func() {
			It("outputs Finish-Span", func() {
				span.Finish()
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"duration":\d+,"log":\{"timestamp":\d+,"event":"Finish-Span"}\}`))
			})
		})

		Context("with parent", func() {
			JustBeforeEach(func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
				}
				buf.Reset()
				span = tracer.StartSpan("x", opentracing.ChildOf(sc))
			})
			It("outputs Finish-Span", func() {
				span.Finish()
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
						`"operation":"x","start":\d+,"duration":\d+,"log":\{"timestamp":\d+,"event":"Finish-Span"}\}`))
			})
		})

		Context("with parent and Baggage", func() {
			JustBeforeEach(func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
					Baggage: map[string]string{
						"btag1": "bval1",
						"btag2": "bval2",
					},
				}
				buf.Reset()
				span = tracer.StartSpan("x", opentracing.ChildOf(sc))
			})
			It("outputs string value", func() {
				span.Finish()
				if err := json.Unmarshal([]byte(lines(buf)[1]), &out); err != nil {
					Fail("Cannot unmarshal JSON")
				}
				bag := out["baggage"].(map[string]interface{})
				Ω(bag["btag1"]).Should(Equal("bval1"))
				Ω(bag["btag2"]).Should(Equal("bval2"))
			})
		})
	})
})
