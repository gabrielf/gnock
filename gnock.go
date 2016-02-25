package gnock

import "regexp"

func Gnock(host string) *Scope {
	return NewScope(nil, host)
}

func GnockRegexp(host string) *Scope {
	return NewRegexpScope(nil, regexp.MustCompile(host))
}
