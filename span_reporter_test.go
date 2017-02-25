package ctrace

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpanReporter", func() {

	var (
		buf bytes.Buffer
		rep SpanReporter
		out map[string]interface{}
	)

	BeforeEach(func() {
		buf.Reset()
		out = make(map[string]interface{})
		rep = NewSpanReporter(&buf, NewSpanEncoder())
	})

	Describe("Report", func() {
		It("reports one span", func() {
			sp := &span{
				operation: "op",
				context: spanContext{
					traceID: 123,
					spanID:  456,
				},
				duration: -1,
			}
			rep.Report(sp)
			Ω(lines(buf)[0]).Should(MatchRegexp(
				`\{"traceId":"000000000000007b","spanId":"00000000000001c8","operation":"op","start":\d{16}\}`))
		})

		It("reports two spans", func() {
			sp := &span{
				operation: "op",
				context: spanContext{
					traceID: 123,
					spanID:  456,
				},
				duration: -1,
			}
			rep.Report(sp)
			sp.duration = 35000
			rep.Report(sp)
			Ω(lines(buf)[0]).Should(MatchRegexp(
				`\{"traceId":"000000000000007b","spanId":"00000000000001c8","operation":"op","start":\d{16}\}`))
			Ω(lines(buf)[1]).Should(MatchRegexp(
				`\{"traceId":"000000000000007b","spanId":"00000000000001c8","operation":"op","start":\d{16},"duration":35\}`))
		})
	})
})
