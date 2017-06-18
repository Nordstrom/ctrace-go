package ctrace

import opentracing "github.com/opentracing/opentracing-go"

// SpanConfig is used by middleware interceptors to return custom OperationName
// and Tags
//
type SpanConfig struct {
	// OperationName is the custom operation name decided by interceptor
	OperationName string

	// Tags are the custom start span options decided by interceptor.
	Tags []opentracing.StartSpanOption
}

// ConfigSpan function is used by middleware interceptors to construct a SpanConfig
// for customizing the starting of a span.  For example:
//
//     ctrace.TracedHttpHandler(
//       http.DefaultServeMux,
//       func (r *http.Request) SpanConfig {
//         return ConfigSpan(
//           "MyOperation:" + r.URL.String(),
//           ext.String("MyTag", "MyValue")
//         )
//       },
//     )
func ConfigSpan(
	operationName string,
	tags ...opentracing.StartSpanOption,
) SpanConfig {
	return SpanConfig{
		OperationName: operationName,
		Tags:          tags,
	}
}
