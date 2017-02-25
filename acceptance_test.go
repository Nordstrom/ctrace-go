package ctrace_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	ctrace "github.com/Nordstrom/ctrace-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
)

func lines(buf bytes.Buffer) []string {
	return strings.Split(buf.String(), "\n")
}

var _ = Describe("Acceptance", func() {

	var (
		buf    bytes.Buffer
		tracer opentracing.Tracer
	)

	BeforeEach(func() {
		buf.Reset()
		tracer = ctrace.NewWithOptions(ctrace.TracerOptions{Writer: &buf})
	})

	Describe("Parent and Child with Standard Tags", func() {
		var (
			lines     []string
			out       map[string]interface{}
			timestamp int64
		)

		BeforeEach(func() {
			timestamp = time.Now().UnixNano() / 1e3
			parent := tracer.StartSpan("parent",
				ctrace.SpanKindClient(),
				ctrace.Component("component"),
				ctrace.PeerHostname("hostname"),
				ctrace.PeerHostIPv6("ip"),
				ctrace.HTTPMethod("method"),
				ctrace.HTTPUrl("https://some.url.outthere.com"),
			)

			child := tracer.StartSpan("child",
				opentracing.ChildOf(parent.Context()),
				ctrace.SpanKindServer(),
				ctrace.Component("child-component"),
				ctrace.PeerService("service"),
				ctrace.PeerPort(80),
				ctrace.PeerHostname("hostname"),
				ctrace.PeerHostIPv4(123),
				ctrace.PeerHostIPv6("ip"),
				ctrace.HTTPMethod("method"),
				ctrace.HTTPUrl("https://some.url.outthere.com"),
			)

			child.SetTag(ctrace.HTTPStatusCodeKey, 200)
			child.Finish()

			parent.SetTag(ctrace.HTTPStatusCodeKey, 200)
			parent.Finish()
			lines = strings.Split(buf.String(), "\n")
			out = make(map[string]interface{})
		})

		It("outputs parent Start-Span", func() {
			if err := json.Unmarshal([]byte(lines[0]), &out); err != nil {
				Fail("Cannot unmarshal JSON")
			}

			Ω(out["traceId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["spanId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["operation"]).Should(Equal("parent"))
			Ω(int64(out["start"].(float64))).Should(BeNumerically(">=", timestamp))

			Ω(out["tags"]).Should(Equal(
				map[string]interface{}{
					"component":     "component",
					"peer.hostname": "hostname",
					"peer.ipv6":     "ip",
					"http.method":   "method",
					"http.url":      "https://some.url.outthere.com",
					"span.kind":     "client"},
			))

			log := out["log"].(map[string]interface{})
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(log["event"]).Should(Equal("Start-Span"))
		})

		It("outputs child Start-Span", func() {
			if err := json.Unmarshal([]byte(lines[1]), &out); err != nil {
				Fail("Cannot unmarshal JSON")
			}

			Ω(out["traceId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["spanId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["operation"]).Should(Equal("child"))
			Ω(int64(out["start"].(float64))).Should(BeNumerically(">=", timestamp))

			Ω(out["tags"]).Should(Equal(
				map[string]interface{}{
					"component":     "child-component",
					"peer.hostname": "hostname",
					"peer.ipv6":     "ip",
					"peer.ipv4":     float64(123),
					"peer.port":     float64(80),
					"peer.service":  "service",
					"http.method":   "method",
					"http.url":      "https://some.url.outthere.com",
					"span.kind":     "server"},
			))

			log := out["log"].(map[string]interface{})
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(log["event"]).Should(Equal("Start-Span"))
		})

		It("outputs child Finish-Span", func() {
			if err := json.Unmarshal([]byte(lines[2]), &out); err != nil {
				Fail("Cannot unmarshal JSON")
			}

			Ω(out["traceId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["spanId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["operation"]).Should(Equal("child"))
			Ω(int64(out["start"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(int64(out["duration"].(float64))).Should(BeNumerically(">=", 0))

			Ω(out["tags"]).Should(Equal(
				map[string]interface{}{
					"component":        "child-component",
					"peer.hostname":    "hostname",
					"peer.ipv6":        "ip",
					"peer.ipv4":        float64(123),
					"peer.port":        float64(80),
					"peer.service":     "service",
					"http.method":      "method",
					"http.status_code": float64(200),
					"http.url":         "https://some.url.outthere.com",
					"span.kind":        "server"},
			))

			log := out["log"].(map[string]interface{})
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", int64(out["start"].(float64))))
			Ω(log["event"]).Should(Equal("Finish-Span"))
		})

		It("outputs parent Finish-Span", func() {
			if err := json.Unmarshal([]byte(lines[3]), &out); err != nil {
				Fail("Cannot unmarshal JSON")
			}

			Ω(out["traceId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["spanId"]).Should(MatchRegexp("[0-9a-f]{16}"))
			Ω(out["operation"]).Should(Equal("parent"))
			Ω(int64(out["start"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(int64(out["duration"].(float64))).Should(BeNumerically(">=", 0))

			Ω(out["tags"]).Should(Equal(
				map[string]interface{}{
					"component":        "component",
					"peer.hostname":    "hostname",
					"peer.ipv6":        "ip",
					"http.method":      "method",
					"http.url":         "https://some.url.outthere.com",
					"http.status_code": float64(200),
					"span.kind":        "client"},
			))

			log := out["log"].(map[string]interface{})
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", int64(out["start"].(float64))))
			Ω(log["event"]).Should(Equal("Finish-Span"))
		})
	})
})
