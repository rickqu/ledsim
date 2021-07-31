package effects

import (
	"fmt"
	"ledsim"
	"math"
	"sort"
)

type Vector struct {
	X, Y, Z float64
}

func (v Vector) Mul(s float64) Vector {
	v.X *= s
	v.Y *= s
	v.Z *= s
	return v
}

func (v Vector) Add(a Vector) Vector {
	v.X += a.X
	v.Y += a.Y
	v.Z += a.Z
	return v
}

func (v Vector) Sub(a Vector) Vector {
	v.X -= a.X
	v.Y -= a.Y
	v.Z -= a.Z
	return v
}

func (v Vector) Sqrt() Vector {
	return Vector{X: math.Sqrt(v.X), Y: math.Sqrt(v.Y), Z: math.Sqrt(v.Z)}
}

func (v Vector) Squared() Vector {
	return Vector{X: v.X * v.X, Y: v.Y * v.Y, Z: v.Z * v.Z}
}

func (v Vector) Norm(a Vector) Vector {
	return Vector{
		X: a.Y*v.Z - a.Z*v.Y,
		Y: a.Z*v.X - a.X*v.Z,
		Z: a.X*v.Y - a.Y*v.X,
	}
}

func (v Vector) Dot(a Vector) float64 {
	return v.X*a.X + v.Y*a.Y + v.Z*a.Z
}

func (v Vector) Angle(a Vector) float64 {
	return math.Acos(v.Dot(a) / (v.Magnitude() * a.Magnitude()))
}

func (v Vector) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector) TransferRatio(a Vector) float64 {
	return math.Cos(v.Angle(a))
}

func (v Vector) Unit() Vector {
	return v.Mul(1.0 / v.Magnitude())
}

type Plane struct {
	Normal Vector
	Point  Vector
}

func (p *Plane) DistanceToPlane(point Vector) float64 {
	return (p.Normal.Dot(point) - p.Point.Dot(p.Normal)) / math.Sqrt(p.Normal.Dot(p.Normal))
}

type Inertia struct {
	LED        *ledsim.LED
	ForwardLED *ledsim.LED
	Velocity   float64 // in units per second
	Progress   float64
	Fluid      bool
	Gravity    float64
	Resistance float64
	Visited    map[*ledsim.LED]bool
}

func (n *Inertia) Evaluate(t float64) []*Inertia {
	if math.IsNaN(t) {
		panic("t is NaN")
	}

	if n.Velocity < 0 {
		n.Velocity += t
		if n.Velocity > 0 {
			n.Velocity = 0
		}
		return []*Inertia{n}
	}

	if n.Velocity < 0.01 {
		n.Velocity = 0.01
	}

	// s = ut + 0.5at^2
	// a = -0.3 units/s^2 on the z-axis
	// t = (1/a)(-u + sqrt(u^2 + 2as))

	vec := Vector{X: n.ForwardLED.X, Y: n.ForwardLED.Y, Z: n.ForwardLED.Z}.
		Sub(Vector{X: n.LED.X, Y: n.LED.Y, Z: n.LED.Z})

	gravity := Vector{X: 0, Y: 0, Z: -n.Gravity}
	a := gravity.TransferRatio(vec)*n.Gravity - (n.Velocity * n.Resistance)
	if a < 0 {
		a = 0
	}

	remainingDist := vec.Magnitude() * (1 - n.Progress)
	if remainingDist < 0 {
		fmt.Println("negative remianin")
	}

	remainingT := (1 / a) * (-n.Velocity +
		math.Sqrt((n.Velocity*n.Velocity)+2*a*remainingDist))
	if math.IsNaN(remainingT) {
		// assume 0 acceleration
		remainingT = remainingDist / n.Velocity
	}

	if remainingT < 0 {
		remainingT = 0
	}

	if remainingT > t {
		// it ends here
		dist := n.Velocity*t + 0.5*a*t*t
		n.Progress += dist / vec.Magnitude()
		if n.Progress >= 1 {
			fmt.Println("warninig: overflow??")
			n.Progress = 0.99
		}

		n.Velocity += a * t

		return []*Inertia{n}
	}

	t -= remainingT
	// transfer the velocity to the next object and re-evaluate

	n.Velocity += a * remainingT

	sort.Slice(n.ForwardLED.Neighbours, func(i, j int) bool {
		return n.ForwardLED.Neighbours[i].Z < n.ForwardLED.Neighbours[j].Z
	})

	var zOptions []float64
	for _, neighbour := range n.ForwardLED.Neighbours {
		zOptions = append(zOptions, neighbour.Z)
	}

	var next []*Inertia
	for _, neighbour := range n.ForwardLED.Neighbours {
		if n.Visited[neighbour] {
			continue
		}

		nextVec := Vector{X: neighbour.X, Y: neighbour.Y, Z: neighbour.Z}.
			Sub(Vector{X: n.ForwardLED.X, Y: n.ForwardLED.Y, Z: n.ForwardLED.Z})

		n.Visited[neighbour] = true

		next = append(next, &Inertia{
			LED:        n.ForwardLED,
			ForwardLED: neighbour,
			Velocity:   vec.TransferRatio(nextVec) * n.Velocity,
			Progress:   0,
			Fluid:      n.Fluid,
			Gravity:    n.Gravity,
			Resistance: n.Resistance,
			Visited:    n.Visited,
		})

		if !n.Fluid {
			break
		}
	}

	var nextNext []*Inertia
	for _, inertia := range next {
		if len(next) > 1 {
			inertia.Velocity = -0.5
		}
		nextNext = append(nextNext, inertia.Evaluate(t)...)
	}

	return nextNext
}
