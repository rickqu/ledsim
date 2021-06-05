package graph

import (
	"fmt"
	_ "fmt"
	"os"
	"testing"
)

var g *undirectedGraph
var h *undirectedGraph

func TestInitialise(t *testing.T) {
	g = newGraph()
	h = newGraph()
}

func fillGraph() {
	g.addVertex(1, 1, 1)
	g.addVertex(1 , 0 , 1)
	g.addVertex(2, 1, 0)
	g.addVertex(1, 2, 3)
	g.addVertex(2,3 ,4 )
	g.addVertex(4, 5, 6)

	a := g.getVertexByCoord(1, 1, 1)
	b := g.getVertexByCoord(1, 0, 1)
	c := g.getVertexByCoord(2, 1, 0)
	d := g.getVertexByCoord(1, 2,3)
	e := g.getVertexByCoord(2, 3, 4)
	f := g.getVertexByCoord(4, 5, 6)

	g.addEdge(a, b)
	g.addEdge(b, c)
	g.addEdge(a, c)
	g.addEdge(c, f)
	g.addEdge(e, f)
	g.addEdge(d, f)
}

func TestAdd(t *testing.T) {
	fillGraph()
	g.toString()
}

func TestRemoveVertex (t *testing.T) {
	fmt.Fprintln(os.Stdout, "Test vertex removal")
	g.removeVertex(1, 1, 1)
	g.toString()
}

func TestRemoveEdge (t *testing.T) {
	fmt.Fprintln(os.Stdout, "Testing Edge removal")
	g.removeEdge(1, 0, 1, 1,1 ,1)
	g.toString()
}

func TestVertices(t *testing.T) {
	_, _ = fmt.Fprintln(os.Stdout, "test vertices")
	for _, v := range g.getVertices() {
		fmt.Println(v.toString())
	}
}

func TestPopulateGraph(t *testing.T) {
	h.populateGraph()
	h.toString()
}

