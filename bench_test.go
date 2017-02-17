package ctrace

import (
	"os"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func BenchmarkParentChildLog(b *testing.B) {
	f, err := os.Create("dump-parent-child-log.json")
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

func BenchmarkParent(b *testing.B) {
	f, err := os.Create("dump-parent.json")
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

		parent.Finish()
	}
	b.StopTimer()
}

func BenchmarkChild(b *testing.B) {
	f, err := os.Create("dump-child.json")
	if err != nil {
		return
	}
	t := NewWithOptions(Options{Writer: f})
	b.ResetTimer()
	parent := t.StartSpan("parent",
		SpanKindServer(),
		Component("component"),
		PeerHostname("hostname"),
		PeerHostIPv6("ip"),
		HTTPMethod("method"),
		HTTPUrl("https://some.url.outthere.com"),
	)

	for i := 0; i < b.N; i++ {
		child := t.StartSpan("child",
			opentracing.ChildOf(parent.Context()),
			SpanKindServer(),
			Component("child-component"),
			PeerHostname("hostname"),
			PeerHostIPv6("ip"),
			HTTPMethod("method"),
			HTTPUrl("https://some.url.outthere.com"),
		)

		child.Finish()
	}
	parent.Finish()

	b.StopTimer()
}

func BenchmarkLog(b *testing.B) {
	f, err := os.Create("dump-log.json")
	if err != nil {
		return
	}
	t := NewWithOptions(Options{Writer: f})
	b.ResetTimer()
	parent := t.StartSpan("parent",
		SpanKindServer(),
		Component("component"),
		PeerHostname("hostname"),
		PeerHostIPv6("ip"),
		HTTPMethod("method"),
		HTTPUrl("https://some.url.outthere.com"),
	)
	for i := 0; i < b.N; i++ {
		parent.LogFields(log.String("event", "event"))
	}
	parent.Finish()

	b.StopTimer()
}
