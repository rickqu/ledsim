package ledsim

import (
	"fmt"
	"os"
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

type StallCheck struct{}

func (c StallCheck) Execute(system *System, next func() error) error {
	start := time.Now()
	t := time.NewTicker(time.Millisecond * 100)
	go func() {
		for range t.C {
			since := time.Since(start)
			if since >= time.Millisecond*500 {
				fmt.Println("stall timeout, quitting")
				os.Exit(1)
			} else if since > time.Millisecond*100 {
				fmt.Println("warn: frame stalled for:", since)
			}
		}
	}()
	err := next()
	t.Stop()

	return err
}
