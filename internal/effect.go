package internal

import "time"

type Effect interface {
	Apply(sys *System, now time.Time)
}
