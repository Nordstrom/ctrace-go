package ctrace_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	ctrace "github.com/Nordstrom/ctrace-go"
	"github.com/Nordstrom/ctrace-go/ext"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opentracing "github.com/opentracing/opentracing-go"
)

var _ = Describe("Acceptance", func() {

	var (
		buf    bytes.Buffer
		tracer opentracing.Tracer
	)

	BeforeEach(func() {
		buf.Reset()
		tracer = ctrace.Init(ctrace.TracerOptions{Writer: &buf, MultiEvent: true})
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
				ext.SpanKindClient(),
				ext.Component("component"),
				ext.PeerHostname("hostname"),
				ext.PeerHostIPv6("ip"),
				ext.HTTPMethod("method"),
				ext.HTTPUrl("https://some.url.outthere.com"),
			)

			child := tracer.StartSpan("child",
				opentracing.ChildOf(parent.Context()),
				ext.SpanKindServer(),
				ext.Component("child-component"),
				ext.PeerService("service"),
				ext.PeerPort(80),
				ext.PeerHostname("hostname"),
				ext.PeerHostIPv4(123),
				ext.PeerHostIPv6("ip"),
				ext.HTTPMethod("method"),
				ext.HTTPUrl("https://some.url.outthere.com"),
			)

			child.SetTag(ext.HTTPStatusCodeKey, 200)
			child.Finish()

			parent.SetTag(ext.HTTPStatusCodeKey, 200)
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

			logs := out["logs"].([]interface{})
			log := (logs[0]).(map[string]interface{})
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

			logs := out["logs"].([]interface{})
			log := (logs[0]).(map[string]interface{})
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

			logs := out["logs"].([]interface{})
			log := (logs[0]).(map[string]interface{})
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

			logs := out["logs"].([]interface{})
			log := (logs[0]).(map[string]interface{})
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", timestamp))
			Ω(int64(log["timestamp"].(float64))).Should(BeNumerically(">=", int64(out["start"].(float64))))
			Ω(log["event"]).Should(Equal("Finish-Span"))
		})
	})
})
