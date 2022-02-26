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
	// ledPeriods []time.Duration
	// delay      []time.Duration
	duration  time.Duration
	closed    []int          // index of which leds have dissolved
	currClust []int          // index of current cluster of LEDs we want to dissolve
	reqColor  colorful.Color // required color
	clustSize int            // How many clusters/groups of LEDs to dissolve at a time

	// period 	   time.Duration
	// baseline   time.Duration
	// deviation  time.Duration
	// EaseFunc func(x float64) float64
}

func NewDissolve(duration time.Duration, reqColor colorful.Color) *Dissolve {
	return &Dissolve{
		duration:  duration,
		reqColor:  reqColor,
		clustSize: 60,
	}
}

// func (d *Dissolve) DFS() {

// }

func (d *Dissolve) OnEnter(sys *ledsim.System) {
	ledCount := len(sys.LEDs)
	d.closed = make([]int, ledCount)
	d.currClust = make([]int, 0)
	// d.reqColor = colorful.Color{0.7, 0.3, 0.2}

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
	t := progress * float64(d.duration) / float64(time.Second)
	var period float64 = 0.1

	if t > float64(period) {
		div := float64(int(t / float64(period)))
		t = t - (div * float64(period))
	}

	// Randomly choose an led
	// Get that LED, and and neighbouring LEDS
	// That group of LEDs will be a cluster

	// slowly fade those down

	// Store those LEDS in array closed
	// Pick another random LED that is not in CLosed
	// Repeat until all LEDS in closed

	clustDissolved := false
	// for _, ledIndx := range d.currClust {

	// 	if sys.LEDs[ledIndx].Color.AlmostEqualRgb(d.reqColor) || (math.Abs(t-period) <= 0.05) {
	// 		clustDissolved = true
	// 		fmt.Println("CLUST DISSOLVED")
	// 		d.closed = append(d.closed, d.currClust...)

	// 		// break
	// 	}

	// 	sys.LEDs[ledIndx].Color = ledsim.BlendRgb(sys.LEDs[ledIndx].Color, colorful.Color{0.7, 0.3, 0.2}, t/period)

	// }

	for _, led := range sys.LEDs {
		if contains(d.currClust, led.ID) {
			if led.Color.AlmostEqualRgb(d.reqColor) || math.Abs(t-period) < 0.05 {
				clustDissolved = true
				// fmt.Println("CLUST DISSOLVED")
				d.closed = append(d.closed, led.ID)
			} else {
				led.Color = ledsim.BlendRgb(led.Color, d.reqColor, t/period)
			}

		}

		if contains(d.closed, led.ID) {
			led.Color = d.reqColor
		}

	}

	ledCount := len(sys.LEDs)
	if clustDissolved {
		// d.closed = append(d.closed, d.currClust...)
		d.currClust = nil
		for i := 1; i <= d.clustSize; i++ {
			var randLedIndx int
			if math.Abs(float64(ledCount-len(d.closed))) <= 10 {
				fmt.Println("Entered")
				for j := 0; j < ledCount; j++ {
					fmt.Println("Indx:", j)
					if !contains(d.closed, j) {
						randLedIndx = j
						break
					}
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

func (d *Dissolve) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*Dissolve)(nil)
