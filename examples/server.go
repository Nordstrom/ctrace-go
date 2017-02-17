package main

import (
	"io/ioutil"
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
	chttp "github.com/Nordstrom/ctrace-go/http"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/opentracing/opentracing-go/log"
)

var httpClient = &http.Client{
	Transport: chttp.NewTransporter("http-client", &http.Transport{}),
}

func handleDemoGateway(w http.ResponseWriter, r *http.Request) {
	resp, err := httpClient.Get("http://localhost:8004/demo")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	w.Write(body)
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	span := opentracing.SpanFromContext(r.Context())
	span.LogFields(log.String("event", "handling-demo"))
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"hello":"world"}`))
}

func main() {
	opentracing.InitGlobalTracer(
		ctrace.New(),
	)

	http.HandleFunc("/gateway/demo", chttp.TracedHandlerFunc(handleDemoGateway))
	http.HandleFunc("/demo", chttp.TracedHandlerFunc(handleDemo))

	http.ListenAndServe(":8004", nil)
}
