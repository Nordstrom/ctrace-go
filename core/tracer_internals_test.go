package core

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracer Internals", func() {
	Describe("New", func() {
		It("creates tracer with stdout writer", func() {
			trc := New()
			t := trc.(*tracer)
			Ω(t.options).ShouldNot(BeNil())
			Ω((t.options.Writer == os.Stdout)).Should(BeTrue())
		})
	})
})
