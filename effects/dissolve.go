package effects

import (
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

	for _, led := range sys.LEDs {
		d.remaining = append(d.remaining, led.ID)
	}

	// Pick random led and add its neighbours to form a starting cluster
	for i := 1; i <= d.clustSize; i++ {
		randLedIndx := rand.Intn(ledCount)
		d.currClust = append(d.currClust, randLedIndx)
		for _, nbr := range sys.LEDs[randLedIndx].Neighbours {
			d.currClust = append(d.currClust, nbr.ID)
		}
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

	clustDissolved := false

	for _, led := range sys.LEDs {

		// If this led is in the current cluster of LEDs we want to dissolve (or if only <100 remaining LEDs to dissolve)
		if contains(d.currClust, led.ID) || d.remCount < 100 {
			// If this cluster is 'dissolved'
			if led.Color.AlmostEqualRgb(d.reqColor) || math.Abs(t-period) < 0.05 {
				clustDissolved = true
				// fmt.Println("CLUST DISSOLVED")
				d.closed = append(d.closed, led.ID)
				remove(d.remaining, led.ID)
				if d.remCount > 0 {
					d.remCount--
					// fmt.Println("remcount is: ", d.remCount)
				}
			} else {
				led.Color = ledsim.BlendRgb(led.Color, d.reqColor, t/period)
			}

		}

		// If a cluster is already 'dissolved', keep it at the required colour (in most cases {0,0,0})
		if contains(d.closed, led.ID) {
			led.Color = d.reqColor
		}

	}

	ledCount := len(sys.LEDs)
	// If current cluster is dissolved, look for new cluster
	if clustDissolved {
		d.currClust = nil
		for i := 1; i <= d.clustSize; i++ {
			var randLedIndx int

			if d.remCount == 0 {
				// fmt.Println("remCount is 0")
				break
			} else if d.remCount < 100 {
				randLedIndx = d.remaining[0]
				remove(d.remaining, randLedIndx)
				if d.remCount > 0 {
					d.remCount--
				}
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

var _ ledsim.Effect = (*Dissolve)(nil)
