package http

import "net/http"

type httpOptions struct {
	opNameFunc func(r *http.Request) string
}

// Option controls the behavior of the ctrace http middleware
type Option func(*httpOptions)

// OperationNameFunc returns a Option that uses given function f to
// generate operation name for each span.
func OperationNameFunc(f func(r *http.Request) string) Option {
	return func(options *httpOptions) {
		options.opNameFunc = f
	}
}

// OperationName returns a Option that uses given opName as operation name
// for each span.
func OperationName(opName string) Option {
	return func(options *httpOptions) {
		options.opNameFunc = func(r *http.Request) string {
			return opName
		}
	}
}
