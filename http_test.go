package ctrace_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"

	ctrace "github.com/Nordstrom/ctrace-go"
	"github.com/Nordstrom/ctrace-go/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
)

var _ = Describe("http", func() {
	Describe("TracedHttpClientTransport", func() {
		var (
			start  time.Time
			mux    *http.ServeMux
			srv    *httptest.Server
			buf    core.Buffer
			tr     core.Tracer
			top    opentracing.Span
			rawTop core.Span
			client *http.Client
			hdrs   http.Header
		)

		BeforeEach(func() {
			start = time.Now()
			buf.Reset()
			ctrace.Init(ctrace.TracerOptions{Writer: &buf})
			tr = ctrace.Global()
			mux = http.NewServeMux()
			mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
				hdrs = r.Header
				w.Write([]byte("OK"))
			})
			mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/ok", http.StatusTemporaryRedirect)
			})
			mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "failure", http.StatusInternalServerError)
			})
			srv = httptest.NewServer(mux)
			top = tr.StartSpan("top")
			rawTop = top.(core.Span)
			client = &http.Client{Transport: ctrace.TracedHTTPClientTransport(http.DefaultTransport)}
		})

		AfterEach(func() {
			srv.Close()
		})

		It("handles ok with top", func() {
			req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
			ctx := opentracing.ContextWithSpan(req.Context(), top)
			req = req.WithContext(ctx)
			res, err := client.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			body, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			Expect(string(body)).To(Equal("OK"))
			sp := buf.Spans()[0]

			Expect(sp.Operation).To(Equal("GET:/ok"))
			Expect(sp.ParentID).To(Equal(rawTop.RawContext().SpanID()))
			Expect(sp.TraceID).To(Equal(rawTop.RawContext().TraceID()))
			Expect(sp.SpanID).To(Not(BeEmpty()))
			d := (time.Now().Sub(start)).Nanoseconds() / 1000
			Expect(sp.Start).To(BeNumerically(">=", start.Nanosecond()/1000))
			Expect(sp.Finish).To(BeNumerically(">=", time.Now().Nanosecond()/1000))
			// Duration should be between d/2 and d where d is now - test-start
			Expect(sp.Duration).To(BeNumerically(">=", d/2))
			Expect(sp.Duration).To(BeNumerically("<=", d))

			tags := sp.Tags
			Expect(tags["span.kind"]).To(Equal("client"))
			Expect(tags["component"]).To(Equal("ctrace.TracedHttpClientTransport"))
			Expect(tags["http.url"]).To(Equal(srv.URL + "/ok"))
			Expect(tags["http.method"]).To(Equal("GET"))
			Expect(int(tags["http.status_code"].(float64))).To(Equal(200))

			hctx, _ := tr.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
			hc := hctx.(core.SpanContext)
			Expect(hc.TraceID()).To(Equal(sp.TraceID))
			Expect(hc.SpanID()).To(Equal(sp.SpanID))
		})
	})

	Describe("TracedHttpHandler", func() {
		var (
			start time.Time
			mux   *http.ServeMux
			tr    core.Tracer
			buf   core.Buffer
			srv   *httptest.Server
		)

		BeforeEach(func() {
			start = time.Now()
			mux = http.NewServeMux()
			tr = ctrace.Global()
			buf.Reset()
			ctrace.Init(ctrace.TracerOptions{Writer: &buf.Buffer})
		})

		Context("for ServeMux or ListenAndServe", func() {
			BeforeEach(func() {
				mux.HandleFunc("/test/error", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(400)
					w.Write([]byte("There was an error"))
				})
				mux.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})

				th := ctrace.TracedHTTPHandler(mux, nil)
				srv = httptest.NewServer(th)
			})

			AfterEach(func() {
				srv.Close()
			})

			It("records success correctly", func() {
				_, err := http.Get(srv.URL + "/test/foo")

				Expect(err).ShouldNot(HaveOccurred())

				sp := buf.Spans()[0]
				Expect(sp.Operation).To(Equal("GET:/test/"))
				Expect(sp.ParentID).To(BeEmpty())
				Expect(sp.TraceID).ToNot(BeEmpty())
				Expect(sp.SpanID).ToNot(BeEmpty())
				Expect(sp.Start).To(BeNumerically(">=", start.UnixNano()/1000))
				Expect(sp.Finish).To(BeNumerically(">=", sp.Start))
				Expect(sp.Duration).To(BeNumerically(">=", sp.Finish-sp.Start-2))
				Expect(sp.Duration).To(BeNumerically("<=", sp.Finish-sp.Start+2))

				tags := sp.Tags

				Expect(tags["span.kind"]).To(Equal("server"))
				Expect(tags["component"]).To(Equal("ctrace.TracedHttpHandler"))
				Expect(tags["http.url"]).To(Equal("/test/foo"))
				Expect(tags["http.method"]).To(Equal("GET"))
				Expect(tags["http.remote_addr"]).ShouldNot(BeEmpty())
				Expect(tags["http.user_agent"]).To(Equal("Go-http-client/1.1"))
				Expect(tags["http.status_code"]).To(Equal(float64(200)))
			})

			It("records error correctly", func() {
				res, err := http.Get(srv.URL + "/test/error")

				Expect(err).ShouldNot(HaveOccurred())
				body, err := ioutil.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(string(body)).To(Equal("There was an error"))
				Expect(res.StatusCode).To(Equal(400))

				sp := buf.Spans()[0]
				Expect(sp.Operation).To(Equal("GET:/test/error"))
				tags := sp.Tags
				Expect(tags["error"]).To(Equal(true))
				Expect(tags["http.status_code"]).To(Equal(float64(400)))
			})
		})

		Context("for ServeMux or ListenAndServe, ignored Paths", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
				mux.HandleFunc("/test/error", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(400)
					w.Write([]byte("There was an error"))
				})
				mux.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
				reg, _ := regexp.Compile(`(\/v1\/health)`)
				th := ctrace.TracedHTTPHandler(mux, reg)
				srv = httptest.NewServer(th)
			})

			AfterEach(func() {
				srv.Close()
			})

			It("records success correctly", func() {
				_, err := http.Get(srv.URL + "/test/foo")

				Expect(err).ShouldNot(HaveOccurred())

				sp := buf.Spans()[0]
				Expect(sp.Operation).To(Equal("GET:/test/"))
				Expect(sp.ParentID).To(BeEmpty())
				Expect(sp.TraceID).ToNot(BeEmpty())
				Expect(sp.SpanID).ToNot(BeEmpty())
				Expect(sp.Start).To(BeNumerically(">=", start.UnixNano()/1000))
				Expect(sp.Finish).To(BeNumerically(">=", sp.Start))
				Expect(sp.Duration).To(BeNumerically(">=", sp.Finish-sp.Start-2))
				Expect(sp.Duration).To(BeNumerically("<=", sp.Finish-sp.Start+2))

				tags := sp.Tags

				Expect(tags["span.kind"]).To(Equal("server"))
				Expect(tags["component"]).To(Equal("ctrace.TracedHttpHandler"))
				Expect(tags["http.url"]).To(Equal("/test/foo"))
				Expect(tags["http.method"]).To(Equal("GET"))
				Expect(tags["http.remote_addr"]).ShouldNot(BeEmpty())
				Expect(tags["http.user_agent"]).To(Equal("Go-http-client/1.1"))
				Expect(tags["http.status_code"]).To(Equal(float64(200)))
			})

			It("ignores health endpoints", func() {
				_, err := http.Get(srv.URL + "/v1/health")

				Expect(err).ShouldNot(HaveOccurred())

				sp := buf.Spans()

				Expect(len(sp)).To(Equal(0))
			})

			It("records error correctly", func() {
				res, err := http.Get(srv.URL + "/test/error")

				Expect(err).ShouldNot(HaveOccurred())
				body, err := ioutil.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(string(body)).To(Equal("There was an error"))
				Expect(res.StatusCode).To(Equal(400))

				sp := buf.Spans()[0]
				Expect(sp.Operation).To(Equal("GET:/test/error"))
				tags := sp.Tags
				Expect(tags["error"]).To(Equal(true))
				Expect(tags["http.status_code"]).To(Equal(float64(400)))
			})
		})

		Context("for Handle", func() {
			It("records default OperationName", func() {
				mux.Handle(
					"/test/",
					ctrace.TracedHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(200)
					}), nil),
				)

				srv = httptest.NewServer(mux)
				defer srv.Close()
				http.Get(srv.URL + "/test/foo")
				Expect(buf.Spans()[0].Operation).To(Equal("GET:/test/foo"))
			})
		})

		Context("for HandleFunc", func() {
			It("records OperationName from Interceptor", func() {
				mux.HandleFunc(
					"/test/",
					ctrace.TracedHTTPHandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(200)
						},
						func(r *http.Request) ctrace.SpanConfig {
							return ctrace.SpanConfig{OperationName: r.Method + ":OVERRIDE"}
						},
					),
				)

				srv = httptest.NewServer(mux)
				defer srv.Close()
				http.Get(srv.URL + "/test/foo")
				Expect(buf.Spans()[0].Operation).To(Equal("GET:OVERRIDE"))
			})
		})
	})
})
