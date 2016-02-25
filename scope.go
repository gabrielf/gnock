package gnock

import (
	"fmt"
	"net/http"
	"regexp"
)

type Scope struct {
	parent         *Scope
	child          *Scope
	host           string
	hostRegexp     *regexp.Regexp
	interceptors   []*Interceptor
	defaultHeaders http.Header
}

var _ http.RoundTripper = (*Scope)(nil)

func NewScope(parent *Scope, host string) *Scope {
	return &Scope{
		parent:         parent,
		host:           host,
		interceptors:   make([]*Interceptor, 0),
		defaultHeaders: make(http.Header, 0),
	}
}

func NewRegexpScope(parent *Scope, hostRegexp *regexp.Regexp) *Scope {
	return &Scope{
		parent:         parent,
		hostRegexp:     hostRegexp,
		interceptors:   make([]*Interceptor, 0),
		defaultHeaders: make(http.Header, 0),
	}
}

func (s *Scope) Gnock(host string) *Scope {
	s.child = NewScope(s, host)
	return s.child
}

func (s *Scope) GnockRegexp(host string) *Scope {
	s.child = NewRegexpScope(s, regexp.MustCompile(host))
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

	panic(fmt.Sprintf("Gnock found no match for request: %s\n\nRegistered interceptors:\n%s", describeRequest(req), describeInterceptors(s)))
}

func (s *Scope) IsDone() {
	for _, interceptor := range s.interceptors {
		if interceptor.times > 0 {
			panic(fmt.Sprintf("Not all interceptors have been used! Found: %+v", interceptor))
		}
	}
}

func (s *Scope) DefaultReplyHeaders(headers http.Header) *Scope {
	s.defaultHeaders = headers
	return s
}

func (s *Scope) Intercept(method, path string) *Interceptor {
	i := NewInterceptor(s, method, path)
	s.interceptors = append(s.interceptors, i)
	return i
}

func (s *Scope) InterceptRegexp(method, path string) *Interceptor {
	i := NewRegexpInterceptor(s, method, regexp.MustCompile(path))
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

func (s *Scope) GetRegexp(path string) *Interceptor {
	return s.InterceptRegexp("GET", path)
}

func (s *Scope) PostRegexp(path string) *Interceptor {
	return s.InterceptRegexp("POST", path)
}

func (s *Scope) PutRegexp(path string) *Interceptor {
	return s.InterceptRegexp("PUT", path)
}

func (s *Scope) OptionsRegexp(path string) *Interceptor {
	return s.InterceptRegexp("OPTIONS", path)
}

func (s *Scope) DeleteRegexp(path string) *Interceptor {
	return s.InterceptRegexp("DELETE", path)
}

func (s *Scope) String() string {
	if s.hostRegexp != nil {
		return s.hostRegexp.String()
	}
	return s.host
}

func (s *Scope) intercepts(req *http.Request) bool {
	schemeAndHost := req.URL.Scheme + "://" + req.URL.Host
	if s.hostRegexp != nil {
		return s.hostRegexp.MatchString(schemeAndHost)
	}
	return s.host == schemeAndHost
}

func describeRequest(req *http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, req.URL.String())
}

func describeInterceptors(s *Scope) string {
	result := ""
	if s.parent != nil {
		result = describeInterceptors(s.parent)
	}
	for _, i := range s.interceptors {
		result += i.String()
	}
	return result
}
