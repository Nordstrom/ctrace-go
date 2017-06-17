package ctrace_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	opentracing "github.com/opentracing/opentracing-go"

	ctrace "github.com/Nordstrom/ctrace-go"
	"github.com/Nordstrom/ctrace-go/core"
	"github.com/Nordstrom/ctrace-go/ext"
)

var _ = Describe("TracedAPIGwLambdaProxyHandler", func() {
	var (
		start         time.Time
		buf           core.Buffer
		handler       ctrace.TracedLambdaFunction
		tracedHandler ctrace.LambdaFunction
		span          core.SpanModel
	)

	BeforeEach(func() {
		start = time.Now()
		buf.Reset()
		ctrace.Init(ctrace.TracerOptions{Writer: &buf})
		handler = func(
			ctx context.Context,
			evt *apigatewayproxyevt.Event,
			lambdaCtx *runtime.Context,
		) (interface{}, error) {
			if evt.Path == "/ok" {
				ctrace.LogInfo(ctx, "OK")
				return "SUCCESS", nil
			} else if evt.Path == "/error" {
				ctrace.LogErrorMessage(ctx, evt.Path)
				return nil, errors.New("there was an error")
			}

			return "DONE", nil
		}

		interceptor := func(
			evt *apigatewayproxyevt.Event,
			lambdaCtx *runtime.Context,
		) ctrace.SpanConfig {
			if evt.Path == "/intercept" {
				return ctrace.ConfigSpan(
					"newopname",
					opentracing.Tag{Key: "mytag", Value: "myval"},
					ext.HTTPMethod("POST"),
				)
			}
			return ctrace.SpanConfig{}
		}

		tracedHandler = ctrace.TracedAPIGwLambdaProxyHandler(handler, interceptor)
	})

	Context("with SUCCESS", func() {
		JustBeforeEach(func() {
			tracedHandler(
				&apigatewayproxyevt.Event{
					Path: "/ok",
					Headers: map[string]string{
						"User-Agent":  "myuseragent",
						"Remote-addr": "remoteaddr",
					},
					HTTPMethod: "GET",
				},
				&runtime.Context{
					FunctionName: "my-func",
				},
			)
			span = buf.Spans()[0]
		})

		It("records core attributes", func() {
			Expect(span.ParentID).To(BeEmpty())
			Expect(span.Operation).To(Equal("my-func"))
		})

		It("records tags", func() {
			tags := span.Tags
			Expect(tags["component"]).To(Equal("ctrace.TracedAPIGwLambdaProxyHandler"))
			Expect(tags["span.kind"]).To(Equal("server"))
			Expect(tags["http.status_code"]).To(Equal(float64(200)))
			Expect(tags["http.method"]).To(Equal("GET"))
			Expect(tags["http.url"]).To(Equal("/ok"))
			Expect(tags["http.user_agent"]).To(Equal("myuseragent"))
			Expect(tags["http.remote_addr"]).To(Equal("remoteaddr"))
		})

		It("records logs", func() {
			logs := span.Logs
			Expect(logs[1]["timestamp"]).To(BeNumerically(">=", start.UnixNano()/1e3))
			Expect(logs[1]["event"]).To(Equal("OK"))
		})
	})

	Context("with Interceptor", func() {
		JustBeforeEach(func() {
			tracedHandler(
				&apigatewayproxyevt.Event{
					Path:       "/intercept",
					HTTPMethod: "GET",
				},
				&runtime.Context{
					FunctionName: "my-func",
				},
			)
			span = buf.Spans()[0]
		})

		It("records operation name", func() {
			Expect(span.Operation).To(Equal("newopname"))
		})

		It("adds new tag", func() {
			Expect(span.Tags["mytag"]).To(Equal("myval"))
		})

		It("replaces existing tag", func() {
			Expect(span.Tags["http.method"]).To(Equal("POST"))
		})
	})

	Context("with ERROR", func() {
		JustBeforeEach(func() {
			tracedHandler(
				&apigatewayproxyevt.Event{
					Path: "/error",
					Headers: map[string]string{
						"User-Agent":  "myuseragent",
						"Remote-addr": "remoteaddr",
					},
					HTTPMethod: "GET",
				},
				&runtime.Context{
					FunctionName: "my-func",
				},
			)
			span = buf.Spans()[0]
		})

		It("records core attributes", func() {
			Expect(span.ParentID).To(BeEmpty())
			Expect(span.Operation).To(Equal("my-func"))
		})

		It("records tags", func() {
			tags := span.Tags
			fmt.Println(tags)
			Expect(tags["component"]).To(Equal("ctrace.TracedAPIGwLambdaProxyHandler"))
			Expect(tags["span.kind"]).To(Equal("server"))
			Expect(tags["http.status_code"]).To(Equal(float64(500)))
			Expect(tags["http.method"]).To(Equal("GET"))
			Expect(tags["http.url"]).To(Equal("/error"))
			Expect(tags["http.user_agent"]).To(Equal("myuseragent"))
			Expect(tags["http.remote_addr"]).To(Equal("remoteaddr"))
			Expect(tags["error"]).To(Equal(true))
		})

		It("records logs", func() {
			logs := span.Logs
			fmt.Println(logs[1])
			Expect(logs[1]["timestamp"]).To(BeNumerically(">=", start.UnixNano()/1e3))
			Expect(logs[1]["event"]).To(Equal("error"))
			Expect(logs[1]["error.kind"]).To(Equal("message"))
			Expect(logs[1]["message"]).To(Equal("/error"))
			Expect(logs[2]["timestamp"]).To(BeNumerically(">=", start.UnixNano()/1e3))
			Expect(logs[2]["event"]).To(Equal("error"))
			Expect(logs[2]["error.kind"]).To(Equal("object"))
			Expect(logs[2]["error.object"]).To(Equal("there was an error"))
		})
	})
})
