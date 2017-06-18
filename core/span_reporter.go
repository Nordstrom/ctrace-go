package core

import (
	"fmt"
	"io"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
)

// SpanReporter reports the current state of a Span.  It is intended to reports
// Start-Span, Log, and Finish-Span events.
type SpanReporter interface {
	Report(opentracing.Span)
}

type spanReporter struct {
	io.Writer
	SpanEncoder
	sync.Mutex
}

// NewSpanReporter creates a new default SpanReporter.
func NewSpanReporter(w io.Writer, e SpanEncoder) SpanReporter {
	return &spanReporter{Writer: w, SpanEncoder: e}
}

func (r *spanReporter) Report(sp opentracing.Span) {
	bytes := r.Encode(sp)
	expectedBytes := len(bytes)

	r.Lock()
	defer r.Unlock()
	n, err := r.Write(bytes)

	if err != nil {
		fmt.Println(err)
		return
	}

	if expectedBytes != n {
		fmt.Printf("Expect %d bytes reported, but had %d instead\n", expectedBytes, n)
	}
}
