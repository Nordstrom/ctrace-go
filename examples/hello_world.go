package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	ctrace "github.com/Nordstrom/ctrace-go"
	chttp "github.com/Nordstrom/ctrace-go/http"
	"github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
)

var httpClient = &http.Client{
	Transport: chttp.NewTracedTransport(&http.Transport{}),
}

func gateway(w http.ResponseWriter, r *http.Request) {
	api := r.URL.Query().Get("api")
	if api == "" {
		api = "ok"
	}
	req, err := http.NewRequest("GET", "http://localhost:8004/"+api+"?"+r.URL.Query().Encode(), nil)
	if err != nil {
		panic(err)
	}

	span := opentracing.SpanFromContext(r.Context())
	span.SetBaggageItem("origin", r.RemoteAddr)

	resp, err := httpClient.Do(req.WithContext(r.Context()))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func hello(ctx context.Context, region string) string {
	fmt.Printf("ctx=%+v\n", ctx)
	span := opentracing.SpanFromContext(ctx)

	msg := fmt.Sprintf("Hello %v!", region)
	span.LogFields(log.Event("generate-msg"), log.Message(msg))
	return msg
}

func ok(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("headers=%+v\n", r.Header)
	region := r.URL.Query().Get("region")
	msg := hello(r.Context(), region)
	w.Write([]byte(msg))
}

func err(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	w.Write([]byte("There was an ERROR!"))
}

func main() {
	opentracing.InitGlobalTracer(
		ctrace.New(),
	)

	http.HandleFunc("/gateway", gateway)
	http.HandleFunc("/ok", ok)
	http.HandleFunc("/err", err)

	fmt.Println("ctrace-go example server starting...")
	http.ListenAndServe(":8004", chttp.TracedHandler(http.DefaultServeMux))
}
