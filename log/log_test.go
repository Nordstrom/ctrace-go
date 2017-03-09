package log

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Log", func() {
	It("ErrorKind", func() {
		fld := ErrorKind("kind")
		Ω(fld.Key()).Should(Equal("error.kind"))
		Ω(fld.Value()).Should(Equal("kind"))
	})

	It("ErrorObject", func() {
		fld := ErrorObject(errors.New("errmsg"))
		Ω(fld.Key()).Should(Equal("error.object"))
		Ω(fld.Value()).Should(Equal("errmsg"))
	})

	It("Event", func() {
		fld := Event("ev")
		Ω(fld.Key()).Should(Equal("event"))
		Ω(fld.Value()).Should(Equal("ev"))
	})

	It("Message", func() {
		fld := Message("msg")
		Ω(fld.Key()).Should(Equal("message"))
		Ω(fld.Value()).Should(Equal("msg"))
	})

	It("ErrorKind", func() {
		fld := Stack("stk")
		Ω(fld.Key()).Should(Equal("stack"))
		Ω(fld.Value()).Should(Equal("stk"))
	})
})
