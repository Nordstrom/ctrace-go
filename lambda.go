package ctrace

import (
	"context"

	"github.com/Nordstrom/ctrace-go/core"
	"github.com/Nordstrom/ctrace-go/ext"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
	opentracing "github.com/opentracing/opentracing-go"
)

// LambdaFunction is the defined function for Lambda handlers
type LambdaFunction func(
	evt *apigatewayproxyevt.Event,
	lambdaCtx *runtime.Context,
) (interface{}, error)

// TracedLambdaFunction is the defined function for traced Lambda handlers
// it adds context.Context as the first parameter.
type TracedLambdaFunction func(
	ctx context.Context,
	evt *apigatewayproxyevt.Event,
	lambdaCtx *runtime.Context,
) (interface{}, error)

// LambdaFunctionInterceptor is the defined function for intercepting the
// TracedApiGwLambdaProxyHandler calls for providing custom OperationName and/or
// custom Tags
type LambdaFunctionInterceptor func(evt *apigatewayproxyevt.Event, ctx *runtime.Context) SpanConfig

func optioniallyInterceptLambda(
	evt *apigatewayproxyevt.Event,
	ctx *runtime.Context,
	i ...LambdaFunctionInterceptor) SpanConfig {
	for _, f := range i {
		return f(evt, ctx)
	}
	return SpanConfig{}
}

// TracedAPIGwLambdaProxyHandler is a decorator (wrapper) that wraps the Lambda
// handler function for tracing.  It handles starting a span when the handler
// is called and finishing it upon completion.  To customize the OperationName
// or Tags pass in a LambdaFunctionInterceptor
func TracedAPIGwLambdaProxyHandler(
	fn TracedLambdaFunction,
	interceptor ...LambdaFunctionInterceptor,
) LambdaFunction {
	tfn := func(
		evt *apigatewayproxyevt.Event,
		lambdaCtx *runtime.Context,
	) (interface{}, error) {
		tracer := Global()
		parentCtx, _ := tracer.Extract(core.TextMap, core.TextMapCarrier(evt.Headers))
		config := optioniallyInterceptLambda(evt, lambdaCtx, interceptor...)

		var op string
		if config.OperationName != "" {
			op = config.OperationName
		} else {
			op = lambdaCtx.FunctionName
		}
		opts := []opentracing.StartSpanOption{
			ChildOf(parentCtx),
			ext.SpanKindServer(),
			ext.Component("ctrace.TracedAPIGwLambdaProxyHandler"),
			ext.HTTPRemoteAddr(httpRemoteAddr(evt.Headers)),
			ext.HTTPMethod(evt.HTTPMethod),
			ext.HTTPUrl(evt.Path),
			ext.HTTPUserAgent(httpUserAgent(evt.Headers)),
		}

		if len(config.Tags) > 0 {
			opts = append(opts, config.Tags...)
		}
		span := tracer.StartSpan(op, opts...)
		defer span.Finish()

		ctx := ContextWithSpan(context.Background(), span)
		rtn, err := fn(ctx, evt, lambdaCtx)

		if err != nil {
			span.SetTag(ext.ErrorKey, true)
			span.SetTag(ext.HTTPStatusCodeKey, 500)
			LogErrorObject(ctx, err)
		} else {
			span.SetTag(ext.HTTPStatusCodeKey, 200)
		}

		return rtn, err
	}

	return tfn
}
