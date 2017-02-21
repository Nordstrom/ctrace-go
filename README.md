# ctrace-go
[![Build Status](https://travis-ci.org/Nordstrom/ctrace-go.svg?branch=new)](https://travis-ci.org/Nordstrom/ctrace-go)

Canonical OpenTracing for GoLang

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
Add instrumentation to the operations you want to track. This is composed primarily of using "spans" around operations of interest and adding log statements to capture useful data relevant to those operations.

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
    ...
}
```

### Creating a Span given an existing Go context.Context
If you use `context.Context` in your application, OpenTracing's Go library will happily rely on it for Span propagation. To start a new (blocking child) `Span`, you can use `StartSpanFromContext`.

```go
func xyz(ctx context.Context, ...) {
    ...
    span, ctx := opentracing.StartSpanFromContext(ctx, "operation_name")
    defer span.Finish()
    span.LogFields(
        log.String("event", "soft error"),
        log.String("type", "cache timeout"),
        log.Int("waited.millis", 1500))
    ...
}
```

### Starting an empty trace by creating a "root span"
It's always possible to create a "root" `Span` with no parent or other causal reference.

```go
func xyz() {
   ...
   sp := opentracing.StartSpan("operation_name")
   defer sp.Finish()
   ...
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
	span := opentracing.SpanFromContext(r.Context())
  span.LogFields(
    log.String("event", "handling-demo"),
    log.String("otherdata", "someotherdata"),
  )
  ...
}

func main() {
  ...

  http.HandleFunc("/demo", chttp.TracedHandlerFunc(handleDemo))

  ...

  http.ListenAndServe(":80", nil)
}
```

### Instrument Outgoing HTTP Requests
To automatically instrument outgoing HTTP Requests use the ctrace http.Transport.

```go
var httpClient = &http.Client{
	Transport: chttp.NewTransporter("http-client", &http.Transport{}),
}

...
req, err := http.NewRequest("GET", "http://some-service.com/demo", nil)
if err != nil {
  panic(err)
}
resp, err := httpClient.Do(req.WithContext(r.Context()))

```
