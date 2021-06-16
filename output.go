package ledsim

type Output interface {
	Display(system *System)
}

func NewOutput(output Output) Middleware {
	return (MiddlewareFunc)(func(system *System, next func() error) error {
		output.Display(system)
		return next()
	})
}
