package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
)

var _ = Describe("TracedHandler", func() {
	var (
		start time.Time
		mux   *http.ServeMux
		tr    *mocktracer.MockTracer
		srv   *httptest.Server
	)

	BeforeEach(func() {
		start = time.Now()
		mux = http.NewServeMux()
		tr = &mocktracer.MockTracer{}
		opentracing.InitGlobalTracer(tr)
	})

	AfterEach(func() {
		srv.Close()
	})

	Context("for ServeMux or ListenAndServe", func() {
		BeforeEach(func() {
			mux.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {})
			mux.HandleFunc("/test/error", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
				w.Write([]byte("There was an error"))
			})

			th := TracedHandler(mux)
			srv = httptest.NewServer(th)
		})

		It("records success correctly", func() {
			_, err := http.Get(srv.URL + "/test/foo")

			Ω(err).ShouldNot(HaveOccurred())

			spans := tr.FinishedSpans()
			Ω(spans).Should(HaveLen(1))
			sp := spans[0]

			Ω(sp.OperationName).Should(Equal("GET:/test/"))
			Ω(sp.ParentID).Should(BeZero())
			Ω(sp.SpanContext.TraceID).ShouldNot(BeZero())
			Ω(sp.SpanContext.SpanID).ShouldNot(BeZero())
			Ω(sp.StartTime).Should(BeTemporally(">=", start))
			Ω(sp.FinishTime).Should(BeTemporally(">=", sp.StartTime))

			Ω(sp.Tag("span.kind")).Should(Equal("server"))
			Ω(sp.Tag("component")).Should(Equal("ctrace.TracedHandler"))
			Ω(sp.Tag("http.url")).Should(Equal("/test/foo"))
			Ω(sp.Tag("http.method")).Should(Equal("GET"))
			Ω(sp.Tag("http.remote_addr")).ShouldNot(BeEmpty())
			Ω(sp.Tag("http.user_agent")).Should(Equal("Go-http-client/1.1"))
			Ω(sp.Tag("http.status_code")).Should(Equal(200))
		})

		It("records error correctly", func() {
			res, err := http.Get(srv.URL + "/test/error")

			Ω(err).ShouldNot(HaveOccurred())
			body, err := ioutil.ReadAll(res.Body)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(body)).Should(Equal("There was an error"))
			Ω(res.StatusCode).Should(Equal(400))

			spans := tr.FinishedSpans()
			Ω(spans).Should(HaveLen(1))
			sp := spans[0]

			Ω(sp.OperationName).Should(Equal("GET:/test/error"))
			Ω(sp.Tag("error")).Should(Equal(true))
			Ω(sp.Tag("http.status_code")).Should(Equal(400))

			logs := sp.Logs()
			Ω(logs).Should(HaveLen(0))
			// TODO: May add this back after decision on response body
			// lg := logs[0]
			// Ω(lg.Fields).Should(HaveLen(3))
			// Ω(lg.Timestamp).Should(BeTemporally(">=", sp.StartTime))
			// Ω(lg.Timestamp).Should(BeTemporally("<=", sp.FinishTime))
			// Ω(lg.Fields[0].Key).Should(Equal("event"))
			// Ω(lg.Fields[0].ValueString).Should(Equal("error"))
			// Ω(lg.Fields[1].Key).Should(Equal("error.kind"))
			// Ω(lg.Fields[1].ValueString).Should(Equal("http-server"))
			// Ω(lg.Fields[2].Key).Should(Equal("http.response.body"))
			// Ω(lg.Fields[2].ValueString).Should(Equal("There was an error"))
		})
	})

	Context("for Handle", func() {
		It("records default OperationName", func() {
			mux.Handle(
				"/test/",
				TracedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			)

			srv = httptest.NewServer(mux)
			http.Get(srv.URL + "/test/foo")
			Ω(tr.FinishedSpans()[0].OperationName).Should(Equal("GET:/test/foo"))
		})
	})

	Context("for HandleFunc", func() {
		It("records OperationName from OperationNameFunc", func() {
			mux.HandleFunc(
				"/test/",
				TracedHandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {},
					OperationNameFunc(func(r *http.Request) string {
						return r.Method + ":OVERRIDE"
					}),
				),
			)

			srv = httptest.NewServer(mux)
			http.Get(srv.URL + "/test/foo")
			Ω(tr.FinishedSpans()[0].OperationName).Should(Equal("GET:OVERRIDE"))
		})

		It("records OperationName from OperationName", func() {
			mux.HandleFunc(
				"/test/",
				TracedHandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {},
					OperationName("OP-OVERRIDE"),
				),
			)

			srv = httptest.NewServer(mux)
			http.Get(srv.URL + "/test/foo")
			Ω(tr.FinishedSpans()[0].OperationName).Should(Equal("OP-OVERRIDE"))
		})
	})
})
