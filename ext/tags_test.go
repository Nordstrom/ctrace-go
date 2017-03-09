package ext

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ot "github.com/opentracing/opentracing-go"
)

var _ = Describe("Tags", func() {
	It("SpanKindClient", func() {
		Ω(SpanKindClient()).Should(Equal(ot.Tag{Key: "span.kind", Value: "client"}))
	})

	It("SpanKindServer", func() {
		Ω(SpanKindServer()).Should(Equal(ot.Tag{Key: "span.kind", Value: "server"}))
	})

	It("Component", func() {
		Ω(Component("comp")).Should(Equal(ot.Tag{Key: "component", Value: "comp"}))
	})

	It("DbInstance", func() {
		Ω(DbInstance("dbinst")).Should(Equal(ot.Tag{Key: "db.instance", Value: "dbinst"}))
	})

	It("DbStatement", func() {
		Ω(DbStatement("dbstat")).Should(Equal(ot.Tag{Key: "db.statement", Value: "dbstat"}))
	})

	It("DbType", func() {
		Ω(DbType("dbtype")).Should(Equal(ot.Tag{Key: "db.type", Value: "dbtype"}))
	})

	It("DbUser", func() {
		Ω(DbUser("dbuser")).Should(Equal(ot.Tag{Key: "db.user", Value: "dbuser"}))
	})

	It("Error", func() {
		Ω(Error(true)).Should(Equal(ot.Tag{Key: "error", Value: true}))
	})

	It("HTTPMethod", func() {
		Ω(HTTPMethod("hmeth")).Should(Equal(ot.Tag{Key: "http.method", Value: "hmeth"}))
	})

	It("HTTPUrl", func() {
		Ω(HTTPUrl("url")).Should(Equal(ot.Tag{Key: "http.url", Value: "url"}))
	})

	It("HTTPRemoteAddr", func() {
		Ω(HTTPRemoteAddr("remaddr")).Should(Equal(ot.Tag{Key: "http.remote_addr", Value: "remaddr"}))
	})

	It("HTTPStatusCode", func() {
		Ω(HTTPStatusCode(200)).Should(Equal(ot.Tag{Key: "http.status_code", Value: 200}))
	})

	It("HTTPUserAgent", func() {
		Ω(HTTPUserAgent("uagent")).Should(Equal(ot.Tag{Key: "http.user_agent", Value: "uagent"}))
	})

	It("PeerHostIPv4", func() {
		Ω(PeerHostIPv4(123)).Should(Equal(ot.Tag{Key: "peer.ipv4", Value: uint32(123)}))
	})

	It("PeerHostIPv6", func() {
		Ω(PeerHostIPv6("ip6")).Should(Equal(ot.Tag{Key: "peer.ipv6", Value: "ip6"}))
	})

	It("PeerHostname", func() {
		Ω(PeerHostname("host")).Should(Equal(ot.Tag{Key: "peer.hostname", Value: "host"}))
	})

	It("PeerPort", func() {
		Ω(PeerPort(123)).Should(Equal(ot.Tag{Key: "peer.port", Value: uint16(123)}))
	})
})
