package http

import (
	"fmt"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CapturingResponseWriter", func() {

	var (
		subject CapturingResponseWriter
		headers http.Header
	)

	JustBeforeEach(func() {
		subject = NewCapturingResponseWriter(headers)
	})

	Describe("Write", func() {
		It("copies the data that is written to it to a different memory address", func() {
			dataToWrite := []byte("first write")
			subject.Write(dataToWrite)
			Expect(subject.ResponseBody()).To(Equal([]byte("first write")))
			Expect(fmt.Sprintf("%p", subject.ResponseBody())).ToNot(Equal(fmt.Sprintf("%p", dataToWrite)))
		})
	})

	Describe("Header", func() {
		BeforeEach(func() {
			headers = http.Header{}
			headers.Add("some", "header")
		})

		It("stores the headers that are passed in", func() {
			Expect(subject.Header()).To(Equal(headers))
		})
	})

	Describe("WriteHeader", func() {
		It("stores the status code to be retreived later", func() {
			subject.WriteHeader(http.StatusTeapot)
			Expect(subject.StatusCode()).To(Equal(http.StatusTeapot))
		})
	})
})
