package gnock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type Interceptor struct {
	scope      *Scope
	method     string
	path       string
	pathRegexp *regexp.Regexp
	responder  Responder
	times      int
}

type Responder func(*http.Request) (*http.Response, error)

func NewInterceptor(scope *Scope, method string, path string) *Interceptor {
	return &Interceptor{
		scope:  scope,
		method: method,
		path:   path,
		times:  1,
	}
}

func NewRegexpInterceptor(scope *Scope, method string, pathRegexp *regexp.Regexp) *Interceptor {
	return &Interceptor{
		scope:      scope,
		method:     method,
		pathRegexp: pathRegexp,
		times:      1,
	}
}

func (i *Interceptor) Times(times int) *Interceptor {
	i.times = times
	return i
}

func (i *Interceptor) Reply(status int, body string) *Scope {
	return i.Respond(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			Request:    req,
			StatusCode: status,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}, nil
	})
}

func (i *Interceptor) ReplyError(err error) *Scope {
	return i.Respond(func(req *http.Request) (*http.Response, error) {
		return nil, err
	})
}

func (i *Interceptor) ReplyJSON(status int, json interface{}) *Scope {
	return i.Respond(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			Request:    req,
			StatusCode: status,
			Body:       ioutil.NopCloser(bytes.NewBufferString(jsonToString(json))),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})
}

func (i *Interceptor) Respond(responder Responder) *Scope {
	i.responder = responder
	return i.scope
}

func (i *Interceptor) String() string {
	return fmt.Sprintf("%s %s%s\n", i.method, i.scope.String(), i.describePath())
}

func (i *Interceptor) describePath() string {
	if i.pathRegexp != nil {
		return i.pathRegexp.String()
	}
	return i.path
}

func (i *Interceptor) intercepts(req *http.Request) bool {
	if i.times < 1 {
		return false
	}
	if !i.scope.intercepts(req) {
		return false
	}
	if req.Method != i.method {
		return false
	}
	if i.pathRegexp != nil {
		return i.pathRegexp.MatchString(req.URL.Path)
	}
	return i.path == req.URL.Path
}

func (i *Interceptor) respond(req *http.Request) (*http.Response, error) {
	i.times--

	res, err := i.responder(req)
	if err != nil {
		// We must return res here since real HTTP requests might do that in some cases
		return res, err
	}

	return i.setDefaultHeaders(res), nil
}

func (i *Interceptor) setDefaultHeaders(res *http.Response) *http.Response {
	if len(i.scope.defaultHeaders) > 0 && res.Header == nil {
		res.Header = make(http.Header, 0)
	}
	for headerKey, headerValues := range i.scope.defaultHeaders {
		if len(res.Header[headerKey]) == 0 {
			res.Header[headerKey] = headerValues
		}
	}
	return res
}

func jsonToString(input interface{}) string {
	if str, ok := input.(string); ok {
		return str
	}

	buf, err := json.Marshal(input)
	if err != nil {
		panic(err.Error())
	}
	return string(buf)
}
