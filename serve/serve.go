package serve

type serve interface {
	Serve()
}

func Launch(s serve) {
	s.Serve()
}
