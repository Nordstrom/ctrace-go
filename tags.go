package ctrace

import (
	opentracing "github.com/opentracing/opentracing-go"
)

var (
	// SpanKindClient hints at client relationship between spans
	SpanKindClient = spanKindTag("span.kind", "client")

	// SpanKindServer hints at server relationship between spans
	SpanKindServer = spanKindTag("span.kind", "server")

	//////////////////////////////////////////////////////////////////////
	// Component name
	//////////////////////////////////////////////////////////////////////

	// Component is a low-cardinality identifier of the module, library,
	// or package that is generating a span.
	Component = stringTagName("component")

	//////////////////////////////////////////////////////////////////////
	// Peer tags. These tags can be emitted by either client-side of
	// server-side to describe the other side/service in a peer-to-peer
	// communications, like an RPC call.
	//////////////////////////////////////////////////////////////////////

	// PeerService records the service name of the peer
	PeerService = stringTagName("peer.service")

	// PeerHostname records the host name of the peer
	PeerHostname = stringTagName("peer.hostname")

	// PeerHostIPv4 records IP v4 host address of the peer
	PeerHostIPv4 = uint32TagName("peer.ipv4")

	// PeerHostIPv6 records IP v6 host address of the peer
	PeerHostIPv6 = stringTagName("peer.ipv6")

	// PeerPort records port number of the peer
	PeerPort = uint16TagName("peer.port")

	//////////////////////////////////////////////////////////////////////
	// HTTP Tags
	//////////////////////////////////////////////////////////////////////

	// HTTPUrl should be the URL of the request being handled in this segment
	// of the trace, in standard URI format. The protocol is optional.
	HTTPUrl = stringTagName("http.url")

	// HTTPMethod is the HTTP method of the request, and is case-insensitive.
	HTTPMethod = stringTagName("http.method")

	// HTTPStatusCode is the numeric HTTP status code (200, 404, etc) of the
	// HTTP response.
	HTTPStatusCode = uint16TagName("http.status_code")

	//////////////////////////////////////////////////////////////////////
	// Error Tag
	//////////////////////////////////////////////////////////////////////

	// Error indicates that operation represented by the span resulted in an error.
	Error = boolTagName("error")
)

func spanKindTag(k string, v string) func() opentracing.Tag {
	return func() opentracing.Tag {
		return opentracing.Tag{Key: k, Value: v}
	}
}

func stringTagName(k string) func(string) opentracing.Tag {
	return func(v string) opentracing.Tag {
		return opentracing.Tag{Key: k, Value: v}
	}
}

func uint32TagName(k string) func(uint32) opentracing.Tag {
	return func(v uint32) opentracing.Tag {
		return opentracing.Tag{Key: k, Value: v}
	}
}

func uint16TagName(k string) func(uint16) opentracing.Tag {
	return func(v uint16) opentracing.Tag {
		return opentracing.Tag{Key: k, Value: v}
	}
}

func boolTagName(k string) func(bool) opentracing.Tag {
	return func(v bool) opentracing.Tag {
		return opentracing.Tag{Key: k, Value: v}
	}
}
