package ctrace

import (
	"os"
	"testing"

	"github.com/Nordstrom/ctrace-go/ext"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func BenchmarkParentChildLog(b *testing.B) {
	f, err := os.Create("dump-parent-child-log.json")
	if err != nil {
		return
	}
	t := NewWithOptions(TracerOptions{Writer: f})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parent := t.StartSpan("parent",
			ext.SpanKindServer(),
			ext.Component("component"),
			ext.PeerHostname("hostname"),
			ext.PeerHostIPv6("ip"),
			ext.HTTPMethod("method"),
			ext.HTTPUrl("https://some.url.outthere.com"),
		)

		child := t.StartSpan("child",
			opentracing.ChildOf(parent.Context()),
			ext.SpanKindServer(),
			ext.Component("child-component"),
			ext.PeerHostname("hostname"),
			ext.PeerHostIPv6("ip"),
			ext.HTTPMethod("method"),
			ext.HTTPUrl("https://some.url.outthere.com"),
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
	t := NewWithOptions(TracerOptions{Writer: f})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parent := t.StartSpan("parent",
			ext.SpanKindServer(),
			ext.Component("component"),
			ext.PeerHostname("hostname"),
			ext.PeerHostIPv6("ip"),
			ext.HTTPMethod("method"),
			ext.HTTPUrl("https://some.url.outthere.com"),
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
	t := NewWithOptions(TracerOptions{Writer: f})
	b.ResetTimer()
	parent := t.StartSpan("parent",
		ext.SpanKindServer(),
		ext.Component("component"),
		ext.PeerHostname("hostname"),
		ext.PeerHostIPv6("ip"),
		ext.HTTPMethod("method"),
		ext.HTTPUrl("https://some.url.outthere.com"),
	)

	for i := 0; i < b.N; i++ {
		child := t.StartSpan("child",
			opentracing.ChildOf(parent.Context()),
			ext.SpanKindServer(),
			ext.Component("child-component"),
			ext.PeerHostname("hostname"),
			ext.PeerHostIPv6("ip"),
			ext.HTTPMethod("method"),
			ext.HTTPUrl("https://some.url.outthere.com"),
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
	t := NewWithOptions(TracerOptions{Writer: f})
	b.ResetTimer()
	parent := t.StartSpan("parent",
		ext.SpanKindServer(),
		ext.Component("component"),
		ext.PeerHostname("hostname"),
		ext.PeerHostIPv6("ip"),
		ext.HTTPMethod("method"),
		ext.HTTPUrl("https://some.url.outthere.com"),
	)
	for i := 0; i < b.N; i++ {
		parent.LogFields(log.String("event", "event"))
	}
	parent.Finish()

	b.StopTimer()
}
