package gnock_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/gabrielf/gnock"
)

func TestGnock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gnock Suite")
}

var _ = Describe("gnock", func() {
	It("fakes GET response", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply(200, "Hello, World!")

		req := newRequest("GET", "http://example.com/", nil)

		res := mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal("Hello, World!"))
	})
	It("panics if host is invalid, includes a path or a query string to avoid ambiguous usage", func() {
		Expect(func() {
			gnock.Gnock(":")
		}).To(Panic())
		Expect(func() {
			gnock.Gnock("http://example.com/with/path")
		}).To(Panic())
		Expect(func() {
			gnock.Gnock("http://example.com/?key=value")
		}).To(Panic())
	})
	It("can be used with the default transport", func() {
		gnock.Gnock("http://would-fail-in-unless-mocked").
			Get("/").
			Reply(123, "Hello, World!").
			ReplaceDefault()
		defer gnock.RestoreDefault()

		req := newRequest("GET", "http://would-fail-in-unless-mocked/", nil)

		res := mustRoundTrip(http.DefaultTransport, req)
		Expect(res.StatusCode).To(Equal(123))
		Expect(toString(res.Body)).To(Equal("Hello, World!"))
	})
	It("can chain multiple responses", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply(200, "Response 1").
			Get("/").
			Reply(200, "Response 2")

		req := newRequest("GET", "http://example.com/", nil)

		res := mustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("Response 1"))

		res = mustRoundTrip(transport, req)
		Expect(toString(res.Body)).To(Equal("Response 2"))
	})
	It("fakes responses for POST, PUT, OPTIONS, DELETE requests", func() {
		transport := gnock.Gnock("http://example.com").
			Post("/").
			Reply(201, "post").
			Put("/").
			Reply(202, "put").
			Options("/").
			Reply(204, "options").
			Delete("/").
			Reply(204, "delete")

		req := newRequest("POST", "http://example.com/", nil)
		res := mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(201))
		Expect(toString(res.Body)).To(Equal("post"))

		req = newRequest("PUT", "http://example.com/", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(202))
		Expect(toString(res.Body)).To(Equal("put"))

		req = newRequest("OPTIONS", "http://example.com/", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(204))
		Expect(toString(res.Body)).To(Equal("options"))

		req = newRequest("DELETE", "http://example.com/", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(204))
		Expect(toString(res.Body)).To(Equal("delete"))
	})
	It("makes URL interpolation easy", func() {
		transport := gnock.Gnock("http://example.com").
			Postf("/%s", "path").
			Reply(201, "post").
			Putf("/%s", "path").
			Reply(202, "put").
			Optionsf("/%s", "path").
			Reply(204, "options").
			Deletef("/%s", "path").
			Reply(204, "delete")

		req := newRequest("POST", "http://example.com/path", nil)
		res := mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(201))
		Expect(toString(res.Body)).To(Equal("post"))

		req = newRequest("PUT", "http://example.com/path", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(202))
		Expect(toString(res.Body)).To(Equal("put"))

		req = newRequest("OPTIONS", "http://example.com/path", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(204))
		Expect(toString(res.Body)).To(Equal("options"))

		req = newRequest("DELETE", "http://example.com/path", nil)
		res = mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(204))
		Expect(toString(res.Body)).To(Equal("delete"))
	})
	It("intercepts custom HTTP methods", func() {
		transport := gnock.Gnock("http://example.com").
			Intercept("PROPFIND", "/").
			Reply(200, `<?xml version="1.0" encoding="utf-8" ?>`)

		req := newRequest("PROPFIND", "http://example.com/", nil)

		res := mustRoundTrip(transport, req)
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal(`<?xml version="1.0" encoding="utf-8" ?>`))
	})
	It("can intercept host and path using Regexps", func() {
		transport := gnock.GnockRegexp("http://.*\\.com").
			InterceptRegexp("GET", "/(fu)?bar").
			Times(2).
			Reply(200, "success")

		res := mustRoundTrip(transport, newRequest("GET", "http://example.com/fubar", nil))
		Expect(res.StatusCode).To(Equal(200))

		res = mustRoundTrip(transport, newRequest("GET", "http://other.com/bar", nil))
		Expect(res.StatusCode).To(Equal(200))
	})
	It("panics when no match is found for the request", func() {
		transport := gnock.Gnock("http://example.com")

		req := newRequest("GET", "http://other.com/", nil)

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("describes existing interceptors and request that failed on panic", func(done Done) {
		transport := gnock.Gnock("http://example.com").
			Get("/path").
			Reply(200, "OK").
			Gnock("http://www.example.com").
			Post("/form").
			ReplyJSON(201, `{"key":"value"}`).
			GnockRegexp("^http://.*\\.example\\.com$").
			PutRegexp("/widgets/[\\d]+").
			Reply(201, "")

		req := newRequest("GET", "http://other.com/index.html", nil)

		defer func() {
			if err := recover(); err != nil {
				Expect(err).To(ContainSubstring("GET http://other.com/index.html"))
				Expect(err).To(ContainSubstring("GET http://example.com/path"))
				Expect(err).To(ContainSubstring("POST http://www.example.com/form"))
				Expect(err).To(ContainSubstring("PUT ^http://.*\\.example\\.com$/widgets/[\\d]+"))
				close(done)
			}
		}()

		transport.RoundTrip(req)
	})
	It("describes in error message if there are no interceptors", func(done Done) {
		transport := gnock.Gnock("http://example.com")

		req := newRequest("GET", "http://other.com/", nil)

		defer func() {
			if err := recover(); err != nil {
				Expect(err).To(ContainSubstring("interceptors:\nnone"))
				close(done)
			}
		}()

		transport.RoundTrip(req)
	})
	It("describes how to add the missing interceptor on panic", func(done Done) {
		transport := gnock.Gnock("http://example.com")

		req := newRequest("GET", "http://other.com/index.html", nil)

		defer func() {
			if err := recover(); err != nil {
				Expect(err).To(ContainSubstring(`gnock.Gnock("http://other.com").
	Get("/index.html").
	Reply(200, "OK")`))
				close(done)
			}
		}()

		transport.RoundTrip(req)
	})
	It("describes how to add the missing interceptor on panic even for custom HTTP method", func(done Done) {
		transport := gnock.Gnock("http://example.com")

		req := newRequest("PROPFIND", "http://other.com/", nil)

		defer func() {
			if err := recover(); err != nil {
				Expect(err).To(ContainSubstring(`gnock.Gnock("http://other.com").
	Intercept("PROPFIND", "/").
	Reply(200, "OK")`))
				close(done)
			}
		}()

		transport.RoundTrip(req)
	})
	It("uses an added interceptor only once", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply(200, "Hello, World!")

		req := newRequest("GET", "http://example.com/", nil)

		mustRoundTrip(transport, req)

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("can repeat a response several times", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Times(4).
			Reply(200, "Hello, World!")

		req := newRequest("GET", "http://example.com/", nil)

		for i := 0; i < 4; i++ {
			mustRoundTrip(transport, req)
		}

		Expect(func() {
			transport.RoundTrip(req)
		}).To(Panic())
	})
	It("can verify that all interceptors have been used with IsDone()", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply(200, "body")

		// Below IsDone is run after this method has completed at which time
		// the above interceptor will have been called.
		defer transport.IsDone()

		// Below IsDone is run now at which time the above interceptor has not
		// been called which leads to a panic.
		Expect(transport.IsDone).To(Panic())

		mustRoundTrip(transport, newRequest("GET", "http://example.com/", nil))
	})
	It("can fake requests to multiple domains", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Reply(200, "body").
			Gnock("http://other.com").
			Get("/").
			Reply(200, "other")

		res := mustRoundTrip(transport, newRequest("GET", "http://example.com/", nil))
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal("body"))

		res = mustRoundTrip(transport, newRequest("GET", "http://other.com/", nil))
		Expect(res.StatusCode).To(Equal(200))
		Expect(toString(res.Body)).To(Equal("other"))
	})
	It("allows responses to be customized with custom responder function", func() {
		transport := gnock.Gnock("http://example.com").
			Get("/").
			Respond(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Request:    req,
					StatusCode: 418,
					Body:       ioutil.NopCloser(bytes.NewBufferString("I'm a teapot")),
				}, nil
			})

		res := mustRoundTrip(transport, newRequest("GET", "http://example.com/", nil))
		Expect(res.StatusCode).To(Equal(418))
		Expect(toString(res.Body)).To(Equal("I'm a teapot"))
	})
	Describe("A partially defined interceptor", func() {
		var transport *gnock.Scope
		BeforeEach(func() {
			transport = gnock.Gnock("http://example.com")
			transport.Get("/")
		})
		It("is not matched", func(done Done) {
			defer func() {
				if err := recover(); err != nil {
					if re, ok := err.(runtime.Error); ok {
						Fail(fmt.Sprintf("Expected a panic but not a runtime error. Got: %s\n%s", re.Error(), string(debug.Stack())))
					}
					close(done)
				}
			}()

			transport.RoundTrip(newRequest("GET", "http://example.com/", nil))
		})
		It("is described as partially defined in panic", func(done Done) {
			defer func() {
				if err := recover(); err != nil {
					if s, ok := err.(string); ok {
						Expect(s).To(ContainSubstring("http://example.com/ (partially defined)"))
						close(done)
						return
					}
					panic(err)
				}
			}()

			transport.RoundTrip(newRequest("GET", "http://example.com/", nil))
		})
	})
	It("can fake errors", func() {
		kaboom := fmt.Errorf("kaboom")

		transport := gnock.Gnock("http://example.com").
			Get("/").
			ReplyError(kaboom)

		_, err := transport.RoundTrip(newRequest("GET", "http://example.com/", nil))
		Expect(err).To(Equal(kaboom))
	})
	Describe("Faking JSON", func() {
		It("can fake JSON responses using strings", func() {
			transport := gnock.Gnock("http://example.com").
				Get("/json").
				ReplyJSON(200, `{"key":"value"}`)

			res := mustRoundTrip(transport, newRequest("GET", "http://example.com/json", nil))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(toString(res.Body)).To(MatchJSON(`{"key":"value"}`))
		})
		It("can fake JSON responses using structs", func() {
			type Widget struct {
				Key string `json:"key"`
			}

			transport := gnock.Gnock("http://example.com").
				Get("/json").
				ReplyJSON(200, Widget{Key: "value"})

			res := mustRoundTrip(transport, newRequest("GET", "http://example.com/json", nil))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(toString(res.Body)).To(MatchJSON(`{"key":"value"}`))
		})
		It("can fake JSON responses using map[string]interface{}", func() {
			json := map[string]interface{}{
				"key": "value",
			}

			transport := gnock.Gnock("http://example.com").
				Get("/json").
				ReplyJSON(200, json)

			res := mustRoundTrip(transport, newRequest("GET", "http://example.com/json", nil))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(toString(res.Body)).To(MatchJSON(`{"key":"value"}`))
		})
	})
	Describe("An interceptor with default reply headers", func() {
		var interceptor *gnock.Interceptor

		BeforeEach(func() {
			interceptor = gnock.Gnock("http://example.com").
				DefaultReplyHeaders(http.Header{
					"Location": []string{"/login"},
					"Date":     []string{"2015-09-10"},
				}).
				Get("/")
		})
		It("responds with the set headers", func() {
			transport := interceptor.Reply(200, "OK")

			res := mustRoundTrip(transport, newRequest("GET", "http://example.com/", nil))
			Expect(res.Header["Location"]).To(Equal([]string{"/login"}))
			Expect(res.Header["Date"]).To(Equal([]string{"2015-09-10"}))
		})
		It("does not overwrite headers already set in a responder", func() {
			transport := interceptor.Respond(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Location": []string{"/logout"},
					},
				}, nil
			})

			res := mustRoundTrip(transport, newRequest("GET", "http://example.com/", nil))
			Expect(res.Header["Location"]).To(Equal([]string{"/logout"}))
			Expect(res.Header["Date"]).To(Equal([]string{"2015-09-10"}))
		})
	})
})

func newRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	Expect(err).ToNot(HaveOccurred())
	return req
}

func mustRoundTrip(transport http.RoundTripper, req *http.Request) *http.Response {
	res, err := transport.RoundTrip(req)
	Expect(err).ToNot(HaveOccurred())
	return res
}

func toString(reader io.Reader) string {
	data, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}
