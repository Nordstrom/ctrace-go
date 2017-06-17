package ctrace

import "strings"

func httpRemoteAddr(hdrs map[string]string) string {
	var h [7]string
	for k, v := range hdrs {
		key := strings.ToLower(k)

		switch key {
		case "http-client-id":
			h[0] = v
		case "x-forwarded-for":
			h[1] = v
		case "x-forwarded":
			h[2] = v
		case "x-cluster-client-ip":
			h[3] = v
		case "forwarded-for":
			h[4] = v
		case "forwarded":
			h[5] = v
		case "remote-addr":
			h[6] = v
		}
	}

	for _, i := range h {
		if i != "" {
			return i
		}
	}

	return ""
}

func httpUserAgent(hdrs map[string]string) string {
	for k, v := range hdrs {
		key := strings.ToLower(k)
		if key == "user-agent" {
			return v
		}
	}
	return ""
}
