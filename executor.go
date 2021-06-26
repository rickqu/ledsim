package ledsim

import (
	"context"
	"time"
)

type Executor struct {
	system     *System
	middleware []Middleware
	frameRate  int
}

func NewExecutor(system *System, frameRate int, middleware ...Middleware) *Executor {
	return &Executor{
		system:     system,
		frameRate:  frameRate,
		middleware: middleware,
	}
}

func (e *Executor) Run(ctx context.Context) error {
	t := time.NewTicker(time.Second / time.Duration(e.frameRate))
	for range t.C {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := e.RunMiddleware(0)
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return nil
}

func (e *Executor) RunMiddleware(i int) error {
	if i >= len(e.middleware) {
		return nil
	}

	return e.middleware[i].Execute(e.system, func() error {
		return e.RunMiddleware(i + 1)
	})
}
