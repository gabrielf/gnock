package gnock

import (
	"fmt"
	"net/http"
)

type Scope struct {
	parent       *Scope
	child        *Scope
	host         string
	interceptors []*Interceptor
}

var _ http.RoundTripper = (*Scope)(nil)

func NewScope(parent *Scope, host string) *Scope {
	return &Scope{
		parent:       parent,
		host:         host,
		interceptors: make([]*Interceptor, 0),
	}
}

func (s *Scope) Gnock(host string) *Scope {
	s.child = NewScope(s, host)
	return s.child
}

func (s *Scope) RoundTrip(req *http.Request) (*http.Response, error) {
	// This method makes sure to start at the root of the scope hierarchy...
	if s.parent != nil {
		return s.parent.RoundTrip(req)
	}

	return s.roundTrip(req)
}

func (s *Scope) roundTrip(req *http.Request) (*http.Response, error) {
	// ...and this method serves matched requests down the scope hierarchy.
	for _, interceptor := range s.interceptors {
		if interceptor.intercepts(req) {
			return interceptor.respond(req)
		}
	}

	if s.child != nil {
		return s.child.roundTrip(req)
	}

	panic(fmt.Sprintf("No match found for request: %+v", req))
}

func (s *Scope) IsDone() {
	for _, interceptor := range s.interceptors {
		if interceptor.times > 0 {
			panic(fmt.Sprintf("Not all interceptors have been used! Found: %+v", interceptor))
		}
	}
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
