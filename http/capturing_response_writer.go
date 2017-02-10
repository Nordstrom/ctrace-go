package http

import "net/http"

// CapturingResponseWriter interface extends the http.ResponseWriter to capture
// whats written so that it can be included in the trace.
type CapturingResponseWriter interface {
	http.ResponseWriter
	StatusCode() int
	ResponseBody() []byte
}

// NewCapturingResponseWriter creates a new CapturingResponseWriter using current
// headers as a starting point.
func NewCapturingResponseWriter(headers http.Header) CapturingResponseWriter {
	return &capturingResponseWriter{
		headers:      headers,
		statusCode:   http.StatusOK,
		responseBody: []byte{},
	}
}

type capturingResponseWriter struct {
	statusCode   int
	responseBody []byte
	headers      http.Header
}

func (w *capturingResponseWriter) Header() http.Header {
	return w.headers
}

func (w *capturingResponseWriter) Write(data []byte) (int, error) {
	w.responseBody = make([]byte, len(data))
	copy(w.responseBody, data)
	return len(data), nil
}

func (w *capturingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *capturingResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *capturingResponseWriter) ResponseBody() []byte {
	return w.responseBody
}
