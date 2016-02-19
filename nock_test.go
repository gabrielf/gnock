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

		req, err := http.NewRequest("GET", "http://example.com/", nil)
		Expect(err).ToNot(HaveOccurred())

		res, err := transport.RoundTrip(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(toString(res.Body)).To(Equal("Hello, World!"))
	})
})

func toString(reader io.Reader) string {
	data, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
