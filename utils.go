package ledsim

import (
	"fmt"
	"time"
)

type TimingStats struct{}

func (t TimingStats) Execute(system *System, next func() error) error {
	start := time.Now()
	err := next()
	dur := time.Since(start)
	fmt.Println("frame time:", dur)

	return err
}
