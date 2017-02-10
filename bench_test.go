package ctrace

import (
	"os"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func BenchmarkSpan(b *testing.B) {
	f, err := os.Create("dump.json")
	if err != nil {
		return
	}
	t := NewWithOptions(Options{Writer: f})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parent := t.StartSpan("parent",
			SpanKindServer(),
			Component("component"),
			PeerHostname("hostname"),
			PeerHostIPv6("ip"),
			HTTPMethod("method"),
			HTTPUrl("https://some.url.outthere.com"),
		)

		child := t.StartSpan("child",
			opentracing.ChildOf(parent.Context()),
			SpanKindServer(),
			Component("child-component"),
			PeerHostname("hostname"),
			PeerHostIPv6("ip"),
			HTTPMethod("method"),
			HTTPUrl("https://some.url.outthere.com"),
		)

		child.LogFields(log.String("event", "event"))
		child.Finish()
		parent.Finish()
	}
	b.StopTimer()
}
