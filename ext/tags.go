package ext

import (
	opentracing "github.com/opentracing/opentracing-go"
)

// Tag keys (and values) for standard and recommended tags a given here.  This
// is to supplement and improve on the [Standard Tags](https://github.com/opentracing/opentracing-go/blob/master/ext/tags.go)
const (
	// SpanKindKey is the SpanKind tag key
	SpanKindKey = "span.kind"

	// SpanKindClientValue is the SpanKindClient tag value
	SpanKindClientValue = "client"

	// SpanKindServerValue is the SpanKindServer tag value
	SpanKindServerValue = "server"

	//////////////////////////////////////////////////////////////////////
	// Component name key
	//////////////////////////////////////////////////////////////////////

	// ComponentKey is the tag key for a low-cardinality identifier of the module,
	// library, or package that is generating a span.
	ComponentKey = "component"

	//////////////////////////////////////////////////////////////////////
	// Db keys
	//////////////////////////////////////////////////////////////////////

	// DbInstanceKey is the tag key for a Database instance name. E.g., In java, if the
	// jdbc.url="jdbc:mysql://127.0.0.1:3306/customers", the instance name is "customers".
	DbInstanceKey = "db.instance"

	// DbStatementKey is the tag key for a database statement for the given database
	// type. E.g., for db.type="SQL", "SELECT * FROM wuser_table"; for db.type="redis",
	// "SET mykey 'WuValue'".
	DbStatementKey = "db.statement"

	// DbTypeKey is the tag key for a Database type. For any SQL database, "sql". For others,
	// the lower-case database category, e.g. "cassandra", "hbase", or "redis".
	DbTypeKey = "db.type"

	// DbUserKey is the tag key for a Username for accessing database. E.g.,
	// "readonly_user" or "reporting_user"
	DbUserKey = "db.user"

	//////////////////////////////////////////////////////////////////////
	// Peer tag keys. These tags can be emitted by either client-side of
	// server-side to describe the other side/service in a peer-to-peer
	// communications, like an RPC call.
	//////////////////////////////////////////////////////////////////////

	// PeerServiceKey is the key for a tag that records the service name of the peer
	PeerServiceKey = "peer.service"

	// PeerHostnameKey is the key for a tag that records the host name of the peer
	PeerHostnameKey = "peer.hostname"

	// PeerHostIPv4Key is the key for a tag that records IP v4 host address of the peer
	PeerHostIPv4Key = "peer.ipv4"

	// PeerHostIPv6Key is the key for a tag that records IP v6 host address of the peer
	PeerHostIPv6Key = "peer.ipv6"

	// PeerPortKey is the key for a tag that records port number of the peer
	PeerPortKey = "peer.port"

	//////////////////////////////////////////////////////////////////////
	// HTTP Tag keys
	//////////////////////////////////////////////////////////////////////

	// HTTPUrlKey is the key for a tag that should be the URL of the request being
	// handled in this segment of the trace, in standard URI format. The protocol
	// is optional.
	HTTPUrlKey = "http.url"

	// HTTPMethodKey is the key for a tag that is the HTTP method of the request,
	// and is case-insensitive.
	HTTPMethodKey = "http.method"

	// HTTPStatusCodeKey is the numeric HTTP status code (200, 404, etc) of the
	// HTTP response.
	HTTPStatusCodeKey = "http.status_code"

	//////////////////////////////////////////////////////////////////////
	// Error Tag key
	//////////////////////////////////////////////////////////////////////

	// ErrorKey is the key for a tag that indicates that operation represented by
	// the span resulted in an error.
	ErrorKey = "error"

	//////////////////////////////////////////////////////////////////////
	// Recommended Tags
	//////////////////////////////////////////////////////////////////////

	// HTTPRemoteAddrKey is the key for a tag that reprents the X-Forwarded-For
	// header or Client IP of the caller
	HTTPRemoteAddrKey = "http.remote_addr"

	// HTTPUserAgentKey is the key for the UserAgent tag
	HTTPUserAgentKey = "http.user_agent"
)

var (
	// SpanKindClient hints at client relationship between spans
	SpanKindClient = spanKindTag(SpanKindKey, SpanKindClientValue)

	// SpanKindServer hints at server relationship between spans
	SpanKindServer = spanKindTag(SpanKindKey, SpanKindServerValue)

	//////////////////////////////////////////////////////////////////////
	// Component name
	//////////////////////////////////////////////////////////////////////

	// Component is a low-cardinality identifier of the module, library,
	// or package that is generating a span.
	Component = stringTagName(ComponentKey)

	//////////////////////////////////////////////////////////////////////
	// Db keys
	//////////////////////////////////////////////////////////////////////

	// DbInstance is the tag for a Database instance name. E.g., In java, if the
	// jdbc.url="jdbc:mysql://127.0.0.1:3306/customers", the instance name is "customers".
	DbInstance = stringTagName(DbInstanceKey)

	// DbStatement is the tag for a database statement for the given database
	// type. E.g., for db.type="SQL", "SELECT * FROM wuser_table"; for db.type="redis",
	// "SET mykey 'WuValue'".
	DbStatement = stringTagName(DbStatementKey)

	// DbType is the tag for a Database type. For any SQL database, "sql". For others,
	// the lower-case database category, e.g. "cassandra", "hbase", or "redis".
	DbType = stringTagName(DbTypeKey)

	// DbUser is the tag for a Username for accessing database. E.g.,
	// "readonly_user" or "reporting_user"
	DbUser = stringTagName(DbUserKey)

	//////////////////////////////////////////////////////////////////////
	// Peer tags. These tags can be emitted by either client-side of
	// server-side to describe the other side/service in a peer-to-peer
	// communications, like an RPC call.
	//////////////////////////////////////////////////////////////////////

	// PeerService records the service name of the peer
	PeerService = stringTagName(PeerServiceKey)

	// PeerHostname records the host name of the peer
	PeerHostname = stringTagName(PeerHostnameKey)

	// PeerHostIPv4 records IP v4 host address of the peer
	PeerHostIPv4 = uint32TagName(PeerHostIPv4Key)

	// PeerHostIPv6 records IP v6 host address of the peer
	PeerHostIPv6 = stringTagName(PeerHostIPv6Key)

	// PeerPort records port number of the peer
	PeerPort = uint16TagName(PeerPortKey)

	//////////////////////////////////////////////////////////////////////
	// HTTP Tags
	//////////////////////////////////////////////////////////////////////

	// HTTPUrl should be the URL of the request being handled in this segment
	// of the trace, in standard URI format. The protocol is optional.
	HTTPUrl = stringTagName(HTTPUrlKey)

	// HTTPMethod is the HTTP method of the request, and is case-insensitive.
	HTTPMethod = stringTagName(HTTPMethodKey)

	// HTTPStatusCode is the numeric HTTP status code (200, 404, etc) of the
	// HTTP response.
	HTTPStatusCode = intTagName(HTTPStatusCodeKey)

	//////////////////////////////////////////////////////////////////////
	// Error Tag
	//////////////////////////////////////////////////////////////////////

	// Error indicates that operation represented by the span resulted in an error.
	Error = boolTagName(ErrorKey)

	//////////////////////////////////////////////////////////////////////
	// Recommended Tags
	//////////////////////////////////////////////////////////////////////

	// HTTPRemoteAddr is the X-Forwarded-For header or Client IP
	HTTPRemoteAddr = stringTagName(HTTPRemoteAddrKey)

	// HTTPUserAgent is the
	HTTPUserAgent = stringTagName(HTTPUserAgentKey)
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

func intTagName(k string) func(int) opentracing.Tag {
	return func(v int) opentracing.Tag {
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
