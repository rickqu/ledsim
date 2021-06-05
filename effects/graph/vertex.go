package graph

import "fmt"

type Vertex struct {
	X float64
	Y float64
	Z float64
	R uint8
	G uint8
	B uint8
}

func (v *Vertex) toString() string {
	return fmt.Sprintf("C: %f %f %f, RGB: %d, %d, %d", v.X, v.Y, v.Z, v.R, v.G, v.B)
}

func newVertex(X float64, Y float64, Z float64) *Vertex {
	return &Vertex{X, Y, Z, 255, 255, 255}
}

