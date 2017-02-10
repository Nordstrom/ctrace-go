package ctrace

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
)

var _ = Describe("Tracer", func() {

	var (
		buf    bytes.Buffer
		tracer opentracing.Tracer
		out    map[string]interface{}
	)

	Describe("New", func() {
		It("creates tracer with stdout writer", func() {
			tracer = New()
			t := tracer.(*ctracer)
			Ω(t.options).ShouldNot(BeNil())
			Ω((t.options.Writer == os.Stdout)).Should(BeTrue())
		})
	})

	Describe("StartSpan", func() {
		JustBeforeEach(func() {
			buf.Reset()
			tracer = NewWithOptions(Options{Writer: &buf})
		})

		Context("with Tags", func() {
			It("outputs Start-Span", func() {
				_ = tracer.StartSpan("x", opentracing.Tag{Key: "stag", Value: "sval"})
				Ω(buf.String()).Should(MatchRegexp(
					"\\{\"traceId\":\"[0-9a-f]{16}\",\"spanId\":\"[0-9a-f]{16}\",\"operation\":\"x\"," +
						"\"start\":\\d+,\"tags\":\\{\"stag\":\"sval\"\\},\"log\":\\{\"timestamp\":\\d+," +
						"\"event\":\"Start-Span\"}\\}"))
			})
		})

		Context("without Tags", func() {
			It("outputs Start-Span", func() {
				_ = tracer.StartSpan("x")
				Ω(buf.String()).Should(MatchRegexp(
					"\\{\"traceId\":\"[0-9a-f]{16}\",\"spanId\":\"[0-9a-f]{16}\",\"operation\":\"x\"," +
						"\"start\":\\d+,\"log\":\\{\"timestamp\":\\d+,\"event\":\"Start-Span\"}\\}"))
			})
		})

		Context("with ChildOf", func() {
			It("outputs Start-Span", func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
				}
				_ = tracer.StartSpan("x", opentracing.ChildOf(sc))
				Ω(buf.String()).Should(MatchRegexp(
					`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
						`"operation":"x","start":\d+,"log":\{"timestamp":\d+,"event":"Start-Span"}\}`))

			})
		})

		Context("with ChildOf and Baggage", func() {
			It("outputs Start-Span Baggage", func() {
				sc := SpanContext{
					TraceID: 123,
					SpanID:  456,
					Baggage: map[string]string{
						"btag1": "bval1",
						"btag2": "bval2",
					},
				}
				_ = tracer.StartSpan("x", opentracing.ChildOf(sc))
				if err := json.Unmarshal([]byte(buf.String()), &out); err != nil {
					Fail("Cannot unmarshal JSON")
				}
				bag := out["baggage"].(map[string]interface{})
				Ω(bag["btag1"]).Should(Equal("bval1"))
				Ω(bag["btag2"]).Should(Equal("bval2"))
			})
		})

	})

	Describe("Inject", func() {
		var (
			ctx    SpanContext
			tracer opentracing.Tracer
		)
		JustBeforeEach(func() {
			ctx = SpanContext{
				TraceID: 123,
				SpanID:  245,
			}
			tracer = New()
		})

		Context("without baggage", func() {
			It("injects HTTP Headers", func() {
				hdrs := http.Header{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				Ω(hdrs.Get("X-Request-Id")).Should(Equal("f5"))
				Ω(hdrs.Get("X-Correlation-Id")).Should(Equal("7b"))
			})

			It("injects Text Map", func() {
				txt := map[string]string{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, opentracing.TextMapCarrier(txt))
				Ω(txt["X-Request-Id"]).Should(Equal("f5"))
				Ω(txt["X-Correlation-Id"]).Should(Equal("7b"))
			})
		})

		Context("with baggage", func() {
			It("injects HTTP Baggage Headers", func() {
				ctx.Baggage = map[string]string{
					"bagitem1": "bagval1",
					"bagitem2": "bagval2",
				}
				hdrs := http.Header{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, hdrs)
				Ω(hdrs.Get("X-Baggage-bagitem1")).Should(Equal("bagval1"))
				Ω(hdrs.Get("X-Baggage-bagitem2")).Should(Equal("bagval2"))
			})

			It("injects Text Map Baggage", func() {
				ctx.Baggage = map[string]string{
					"bagitem1": "bagval1",
					"bagitem2": "bagval2",
				}
				txt := opentracing.TextMapCarrier{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, txt)
				Ω(txt["X-Baggage-bagitem1"]).Should(Equal("bagval1"))
				Ω(txt["X-Baggage-bagitem2"]).Should(Equal("bagval2"))
			})
		})
	})

	Describe("Extract", func() {
		var (
			hdrs   http.Header
			tracer opentracing.Tracer
		)
		JustBeforeEach(func() {
			hdrs = http.Header{
				"X-Request-Id":     []string{"f5"},
				"X-Correlation-Id": []string{"7b"},
			}
			tracer = New()
		})

		Context("without baggage", func() {
			It("extracts HTTP Headers", func() {
				c, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				ctx := c.(SpanContext)

				Ω(ctx.TraceID).Should(Equal(uint64(123)))
				Ω(ctx.SpanID).Should(Equal(uint64(245)))
			})
		})

		Context("with baggage", func() {
			It("extracts HTTP Baggage Headers", func() {
				hdrs.Add("X-Baggage-bagitem1", "bagval1")
				hdrs.Add("X-Baggage-bagitem2", "bagval2")

				c, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				ctx := c.(SpanContext)

				Ω(ctx.Baggage["bagitem1"]).Should(Equal("bagval1"))
				Ω(ctx.Baggage["bagitem2"]).Should(Equal("bagval2"))
			})
		})
	})
})
