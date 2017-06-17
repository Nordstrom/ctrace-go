package core

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("jsonEncoder", func() {

	var (
		json  jsonEncoder
		bytes []byte
	)

	BeforeEach(func() {
		json = jsonEncoder{}
		bytes = []byte{}
	})

	Describe("encodeKey", func() {
		It("encodes key", func() {
			bytes = json.encodeKey(bytes, "mykey")
			Ω(string(bytes)).Should(Equal(`"mykey":`))
		})

		It("encodes key with symbols", func() {
			bytes = json.encodeKey(bytes, "mykey#$%^^&++=~")
			Ω(string(bytes)).Should(Equal(`"mykey#$%^^&++=~":`))
		})
	})

	Describe("encodeKeyBool", func() {
		It("encodes true", func() {
			bytes = json.encodeKeyBool(bytes, "mykey", true)
			Ω(string(bytes)).Should(Equal(`"mykey":true`))
		})

		It("encodes false", func() {
			bytes = json.encodeKeyBool(bytes, "mykey", false)
			Ω(string(bytes)).Should(Equal(`"mykey":false`))
		})
	})

	Describe("encodeKeyFloat", func() {
		It("encodes simple float", func() {
			bytes = json.encodeKeyFloat(bytes, "mykey", 1.5)
			Ω(string(bytes)).Should(Equal(`"mykey":1.5`))
		})
	})

	Describe("encodeKeyID", func() {
		It("encodes id with padding", func() {
			bytes = json.encodeKeyID(bytes, "mykey", 123)
			Ω(string(bytes)).Should(Equal(`"mykey":"000000000000007b"`))
		})

		It("encodes id without padding", func() {
			bytes = json.encodeKeyID(bytes, "mykey", 0x8a89382918382c7b)
			Ω(string(bytes)).Should(Equal(`"mykey":"8a89382918382c7b"`))
		})
	})

	Describe("encodeKeyInt", func() {
		It("encodes simple int", func() {
			bytes = json.encodeKeyInt(bytes, "mykey", 123)
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})
	})

	Describe("encodeKeyString", func() {
		It("encodes simple string", func() {
			bytes = json.encodeKeyString(bytes, "mykey", "mystring")
			Ω(string(bytes)).Should(Equal(`"mykey":"mystring"`))
		})

		It("encodes string with quotes", func() {
			bytes = json.encodeKeyString(bytes, "mykey", `"mystring"`)
			Ω(string(bytes)).Should(Equal(`"mykey":"\"mystring\""`))
		})
	})

	Describe("encodeKeyUint", func() {
		It("encodes simple uint", func() {
			bytes = json.encodeKeyUint(bytes, "mykey", 123)
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})
	})

	Describe("encodeKeyValue", func() {
		It("encodes bool", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", true)
			Ω(string(bytes)).Should(Equal(`"mykey":true`))
		})

		It("encodes float", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", 1.5)
			Ω(string(bytes)).Should(Equal(`"mykey":1.5`))
		})

		It("encodes float32", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", float32(1.5))
			Ω(string(bytes)).Should(Equal(`"mykey":1.5`))
		})

		It("encodes float64", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", float64(1.5))
			Ω(string(bytes)).Should(Equal(`"mykey":1.5`))
		})

		It("encodes int", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", 123)
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes int8", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", int8(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes int16", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", int16(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes int32", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", int32(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes int64", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", int64(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes uint", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", uint(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes uint8", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", uint8(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes uint16", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", uint16(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes uint32", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", uint32(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes uint64", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", uint64(123))
			Ω(string(bytes)).Should(Equal(`"mykey":123`))
		})

		It("encodes string", func() {
			bytes = json.encodeKeyValue(bytes, "mykey", "mystring")
			Ω(string(bytes)).Should(Equal(`"mykey":"mystring"`))
		})
	})

	Describe("encodeString", func() {
		It("encodes simple string", func() {
			bytes = json.encodeString(bytes, "mystring")
			Ω(string(bytes)).Should(Equal(`mystring`))
		})

		It("encodes string with unsafe chars", func() {
			bytes = json.encodeString(bytes, "mystring\"\t\n\r")
			Ω(string(bytes)).Should(Equal(`mystring\"\t\n\r`))
		})
	})
})
