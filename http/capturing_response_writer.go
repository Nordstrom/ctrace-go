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
func NewCapturingResponseWriter(w http.ResponseWriter) CapturingResponseWriter {
	return &capturingResponseWriter{writer: w}
}

type capturingResponseWriter struct {
	writer       http.ResponseWriter
	statusCode   int
	responseBody []byte
}

func (w *capturingResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *capturingResponseWriter) Write(data []byte) (int, error) {
	w.responseBody = make([]byte, len(data))
	copy(w.responseBody, data)
	return w.writer.Write(data)
}

func (w *capturingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.writer.WriteHeader(code)
}

func (w *capturingResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *capturingResponseWriter) ResponseBody() []byte {
	return w.responseBody
}
