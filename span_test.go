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
		buf bytes.Buffer
		trc opentracing.Tracer
		sp  opentracing.Span
		out map[string]interface{}
	)

	Context("with Single-Event Mode", func() {
		BeforeEach(func() {
			buf.Reset()
			trc = NewWithOptions(TracerOptions{Writer: &buf})
			sp = trc.StartSpan("x")
			out = make(map[string]interface{})
		})

		Describe("LogFields", func() {
			Context("without parent", func() {
				It("does not output before finish", func() {
					Ω(buf.String()).Should(HaveLen(0))
				})

				It("outputs string value", func() {
					sp.LogFields(log.String("key_str", "value"))
					sp.Finish()
					Ω(lines(buf)[0]).Should(MatchRegexp(
						`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"duration":\d+,"logs":\[\{"timestamp":\d+,"event":"Start-Span"\},` +
							`\{"timestamp":\d+,"key_str":"value"\},\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})

				It("outputs uint32 value", func() {
					sp.LogFields(log.Uint32("32bit", 4294967295))
					sp.Finish()
					Ω(lines(buf)[0]).Should(MatchRegexp(
						`{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"duration":\d+,"logs":\[\{"timestamp":\d+,"event":"Start-Span"\},` +
							`\{"timestamp":\d+,"32bit":4294967295\},\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})
			})
		})

		Describe("Finish", func() {
			Context("without parent", func() {
				It("outputs Finish-Span", func() {
					sp.Finish()
					Ω(lines(buf)[0]).Should(MatchRegexp(
						`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"duration":\d+,"logs":\[\{"timestamp":\d+,"event":"Start-Span"\},` +
							`\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})
			})

			Context("with parent", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs Finish-Span", func() {
					sp.Finish()
					Ω(lines(buf)[0]).Should(MatchRegexp(
						`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
							`"operation":"x","start":\d+,"duration":\d+,` +
							`"logs":\[\{"timestamp":\d+,"event":"Start-Span"\},\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})
			})

			Context("with parent and Baggage", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
						baggage: map[string]string{
							"btag1": "bval1",
							"btag2": "bval2",
						},
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs string value", func() {
					sp.Finish()
					if err := json.Unmarshal([]byte(lines(buf)[0]), &out); err != nil {
						Fail("Cannot unmarshal JSON")
					}
					bag := out["baggage"].(map[string]interface{})
					Ω(bag["btag1"]).Should(Equal("bval1"))
					Ω(bag["btag2"]).Should(Equal("bval2"))
				})
			})
		})
	})

	Context("with Multi-Event Mode", func() {
		BeforeEach(func() {
			buf.Reset()
			trc = NewWithOptions(TracerOptions{Writer: &buf, MultiEvent: true})
			sp = trc.StartSpan("x")
			out = make(map[string]interface{})
		})

		Describe("LogFields", func() {
			Context("without parent", func() {
				It("outputs string value", func() {
					sp.LogFields(log.String("key_str", "value"))
					Ω(lines(buf)[1]).Should(MatchRegexp(
						`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"logs":\[\{"timestamp":\d+,"key_str":"value"}\]\}`))
				})

				It("outputs uint32 value", func() {
					sp.LogFields(log.Uint32("32bit", 4294967295))
					Ω(lines(buf)[1]).Should(MatchRegexp(
						`{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"logs":\[\{"timestamp":\d+,` +
							`"32bit":4294967295\}\]\}`))
				})
			})

			Context("with parent", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs string value", func() {
					sp.LogFields(log.String("key_str", "value"))
					Ω(lines(buf)[1]).Should(MatchRegexp(
						`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
							`"operation":"x","start":\d+,"logs":\[\{"timestamp":\d+,"key_str":"value"\}\]\}`))
				})
			})

			Context("with parent and Baggage", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
						baggage: map[string]string{
							"btag1": "bval1",
							"btag2": "bval2",
						},
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs string value", func() {
					sp.LogFields(log.String("key_str", "value"))
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
				sp.LogKV("lkey1", "lval1", "lkey2", "lval2")
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"logs":\[\{"timestamp":\d+,"lkey1":"lval1","lkey2":"lval2"\}\]\}`))
			})
		})

		Describe("LogEvent", func() {
			It("outputs log record", func() {
				sp.LogEvent("evt1")
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"logs":\[\{"timestamp":\d+,"event":"evt1"\}\]\}`))
			})
		})

		Describe("SetTag", func() {
			It("outputs on Finish", func() {
				sp.SetTag("ftag", "fval")
				sp.Finish()
				Ω(lines(buf)[1]).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"duration":\d+,"tags":\{"ftag":"fval"},` +
						`"logs":\[\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
			})
		})

		Describe("SetOperationName", func() {
			It("sets data.operation", func() {
				sp.SetOperationName("newname")
				cs := sp.(*span)
				Ω(cs.operation).Should(Equal("newname"))
			})
		})

		Describe("Tracer", func() {
			It("gets tracer pointer", func() {
				Ω(sp.Tracer()).Should(Equal(trc))
			})
		})

		Describe("Finish", func() {
			Context("without parent", func() {
				It("outputs Finish-Span", func() {
					sp.Finish()
					Ω(lines(buf)[1]).Should(MatchRegexp(
						`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
							`"start":\d+,"duration":\d+,"logs":\[\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})
			})

			Context("with parent", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs Finish-Span", func() {
					sp.Finish()
					Ω(lines(buf)[1]).Should(MatchRegexp(
						`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
							`"operation":"x","start":\d+,"duration":\d+,"logs":\[\{"timestamp":\d+,"event":"Finish-Span"\}\]\}`))
				})
			})

			Context("with parent and Baggage", func() {
				JustBeforeEach(func() {
					sc := spanContext{
						traceID: 123,
						spanID:  456,
						baggage: map[string]string{
							"btag1": "bval1",
							"btag2": "bval2",
						},
					}
					buf.Reset()
					sp = trc.StartSpan("x", opentracing.ChildOf(sc))
				})
				It("outputs string value", func() {
					sp.Finish()
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
})
