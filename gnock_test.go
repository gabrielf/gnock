package gnock_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gabrielf/gnock"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gnock Suite")
}

var _ = Describe("gnock", func() {
	It("fakes GET respons", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		res := MustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal("Hello, World!"))
	})
	It("can chain multiple responses", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply("Response 1").
			Get("/").
			Reply("Response 2")

		req := NewRequest("GET", "http://example.com/", nil)

		res := MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("Response 1"))

		res = MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("Response 2"))
	})
	It("fakes responses for POST, PUT, OPTIONS, DELETE requests", func() {
		transport := gnock.Gnock("http://example.com").
			Post("/").
			Reply("post").
			Put("/").
			Reply("put").
			Options("/").
			Reply("options").
			Delete("/").
			Reply("delete")

		req := NewRequest("POST", "http://example.com/", nil)
		res := MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("post"))

		req = NewRequest("PUT", "http://example.com/", nil)
		res = MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("put"))

		req = NewRequest("OPTIONS", "http://example.com/", nil)
		res = MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("options"))

		req = NewRequest("DELETE", "http://example.com/", nil)
		res = MustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("delete"))
	})
	It("intercepts custom HTTP methods", func() {
		transport := gnock.Gnock("http://example.com").
			Intercept("PROPFIND", "/").
			Reply(`<?xml version="1.0" encoding="utf-8" ?>`)

		req := NewRequest("PROPFIND", "http://example.com/", nil)

		res := MustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal(`<?xml version="1.0" encoding="utf-8" ?>`))
	})
	It("panics when no match is found for the request", func() {
		transport := gnock.Gnock("http://example.com")

		req := NewRequest("GET", "http://other.com/", nil)

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("uses an added interceptor only once", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		MustRoundTrip(transport, req)

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("can repeat response several times", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Times(4).
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		for i := 0; i < 4; i++ {
			MustRoundTrip(transport, req)
		}

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
})

func NewRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	Expect(err).ToNot(HaveOccurred())
	return req
}

func MustRoundTrip(transport http.RoundTripper, req *http.Request) *http.Response {
	res, err := transport.RoundTrip(req)
	Expect(err).ToNot(HaveOccurred())
	return res
}

func toString(reader io.Reader) string {
	data, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
