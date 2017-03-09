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

var _ = Describe("Transport", func() {
	var (
		start  time.Time
		mux    *http.ServeMux
		tr     *mocktracer.MockTracer
		srv    *httptest.Server
		top    opentracing.Span
		rawTop *mocktracer.MockSpan
		client *http.Client
		hdrs   http.Header
	)

	BeforeEach(func() {
		start = time.Now()
		mux = http.NewServeMux()
		tr = mocktracer.New()
		opentracing.InitGlobalTracer(tr)
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
		rawTop = top.(*mocktracer.MockSpan)
		client = &http.Client{Transport: NewTracedTransport(http.DefaultTransport)}
	})

	AfterEach(func() {
		srv.Close()
	})

	It("handles ok with top", func() {
		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
		ctx := opentracing.ContextWithSpan(req.Context(), top)
		req = req.WithContext(ctx)
		res, err := client.Do(req)
		Ω(err).ShouldNot(HaveOccurred())
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		Ω(string(body)).Should(Equal("OK"))
		spans := tr.FinishedSpans()
		Ω(spans).Should(HaveLen(1))

		sp := spans[0]
		Ω(sp.OperationName).Should(Equal("GET:/ok"))
		Ω(sp.ParentID).Should(Equal(rawTop.SpanContext.SpanID))
		Ω(sp.SpanContext.TraceID).Should(Equal(rawTop.SpanContext.TraceID))
		Ω(sp.SpanContext.SpanID).ShouldNot(BeZero())
		Ω(sp.StartTime).Should(BeTemporally(">=", start))
		Ω(sp.FinishTime).Should(BeTemporally(">=", sp.StartTime))

		Ω(sp.Tag("span.kind")).Should(Equal("client"))
		Ω(sp.Tag("component")).Should(Equal("ctrace.TracedTransport"))
		Ω(sp.Tag("http.url")).Should(Equal(srv.URL + "/ok"))
		Ω(sp.Tag("http.method")).Should(Equal("GET"))
		Ω(sp.Tag("http.status_code")).Should(Equal(200))

		hctx, _ := tr.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs))
		hc := hctx.(mocktracer.MockSpanContext)
		Ω(hc.TraceID).Should(Equal(sp.SpanContext.TraceID))
		Ω(hc.SpanID).Should(Equal(sp.SpanContext.SpanID))
	})
})
