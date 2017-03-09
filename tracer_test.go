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
		buf bytes.Buffer
		trc opentracing.Tracer
		out map[string]interface{}
	)

	Describe("New", func() {
		It("creates tracer with stdout writer", func() {
			trc = New()
			t := trc.(*tracer)
			Ω(t.options).ShouldNot(BeNil())
			Ω((t.options.Writer == os.Stdout)).Should(BeTrue())
		})
	})

	Describe("StartSpan", func() {
		JustBeforeEach(func() {
			buf.Reset()
			trc = NewWithOptions(TracerOptions{Writer: &buf, MultiEvent: true})
		})

		Context("with Tags", func() {
			It("outputs Start-Span", func() {
				_ = trc.StartSpan("x", opentracing.Tag{Key: "stag", Value: "sval"})
				Ω(buf.String()).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"tags":\{"stag":"sval"\},` +
						`"logs":\[\{"timestamp":\d+,"event":"Start-Span"\}\]\}`))
			})
		})

		Context("without Tags", func() {
			It("outputs Start-Span", func() {
				_ = trc.StartSpan("x")
				Ω(buf.String()).Should(MatchRegexp(
					`\{"traceId":"[0-9a-f]{16}","spanId":"[0-9a-f]{16}","operation":"x",` +
						`"start":\d+,"logs":\[\{"timestamp":\d+,"event":"Start-Span"\}\]\}`))
			})
		})

		Context("with ChildOf", func() {
			It("outputs Start-Span", func() {
				sc := spanContext{
					traceID: 123,
					spanID:  456,
				}
				_ = trc.StartSpan("x", opentracing.ChildOf(sc))
				Ω(buf.String()).Should(MatchRegexp(
					`\{"traceId":"000000000000007b","spanId":"[0-9a-f]{16}","parentId":"00000000000001c8",` +
						`"operation":"x","start":\d+,"logs":\[\{"timestamp":\d+,"event":"Start-Span"\}\]\}`))

			})
		})

		Context("with ChildOf and Baggage", func() {
			It("outputs Start-Span Baggage", func() {
				sc := spanContext{
					traceID: 123,
					spanID:  456,
					baggage: map[string]string{
						"btag1": "bval1",
						"btag2": "bval2",
					},
				}
				_ = trc.StartSpan("x", opentracing.ChildOf(sc))
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
			ctx    spanContext
			tracer opentracing.Tracer
		)
		JustBeforeEach(func() {
			ctx = spanContext{
				traceID: 123,
				spanID:  245,
			}
			tracer = New()
		})

		Context("without baggage", func() {
			It("injects HTTP Headers", func() {
				hdrs := http.Header{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				Ω(hdrs.Get("X-CT-Span-Id")).Should(Equal("f5"))
				Ω(hdrs.Get("X-CT-Trace-Id")).Should(Equal("7b"))
			})

			It("injects Text Map", func() {
				txt := map[string]string{}
				tracer.Inject(ctx, opentracing.TextMap, opentracing.TextMapCarrier(txt))
				Ω(txt["ct-span-id"]).Should(Equal("f5"))
				Ω(txt["ct-trace-id"]).Should(Equal("7b"))
			})
		})

		Context("with baggage", func() {
			It("injects HTTP Baggage Headers", func() {
				ctx.baggage = map[string]string{
					"bagitem1": "bagval1",
					"bagitem2": "bagval2",
				}
				hdrs := http.Header{}
				tracer.Inject(ctx, opentracing.HTTPHeaders, hdrs)
				Ω(hdrs.Get("X-CT-Bag-bagitem1")).Should(Equal("bagval1"))
				Ω(hdrs.Get("X-CT-Bag-bagitem2")).Should(Equal("bagval2"))
			})

			It("injects Text Map Baggage", func() {
				ctx.baggage = map[string]string{
					"bagitem1": "bagval1",
					"bagitem2": "bagval2",
				}
				txt := opentracing.TextMapCarrier{}
				tracer.Inject(ctx, opentracing.TextMap, txt)
				Ω(txt["ct-bag-bagitem1"]).Should(Equal("bagval1"))
				Ω(txt["ct-bag-bagitem2"]).Should(Equal("bagval2"))
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
				"X-CT-Span-Id":  []string{"f5"},
				"X-CT-Trace-Id": []string{"7b"},
			}
			tracer = New()
		})

		Context("without baggage", func() {
			It("extracts HTTP Headers", func() {
				c, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				Ω(err).ShouldNot(HaveOccurred())
				ctx := c.(spanContext)

				Ω(ctx.traceID).Should(Equal(uint64(123)))
				Ω(ctx.spanID).Should(Equal(uint64(245)))
			})
		})

		Context("with baggage", func() {
			It("extracts HTTP Baggage Headers", func() {
				hdrs["X-CT-Bag-bagitem1"] = []string{"bagval1"}
				hdrs["X-CT-Bag-bag-item2"] = []string{"bagval2"}

				c, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
				ctx := c.(spanContext)

				Ω(ctx.baggage["bagitem1"]).Should(Equal("bagval1"))
				Ω(ctx.baggage["bag-item2"]).Should(Equal("bagval2"))
			})
		})
	})
})
