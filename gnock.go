package gnock

func Gnock(host string) *Scope {
	return NewScope(nil, host)
}
