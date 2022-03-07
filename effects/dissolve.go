package effects

import (
	"fmt"
	"ledsim"
	"math"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type Dissolve struct {
	duration  time.Duration
	closed    []int          // index of which leds have dissolved
	currClust []int          // index of current cluster of LEDs we want to dissolve
	period    float64        // How long it will take a cluster to dissolve; between 0 and 1 is good
	reqColor  colorful.Color // required color; Use {0,0,0} for dissolving to black
	clustSize int            // How many clusters/groups of LEDs to dissolve at a time; between 30 - 80 is good
	remaining []int          // Index of LEDs that are remaining to be dissolved
	remCount  int            // Counts number LEDs remaining to be dissolved
	state     []colorful.Color
}

func NewDissolve(duration time.Duration, reqColor colorful.Color, clustSize int, period float64) *Dissolve {
	return &Dissolve{
		duration:  duration,
		reqColor:  reqColor,  // Use {0,0,0} for dissolving to black
		clustSize: clustSize, // between 30 - 80 is good
		period:    period,    // between 0 and 1 is good
	}
}

func (d *Dissolve) OnEnter(sys *ledsim.System) {
	ledCount := len(sys.LEDs)
	d.closed = make([]int, 0)
	d.currClust = make([]int, 0)
	d.remaining = make([]int, 0)
	d.remCount = ledCount
	d.state = make([]colorful.Color, ledCount+1)

	for _, led := range sys.LEDs {
		d.remaining = append(d.remaining, led.ID)
		d.state[led.ID] = led.Color
	}

}

func (d *Dissolve) Eval(progress float64, sys *ledsim.System) {
	// Randomly choose an led
	// Get that LED, and and neighbouring LEDS
	// That group of LEDs will be a cluster

	// slowly fade those down

	// Store those LEDS in array closed
	// Pick another random LED that is not in CLosed
	// Repeat until all LEDS in closed
	t := progress * float64(d.duration) / float64(time.Second)
	var period float64 = d.period

	if t > float64(period) {
		div := float64(int(t / float64(period)))
		t = t - (div * float64(period))
	}

	clustDissolved := true
	for _, clustLed := range d.currClust {
		if !d.state[clustLed].AlmostEqualRgb(d.reqColor) {
			clustDissolved = false
		}
	}

	undissolvedCount := len(sys.LEDs)
	for _, led := range sys.LEDs {

		// If this led is in the current cluster of LEDs we want to dissolve (or if only <100 remaining LEDs to dissolve)
		if contains(d.currClust, led.ID) /*|| d.remCount < 100*/ {
			// If this cluster is 'dissolved'
			if d.state[led.ID].AlmostEqualRgb(d.reqColor) {
				d.closed = append(d.closed, led.ID)
				remove(d.remaining, led.ID)

				led.Color = d.reqColor
				d.state[led.ID] = d.reqColor
			} else {
				led.Color = ledsim.BlendRgb(led.Color, d.reqColor, t/period)
				d.state[led.ID] = ledsim.BlendRgb(led.Color, d.reqColor, t/period)
			}

		} else {
			led.Color = d.state[led.ID]
		}

		if d.state[led.ID].AlmostEqualRgb(d.reqColor) {
			undissolvedCount--
		}

	}

	ledCount := len(sys.LEDs)
	// Once we have reached limit of period, start looking for new clusters
	if clustDissolved && math.Abs(t-period) > period-0.1 {
		fmt.Println("TRUE")
		d.currClust = nil
		for i := 1; i <= d.clustSize; i++ {
			var randLedIndx int
			if undissolvedCount == 0 {
				// fmt.Println("FINISHED")
				break
			} else if undissolvedCount < 100 {
				randLedIndx = d.remaining[0]
				remove(d.remaining, randLedIndx)

			} else {
				randLedIndx = rand.Intn(ledCount)
				for contains(d.closed, randLedIndx) {
					randLedIndx = rand.Intn(ledCount)
				}
			}

			d.currClust = append(d.currClust, randLedIndx)
			for _, nbr := range sys.LEDs[randLedIndx].Neighbours {
				d.currClust = append(d.currClust, nbr.ID)
			}
		}
		fmt.Println("Curr clust: ", len(d.currClust))
		fmt.Println("t: ", t)

	}

}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// func remove(s []int, i int) []int {
// 	s[i] = s[len(s)-1]
// 	return s[:len(s)-1]
// }

func remove(l []int, item int) []int {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func (d *Dissolve) OnExit(sys *ledsim.System) {

}

func AlmostEqualRgb(c1 colorful.Color, c2 colorful.Color) bool {
	return math.Abs(c1.R-c2.R)+
		math.Abs(c1.G-c2.G)+
		math.Abs(c1.B-c2.B) < 0.01
}

var _ ledsim.Effect = (*Dissolve)(nil)
