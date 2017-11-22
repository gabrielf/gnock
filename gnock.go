package gnock

import (
	"net/http"
	"regexp"
)

func Gnock(host string) *Scope {
	return NewScope(nil, host)
}

func GnockRegexp(host string) *Scope {
	return NewRegexpScope(nil, regexp.MustCompile(host))
}

var originalDefaultTransport http.RoundTripper

func RestoreDefault() {
	if originalDefaultTransport != nil {
		http.DefaultTransport = originalDefaultTransport
		originalDefaultTransport = nil
	}
}
