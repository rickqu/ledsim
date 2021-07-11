package effects

import (
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

// type Bead struct {
// 	Position *ledsim.LED
// }

type ShootingStar struct {
	startPoint Vector
	endPoint   Vector
	i          Vector
	j          Vector
	i2         float64
	j2         float64
	p1         Vector
	normal     Vector
	vertical   Vector
}

// type Inertia struct {
// 	LED        *ledsim.LED
// 	ForwardLED *ledsim.LED
// 	Velocity   float64 // in units per second
// 	Progress   float64
// 	Fluid      bool
// 	Gravity    float64
// }

// speed is in LEDs per second
func NewShootingStar(start Vector, end Vector) *ShootingStar {
	return &ShootingStar{
		startPoint: start,
		endPoint:   end,
	}
}

func (b *ShootingStar) OnEnter(sys *ledsim.System) {
	// project system onto plane with normal of shooting star and the ground.
	vec := b.endPoint.Sub(b.startPoint)
	b.normal = vec.Norm(Vector{0, 0, -1})
	// extrude the starting point out to create a cuboid
	ext1 := b.startPoint.Add(b.normal.Mul(100))
	ext2 := b.startPoint.Add(b.normal.Mul(-100))
	// ext3 := b.endPoint.Add(b.normal.Mul(100))
	// ext4 := b.endPoint.Add(b.normal.Mul(-100))
	b.vertical = vec.Norm(b.normal).Unit()
	p1 := ext1.Add(b.vertical.Mul(0.01))
	p2 := ext1.Add(b.vertical.Mul(-0.01))
	// p3 := ext2.Add(b.vertical.Mul(-0.01))
	p4 := ext2.Add(b.vertical.Mul(0.01))
	// p5 := ext3.Add(b.vertical.Mul(0.01))
	// p6 := ext3.Add(b.vertical.Mul(-0.01))
	// p7 := ext4.Add(b.vertical.Mul(-0.01))
	// p8 := ext4.Add(b.vertical.Mul(0.01))

	b.i = p2.Sub(p1)
	b.j = p4.Sub(p1)
	// b.k = p5.Sub(p1)
	b.i2 = b.i.Dot(b.i)
	b.j2 = b.j.Dot(b.j)
	// b.k2 = b.k.Dot(b.k)
	b.p1 = p1

	// project every LED onto the plane
	// for _, led := range sys.LEDs {
	// 	v := Vector{
	// 		X: led.X,
	// 		Y: led.Y,
	// 		Z: led.Z,
	// 	}.Sub(p1)
	// 	vi := v.Dot(i)
	// 	vj := v.Dot(j)
	// 	vk := v.Dot(k)

	// 	if 0 < vi && vi < i2 && 0 < vj && vj < j2 && 0 < vk && vk < k2 {
	// 		led.Color = colorful.Color{0, 1, 0}
	// 	}
	// }
}

func (b *ShootingStar) OnExit(sys *ledsim.System) {
}

func (b *ShootingStar) Eval(progress float64, sys *ledsim.System) {
	vec := b.endPoint.Sub(b.startPoint)

	ext3 := b.startPoint.Add(vec.Mul(progress)).Add(b.normal.Mul(100))
	p5 := ext3.Add(b.vertical.Mul(0.01))
	k := p5.Sub(b.p1)
	k2 := k.Dot(k)

	for _, led := range sys.LEDs {
		v := Vector{
			X: led.X,
			Y: led.Y,
			Z: led.Z,
		}.Sub(b.p1)
		vi := v.Dot(b.i)
		vj := v.Dot(b.j)
		vk := v.Dot(k)

		if 0 < vi && vi < b.i2 && 0 < vj && vj < b.j2 && 0 < vk && vk < k2 {
			led.Color = colorful.Color{0, 1, 0}
		}
	}
}

var _ ledsim.Effect = (*ShootingStar)(nil)
