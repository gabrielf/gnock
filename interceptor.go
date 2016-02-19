package nock

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type Interceptor struct {
	scope  *Scope
	method string
	path   string
	body   string
}

func NewInterceptor(scope *Scope, method, path string) *Interceptor {
	return &Interceptor{
		scope:  scope,
		method: method,
		path:   path,
	}
}

func (i *Interceptor) Reply(body string) *Scope {
	i.body = body
	return i.scope
}

func (i *Interceptor) intercepts(req *http.Request) bool {
	if !i.scope.intercepts(req) {
		return false
	}
	if req.Method != i.method || req.URL.Path != i.path {
		return false
	}
	return true
}

func (i *Interceptor) respond(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(i.body)),
	}, nil
}
