// Package gnock is used to mock HTTP responses in tests.
package gnock

import (
	"net/http"
	"regexp"
)

// Gnock is one of the entry points for setting up mock responses. It takes
// the host that Gnock will intercept and returns a Scope which is used to
// setup the response(s) to send.
func Gnock(host string) *Scope {
	return NewScope(nil, host)
}

// GnockRegexp is another entry points for setting up mock responses. The
// difference from the Gnock(…) function is that the host will be compiled
// as a normal go regexp to enable matching a broader set of hosts. If you
// want to match any host simply use this method with the parameter ".*"
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
