package core

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var _ = Describe("SpanEncoder", func() {

	var (
		enc   SpanEncoder
		bytes []byte
	)

	BeforeEach(func() {
		enc = NewSpanEncoder()
	})

	Describe("Encode", func() {
		It("encodes minimal span", func() {
			sp := &span{
				operation: "op",
				context: spanContext{
					traceID: 123,
					spanID:  456,
				},
				duration: -1,
			}
			bytes = enc.Encode(sp)
			Ω(string(bytes)).Should(MatchRegexp(
				`\{"traceId":"000000000000007b","spanId":"00000000000001c8","operation":"op","start":\d{16}\}`))
		})

		It("encodes full span", func() {
			sp := &span{
				operation: "op",
				context: spanContext{
					traceID: 123,
					spanID:  456,
					baggage: map[string]string{
						"bkey1": "bval1",
					},
				},
				parentID: 789,
				start:    time.Now(),
				duration: 35000,
				tags: map[string]interface{}{
					"key1": "val1",
				},
				logs: []opentracing.LogRecord{
					{
						Timestamp: time.Now(),
						Fields: []log.Field{
							log.String("event", "evt1"),
							log.Int("key1", 99),
						},
					},
				},
			}
			bytes = enc.Encode(sp)
			Ω(string(bytes)).Should(MatchRegexp(
				`\{"traceId":"000000000000007b","spanId":"00000000000001c8","parentId":"0000000000000315",` +
					`"operation":"op","start":\d{16},"duration":35,` +
					`"tags":\{"key1":"val1"\},` +
					`"baggage":\{"bkey1":"bval1"\},` +
					`"logs":\[\{"timestamp":\d{16},"event":"evt1","key1":99\}\]\}`))

		})
	})
})
