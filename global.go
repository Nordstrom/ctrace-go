package ctrace

import (
	"github.com/Nordstrom/ctrace-go/core"
	opentracing "github.com/opentracing/opentracing-go"
	godebug "github.com/tj/go-debug"
)

// TracerOptions allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.  This is an alias
// of core.TracerOptions.
type TracerOptions core.TracerOptions

var (
	debug = godebug.Debug("ctrace")
)

func init() {
	debug("Initializing ctrace...") // start with empty line for testing
	Init(TracerOptions{})
}

// Init initializes the global Tracer returned by Global().
func Init(opts TracerOptions) core.Tracer {
	opentracing.SetGlobalTracer(core.NewWithOptions(
		core.TracerOptions{
			MultiEvent:  opts.MultiEvent,
			Writer:      opts.Writer,
			ServiceName: opts.ServiceName,
		}))

	return Global()
}

// Global returns the global Tracer
func Global() core.Tracer {
	return opentracing.GlobalTracer().(core.Tracer)
}
