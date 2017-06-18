package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	ctrace "github.com/Nordstrom/ctrace-go"
	"github.com/Nordstrom/ctrace-go/log"
	opentracing "github.com/opentracing/opentracing-go"
)

var httpClient = &http.Client{
	Transport: ctrace.TracedHTTPClientTransport(&http.Transport{}),
}

func send(ctx context.Context, u string, r string) ([]byte, error) {
	req, err := http.NewRequest("GET", u+"?region="+r, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(req.URL.String())
	res, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func gateway(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	u, _ := url.QueryUnescape(q.Get("url"))
	region := q.Get("region")

	body, err := send(r.Context(), u, region)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("ERROR!"))
	} else {
		w.WriteHeader(200)
		w.Write(body)
	}
}

func ok(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	region, _ := url.QueryUnescape(q.Get("region"))
	msg := hello(r.Context(), region)
	w.Write([]byte(msg))
}

func hello(ctx context.Context, region string) string {
	span := opentracing.SpanFromContext(ctx)
	msg := fmt.Sprintf("Hello %v!", region)
	span.LogFields(log.Event("generate-msg"), log.Message(msg))

	return msg
}

func err(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fmt.Println("Hello Gateway...")

	http.HandleFunc("/gw", gateway)
	http.HandleFunc("/ok", ok)
	http.HandleFunc("/err", err)
	addr := ":80"
	fmt.Printf("Listening at %v ...\n", addr)

	e := http.ListenAndServe(addr, ctrace.TracedHTTPHandler(http.DefaultServeMux))
	if e != nil {
		panic(e)
	}
}
