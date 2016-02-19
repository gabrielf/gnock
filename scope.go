package gnock

import (
	"fmt"
	"net/http"
)

type Scope struct {
	host         string
	interceptors []*Interceptor
}

var _ http.RoundTripper = (*Scope)(nil)

func NewScope(host string) *Scope {
	return &Scope{
		host:         host,
		interceptors: make([]*Interceptor, 0),
	}
}

func (s *Scope) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, interceptor := range s.interceptors {
		if interceptor.intercepts(req) {
			return interceptor.respond(req)
		}
	}

	panic(fmt.Sprintf("No match found for request: %+v", req))
}

func (s *Scope) Intercept(method, path string) *Interceptor {
	i := NewInterceptor(s, method, path)
	s.interceptors = append(s.interceptors, i)
	return i
}

func (s *Scope) Get(path string) *Interceptor {
	return s.Intercept("GET", path)
}

func (s *Scope) Post(path string) *Interceptor {
	return s.Intercept("POST", path)
}

func (s *Scope) Put(path string) *Interceptor {
	return s.Intercept("PUT", path)
}

func (s *Scope) Options(path string) *Interceptor {
	return s.Intercept("OPTIONS", path)
}

func (s *Scope) Delete(path string) *Interceptor {
	return s.Intercept("DELETE", path)
}

func (s *Scope) intercepts(req *http.Request) bool {
	return req.URL.Scheme+"://"+req.URL.Host == s.host
}
