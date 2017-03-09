package ext_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestExt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ext Suite")
}
