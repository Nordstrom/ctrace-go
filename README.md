# ctrace-go
[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Report Card][rep-img]][rep] [![Coverage Status][cov-img]][cov] [![OpenTracing 1.0 Enabled][ot-img]][ot-url]

[Canonical OpenTracing](https://github.com/Nordstrom/ctrace) for Go

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

### Initialization
First import ctrace.

```go
import (
	ctrace "github.com/Nordstrom/ctrace-go"
)
```

### Instrumentation
The primary way to use ctrace-go is to hook in one or more of the auto-instrumentation
decorators.  These are the decorators we currently support.

- **TracedHttpHandler** - this wraps the default serve mux to trace **Incoming HTTP Requests**.
- **TracedHttpClientTransport** - this wraps the http.Transport for tracing **Outgoing HTTP Requests**.
- **TracedAPIGwLambdaProxyHandler** - this wraps **AWS Gateway Lambda Proxy** handler using [AWS Lambda Go Shim](https://github.com/eawsy/aws-lambda-go-shim) to trace requests coming into an AWS Lambda from an API Gateway proxy event.

### TracedHttpHandler
To automatically instrument incoming HTTP Requests use the TracedHttpHandler.

```go
import (
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
)

func ok(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	msg := hello(r.Context(), region)
	w.Write([]byte(msg))
}

func main() {
	http.HandleFunc("/ok", ok)
	http.ListenAndServe(":8004", ctrace.TracedHTTPHandler(http.DefaultServeMux))
}

```

### TracedHTTPClientTransport
To automatically instrument outgoing HTTP Requests use the TracedHttpClientTransport.

```go
import (
	"context"
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
)

var httpClient = &http.Client{
	Transport: ctrace.TracedHTTPClientTransport(&http.Transport{}),
}

func send(ctx context.Context, method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req.WithContext(ctx))
}
```

### TracedAPIGwLambdaProxyHandler
To automatically instrument incoming API Gateway Lambda proxy requests use TracedAPIGwLambdaProxyHandler.

```go
import (
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
	ctrace "github.com/Nordstrom/ctrace-go"
)

handler := func(
	ctx context.Context,
	evt *apigatewayproxyevt.Event,
	lambdaCtx *runtime.Context,
) (interface{}, error) {
  // ...
}

var TracedHandler = ctrace.TracedAPIGwLambdaProxyHandler(handler)

```

### Log Events
To log events within the context of the current Span, use span.LogFields.

```go
import (
	ctrace "github.com/Nordstrom/ctrace-go"
	log "github.com/Nordstrom/ctrace-go/log"
)

func hello(ctx context.Context, region string) string {
	msg := fmt.Sprintf("Hello %v!", region)
	ctrace.LogInfo(ctx, "generate-msg", log.Message(message))
	return msg
}

```

## Advanced Usage
If middleware does not fully meet your needs, you can manually instrument spans
operations of interest and adding log statements to capture useful data relevant
to those operations.

```go
func main() {
	ctrace.Init(TracerOptions{
		MultiEvent: true,
		Writer:     fileWriter,
	})
}
```

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
[rep-img]: https://goreportcard.com/badge/github.com/Nordstrom/ctrace-go
[rep]: https://goreportcard.com/report/github.com/Nordstrom/ctrace-go
[cov-img]: https://coveralls.io/repos/github/Nordstrom/ctrace-go/badge.svg
[cov]: https://coveralls.io/github/Nordstrom/ctrace-go
[ot-img]: https://img.shields.io/badge/OpenTracing--1.0-enabled-blue.svg
[ot-url]: http://opentracing.io
