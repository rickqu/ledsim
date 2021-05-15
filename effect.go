package main

import "time"

type Effect interface {
	Apply(sys *System, now time.Time)
}
