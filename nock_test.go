package nock_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gabrielf/nock"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nock Suite")
}

var _ = Describe("nock", func() {
	It("fakes GET respons", func() {
		transport := nock.Nock("http://example.com").
			Get("/").
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		res, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal("Hello, World!"))
	})
	It("can chain multiple responses", func() {
		transport := nock.Nock("http://example.com").
			Get("/").
			Reply("Response 1").
			Get("/").
			Reply("Response 2")

		req := NewRequest("GET", "http://example.com/", nil)

		res, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("Response 1"))

		res, err = transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("Response 2"))
	})
	It("fakes responses for POST, PUT, OPTIONS, DELETE requests", func() {
		transport := nock.Nock("http://example.com").
			Post("/").
			Reply("post").
			Put("/").
			Reply("put").
			Options("/").
			Reply("options").
			Delete("/").
			Reply("delete")

		req := NewRequest("POST", "http://example.com/", nil)

		res, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("post"))

		req = NewRequest("PUT", "http://example.com/", nil)

		res, err = transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("put"))

		req = NewRequest("OPTIONS", "http://example.com/", nil)

		res, err = transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("options"))

		req = NewRequest("DELETE", "http://example.com/", nil)

		res, err = transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("delete"))
	})
	It("intercepts custom HTTP methods", func() {
		transport := nock.Nock("http://example.com").
			Intercept("PROPFIND", "/").
			Reply(`<?xml version="1.0" encoding="utf-8" ?>`)

		req := NewRequest("PROPFIND", "http://example.com/", nil)

		res, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal(`<?xml version="1.0" encoding="utf-8" ?>`))
	})
	It("panics when no match is found for the request", func() {
		transport := nock.Nock("http://example.com")

		req := NewRequest("GET", "http://other.com/", nil)

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("uses an added interceptor only once", func() {
		transport := nock.Nock("http://example.com").
			Get("/").
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		_, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("can repeat response several times", func() {
		transport := nock.Nock("http://example.com").
			Get("/").
			Times(4).
			Reply("Hello, World!")

		req := NewRequest("GET", "http://example.com/", nil)

		for i := 0; i < 4; i++ {
			_, err := transport.RoundTrip(req)
			Expect(err).ToNot(HaveOccurred())
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

func toString(reader io.Reader) string {
	data, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
