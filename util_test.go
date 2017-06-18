package ctrace

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("util", func() {
	Describe("httpRemoteAddr", func() {
		It("handles Http-Client-Id", func() {
			Expect(httpRemoteAddr(map[string]string{"http-Client-ID": "hcival"})).To(Equal("hcival"))
			Expect(httpRemoteAddr(map[string]string{"HTTP-Client-ID": "hcival"})).To(Equal("hcival"))
		})

		It("handles X-Forwarded-For", func() {
			Expect(httpRemoteAddr(map[string]string{"x-Forwarded-For": "xffval"})).To(Equal("xffval"))
		})

		It("handles X-Forwarded", func() {
			Expect(httpRemoteAddr(map[string]string{"x-Forwarded": "xfval"})).To(Equal("xfval"))
		})

		It("handles X-Cluster-Client-Ip", func() {
			Expect(httpRemoteAddr(map[string]string{"x-ClusTer-client-ip": "xccival"})).To(Equal("xccival"))
		})

		It("handles Forwarded-For", func() {
			Expect(httpRemoteAddr(map[string]string{"forwarded-For": "ffval"})).To(Equal("ffval"))
		})

		It("handles Forwarded", func() {
			Expect(httpRemoteAddr(map[string]string{"ForwardeD": "fval"})).To(Equal("fval"))
		})

		It("handles Remote-Addr", func() {
			Expect(httpRemoteAddr(map[string]string{"RemoTe-addr": "raval"})).To(Equal("raval"))
		})

		It("handles Priority check 1", func() {
			Expect(httpRemoteAddr(map[string]string{
				"x-Forwarded-For": "xffval",
				"http-client-id":  "hcival",
				"remote-addr":     "raval",
			})).To(Equal("hcival"))
		})

		It("handles Priority check 1", func() {
			Expect(httpRemoteAddr(map[string]string{
				"x-Forwarded-For": "xffval",
				"forwarded-for":   "ffval",
			})).To(Equal("xffval"))
		})

		It("handles none", func() {
			Expect(httpRemoteAddr(map[string]string{
				"x-Forwarded-Forxx": "xffval",
				"forwarded-foryy":   "ffval",
			})).To(Equal(""))
		})
	})

	Describe("httpUserAgent", func() {
		It("handles val", func() {
			Expect(httpUserAgent(map[string]string{
				"user-AgenT": "uaval",
			})).To(Equal("uaval"))
		})

		It("handles none", func() {
			Expect(httpUserAgent(map[string]string{
				"user-agentxx": "uaval",
			})).To(Equal(""))
		})
	})
})
