# ctrace-go
[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![OpenTracing 1.0 Enabled][ot-img]][ot-url]

Canonical OpenTracing for Go

## Why
[OpenTracing](http://opentracing.io) is a young specification and for most (if not all) SDK implementations, output format and wire protocol are specific to the backend platform implementation.  ctrace attempts to decouple the format and wire protocol from the backend tracer implementation.

## What
ctrace specifies a canonical format for trace logs.  By default the logs are output to stdout but you can configure them to go to any WritableStream.

## Required Reading
To fully understand this platform API, it's helpful to be familiar with the [OpenTracing project](http://opentracing.io) project, [terminology](http://opentracing.io/documentation/pages/spec.html), and [ctrace specification](https://github.com/Nordstrom/ctrace) more specifically.

## Install
Install via glide as follows:

```
$ glide get github.com/Nordstrom/ctrace-go
```

## Usage
Add instrumentation to the operations you want to track.  Most of this is done by middleware.

### Singleton Initialization
First initialize the Global Tracer Singleton as early as possible.

```go
import (
  opentracing "github.com/opentracing/opentracing-go"
  ctrace "github.com/Nordstrom/ctrace-go"
)

func main() {
    opentracing.InitGlobalTracer(
        // tracing impl specific:
        ctrace.New(...),
    )
    // ...
}
```

### Instrument Incoming HTTP Requests
To automatically instrument incoming HTTP Requests use the TracedHandlerFunc wrapper.

```go
import (
  opentracing "github.com/opentracing/opentracing-go"
  log "github.com/opentracing/opentracing-go/log"
  chttp "github.com/Nordstrom/ctrace-go/http"
)

func handleDemo(w http.ResponseWriter, r *http.Request) {
  // ...
  processDemo(r.Context())
  // ...
}

func main() {
  // ...

  http.HandleFunc("/demo", chttp.TracedHandlerFunc(handleDemo))

  // ...

  http.ListenAndServe(":80", nil)
}
```

### Log Events
To log events within the context of the current Span, use span.LogFields.

```go
func processDemo(ctx context.Context, data string) {
  span := opentracing.SpanFromContext(ctx)
  span.LogFields(
    log.String("event", "processing-demo"),
    log.String("data", data),
  )
  // ...
}
```

### Instrument Outgoing HTTP Requests
To automatically instrument outgoing HTTP Requests use the ctrace http.Transport.

```go
var httpClient = &http.Client{
  Transport: chttp.NewTransporter("http-client", &http.Transport{}),
}

func makeOutgoing(ctx context.Context) {
  req, err := http.NewRequest("GET", "http://some-service.com/demo", nil)
  if err != nil {
    panic(err)
  }
  resp, err := httpClient.Do(req.WithContext(ctx))
}
```

## Advanced Usage
If middleware does not fully meet your needs, you can manually instrument spans
operations of interest and adding log statements to capture useful data relevant
to those operations.

### Creating a Span given an existing Go context.Context
If you use `context.Context` in your application, OpenTracing's Go library will happily rely on it for Span propagation. To start a new (blocking child) `Span`, you can use `StartSpanFromContext`.

```go
func xyz(ctx context.Context, ...) {
    // ...
    span, ctx := opentracing.StartSpanFromContext(ctx, "operation_name")
    defer span.Finish()
    span.LogFields(
        log.String("event", "soft error"),
        log.String("type", "cache timeout"),
        log.Int("waited.millis", 1500))
    // ...
}
```

### Starting an empty trace by creating a "root span"
It's always possible to create a "root" `Span` with no parent or other causal reference.

```go
func xyz() {
   // ...
   sp := opentracing.StartSpan("operation_name")
   defer sp.Finish()
   // ...
}
```

## Best Practices
The following are recommended practices for using opentracing and ctrace-go within a
GoLang project.

### Context Propagation
Require a context.Context argument as the first parameter of every(a) API call

```go
func (h handler) HandleRequest(ctx context.Context, r MyRequest) error {
  // ...
}
```

Rule of thumb: because contexts are always request-scopeed, never hold a reference
to them on a struct.  Always pass as a function parameter.

a. Obviously not every function in your codebase.  You'll get a feel for the balance
when you start writing context-aware code.

### Custom Context Types
At some point, you will be tempted to invent your own "custom" context type.
For example to provide convenience methods, since Value() takes an interface{}
key and returns an interface{}.

You will regret it. Use extractor functions instead.

## Contributing
Please see the [contributing guide](CONTRIBUTING.md) for details on contributing to ctrace-go.

## License
[Apache License 2.0](LICENSE)

[doc-img]: https://godoc.org/github.com/Nordstrom/ctrace-go?status.svg
[doc]: https://godoc.org/github.com/Nordstrom/ctrace-go
[ci-img]: https://travis-ci.org/Nordstrom/ctrace-go.svg
[ci]: https://travis-ci.org/Nordstrom/ctrace-go
[cov-img]: https://coveralls.io/repos/github/Nordstrom/ctrace-go/badge.svg
[cov]: https://coveralls.io/github/Nordstrom/ctrace-go
[ot-img]: https://img.shields.io/badge/OpenTracing--1.0-enabled-blue.svg
[ot-url]: http://opentracing.io
