package ctrace

import (
	"bytes"
	"sync"

	. "github.com/onsi/ginkgo"
	opentracing "github.com/opentracing/opentracing-go"
)

const op = "test"

var _ = Describe("Concurrency", func() {
	It("usage", func() {
		var buf bytes.Buffer
		tracer := NewWithOptions(Options{
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
					sp := tracer.StartSpan(op)
					sp.LogEvent("test event")
					sp.SetTag("foo", "bar")
					sp.SetBaggageItem("boo", "far")
					sp.SetOperationName("x")
					csp := tracer.StartSpan(
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
