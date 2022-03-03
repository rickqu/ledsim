package effects

import (
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

const StandardPeriod = 12800 * time.Millisecond

var Golds = []colorful.Color{
	// {255, 255, 0},
	// {212, 175, 55},
	// {207, 181, 59},
	// {197, 179, 88},
	{250 / 255.0, 130 / 255.0, 0},
	{255 / 255.0, 255 / 255.0, 255 / 255.0},
	{230 / 255.0, 207 / 255.0, 0},
	{230 / 255.0, 175.0 / 255.0, 0 / 255.0},
	{255 / 255.0, 166.0 / 255.0, 76 / 255.0},
	// {250 / 255.0 / 10, 130 / 255.0 / 10, 0},
	// {153, 101, 21},
	// {244, 163, 0},
}
