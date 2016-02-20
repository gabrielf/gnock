package gnock

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type Interceptor struct {
	scope     *Scope
	method    string
	path      string
	status    int
	body      string
	responder Responder
	times     int
}

type Responder func(*http.Request) (*http.Response, error)

func NewInterceptor(scope *Scope, method, path string) *Interceptor {
	return &Interceptor{
		scope:  scope,
		method: method,
		path:   path,
		times:  1,
	}
}

func (i *Interceptor) Times(times int) *Interceptor {
	i.times = times
	return i
}

func (i *Interceptor) Reply(status int, body string) *Scope {
	i.status = status
	i.body = body
	return i.scope
}

func (i *Interceptor) Respond(responder Responder) *Scope {
	i.responder = responder
	return i.scope
}

func (i *Interceptor) intercepts(req *http.Request) bool {
	if i.times < 1 {
		return false
	}
	if !i.scope.intercepts(req) {
		return false
	}
	if req.Method != i.method || req.URL.Path != i.path {
		return false
	}
	return true
}

func (i *Interceptor) respond(req *http.Request) (*http.Response, error) {
	i.times--

	if i.responder != nil {
		return i.responder(req)
	}

	return &http.Response{
		Request:    req,
		StatusCode: i.status,
		Body:       ioutil.NopCloser(bytes.NewBufferString(i.body)),
	}, nil
}
