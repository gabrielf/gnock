package nock

func Nock(host string) *Scope {
	return NewScope(host)
}
