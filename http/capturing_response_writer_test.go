package http

import (
	"fmt"

	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CapturingResponseWriter", func() {

	var (
		recorder *httptest.ResponseRecorder
		subject  CapturingResponseWriter
	)

	JustBeforeEach(func() {
		recorder = httptest.NewRecorder()
		subject = NewCapturingResponseWriter(recorder)
	})

	Describe("Write", func() {
		It("copies the data that is written to it to a different memory address", func() {
			dataToWrite := []byte("first write")
			subject.Write(dataToWrite)
			Expect(subject.ResponseBody()).To(Equal([]byte("first write")))
			Expect(fmt.Sprintf("%p", subject.ResponseBody())).ToNot(Equal(fmt.Sprintf("%p", dataToWrite)))
			Expect(recorder.Body.Bytes()).To(Equal([]byte("first write")))
		})
	})

	Describe("WriteHeader", func() {
		It("stores the status code to be retreived later", func() {
			subject.WriteHeader(http.StatusTeapot)
			Expect(subject.StatusCode()).To(Equal(http.StatusTeapot))
		})
	})
})
