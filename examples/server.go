package main

import (
	"io/ioutil"
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
	chttp "github.com/Nordstrom/ctrace-go/http"
	opentracing "github.com/opentracing/opentracing-go"
)

func handleDemoGateway(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://localhost:8004/demo")
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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"hello":"world"}`))
}

func main() {
	opentracing.InitGlobalTracer(
		ctrace.New(),
	)

	http.HandleFunc("/gateway/demo", chttp.TracedHandlerFunc("demo-gateway", "GetDemo", handleDemoGateway))
	http.HandleFunc("/demo", chttp.TracedHandlerFunc("demo-service", "GetDemo", handleDemo))

	http.ListenAndServe(":8004", nil)
}
