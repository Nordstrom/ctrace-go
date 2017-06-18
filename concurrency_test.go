package ctrace_test

import (
	"bytes"
	"sync"

	ctrace "github.com/Nordstrom/ctrace-go"
	. "github.com/onsi/ginkgo"
	opentracing "github.com/opentracing/opentracing-go"
)

const op = "test"

var _ = Describe("Concurrency", func() {
	It("usage", func() {
		var buf bytes.Buffer
		trc := ctrace.Init(ctrace.TracerOptions{
			Writer: &buf,
			DebugAssertSingleGoroutine: true,
		})

		var wg sync.WaitGroup
		const num = 100
		wg.Add(num)
		for i := 0; i < num; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < num; j++ {
					sp := trc.StartSpan(op)
					sp.LogEvent("test event")
					sp.SetTag("foo", "bar")
					sp.SetBaggageItem("boo", "far")
					sp.SetOperationName("x")
					csp := trc.StartSpan(
						"csp",
						opentracing.ChildOf(sp.Context()))
					csp.Finish()
					defer sp.Finish()
				}
			}()
		}
		wg.Wait()
	})
})
