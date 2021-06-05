// adapted from https://flaviocopes.com/golang-data-structure-graph/

package graph

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
)

type Graph interface {
}

// thread safe struct
type undirectedGraph struct {
	vertices[] *Vertex
	edges map[*Vertex][]*Vertex
	lock sync.RWMutex
}

func newGraph() *undirectedGraph {
	return &undirectedGraph{
		vertices: make([]*Vertex, 0),
		edges:    make(map[*Vertex][]*Vertex),
	}
}

func (g *undirectedGraph) getEdges() map[*Vertex][]*Vertex {
	return g.edges
}

func (g *undirectedGraph) getVertices() []*Vertex {
	return g.vertices
}

// set colour to white by default
func (g *undirectedGraph) addVertex (x float64, y float64, z float64){
	g.lock.Lock()
	g.vertices = append(g.vertices, &Vertex{x, y, z, 255, 255, 255})
	g.lock.Unlock()
}

func (g *undirectedGraph) addEdge(u *Vertex, v *Vertex) {
	g.lock.Lock()
	if u == v { //same vertex
		g.edges[u] = append(g.edges[u], v)
	} else { // general case
		g.edges[u] = append(g.edges[u], v)
		g.edges[v] = append(g.edges[v], u)
	}

	g.lock.Unlock()
}

// move item to be removed to last on slice and then return slice without last element
func removeVertexFromSlice(index int, s []*Vertex) []*Vertex {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}


// removal operations are slow (linear time) because since we have a set graph I assume we won't have to use them much
func (g *undirectedGraph) removeVertex(X float64, Y float64, Z float64) {
	for i, v := range g.vertices {
		if v.X == X && v.Y == Y && v.Z == Z {
			g.vertices = removeVertexFromSlice(i, g.vertices)
			delete(g.edges, v)
		}
	}
}

// get vertex from the vertex list by coordinate as 3d coordinates are unique
func (g *undirectedGraph) getVertexByCoord(X float64, Y float64, Z float64) *Vertex {
	for _, v := range g.vertices {
		if v.X == X && v.Y == Y && v.Z == Z {
			return v
		}
	}
	return nil
}

func (g *undirectedGraph) removeEdge(x0 float64, y0 float64, z0 float64, x1 float64, y1 float64, z1 float64) {
	u := g.getVertexByCoord(x0, y0, z0)
	for i, v := range g.edges[u] {
		if v.X == x1 && v.Y == y1 && v.Z == z1 {
			g.edges[u] = removeVertexFromSlice(i, g.edges[u])
		}
	}
	// remove entire list if there are no edges in it
	if len(g.edges[u]) == 0 {
		delete(g.edges, u)
	}

}

// basic representation function, {vertex} -> {connected} {to} {these}
func (g *undirectedGraph) toString() {
	g.lock.RLock()
	for key, value := range g.getEdges() {
		fmt.Printf("{%v} -> ", key.toString())
		for i := 0; i < len(value); i++ {
			fmt.Printf("%v ", *value[i])
		}
		fmt.Println()
	}
	g.lock.RUnlock()
}

// uses the ledpos and mapping text to build static graph
func (g *undirectedGraph) populateGraph() {
	crack := regexp.MustCompile("\\s{12}\\{\\d*\\}")
	coordSection := regexp.MustCompile("\\{-?\\d*\\.\\d*,\\s-?\\d*\\.\\d*,\\s-?\\d*\\.\\d*\\}")
	coord := regexp.MustCompile("-?\\d*\\.\\d*")

	vertexPair := regexp.MustCompile("-?\\d*\\.\\d*")

	file0, err0 := os.Open("input/ledpos.txt")
	file1, err1 := os.Open("input/mapping.txt")

	if err0 != nil {
		log.Fatal(err0)
	} else if err1 != nil {
		log.Fatal(err1)
	}
	defer file0.Close()
	defer file1.Close()

	ledposScanner := bufio.NewScanner(file0)
	mappingScanner := bufio.NewScanner(file1)

	currLedRun := make([]*Vertex, 0)
	for ledposScanner.Scan() {
		if crack.MatchString(ledposScanner.Text()) {
			// parse through list and make edges
			// make new list
			g.vertices = append(g.vertices, currLedRun...)
			for i := 1; i < len(currLedRun); i++ {
				g.addEdge(currLedRun[i -1], currLedRun[i])
			}
			currLedRun = make([]*Vertex, 0)
		} else {
			if coordSection.MatchString(ledposScanner.Text()) {
				coordStr := coordSection.FindAllString(ledposScanner.Text(), 1)
				coords := coord.FindAllStringSubmatch(coordStr[0], 3)
				X, _ := strconv.ParseFloat(coords[0][0], 64)
				Y, _ := strconv.ParseFloat(coords[1][0], 64)
				Z, _ := strconv.ParseFloat(coords[2][0], 64)
				//fmt.Fprintf(os.Stdout, "%f, %f, %f\n", X, Y, Z)
				currLedRun = append(currLedRun, &Vertex{X, Y, Z, 255, 255, 255})
			} else {
				fmt.Fprintln(os.Stdout, "no coord found")
			}
		}
	}

	for mappingScanner.Scan() {
		if vertexPair.MatchString(mappingScanner.Text()) {
			pairs := vertexPair.FindAllStringSubmatch(mappingScanner.Text(), 6)
			X0, _ := strconv.ParseFloat(pairs[0][0], 64)
			Y0, _ := strconv.ParseFloat(pairs[1][0], 64)
			Z0, _ := strconv.ParseFloat(pairs[2][0], 64)
			X1, _ := strconv.ParseFloat(pairs[3][0], 64)
			Y1, _ := strconv.ParseFloat(pairs[4][0], 64)
			Z1, _ := strconv.ParseFloat(pairs[5][0], 64)
			u := &Vertex{X0, Y0, Z0, 255, 255, 255}
			v := &Vertex{X1, Y1, Z1, 255, 255, 255}
			g.addEdge(u, v)
		}
	}


	if err0 := ledposScanner.Err(); err0 != nil {
		log.Fatal(err0)
	} else if err1 := mappingScanner.Err(); err1 != nil {
		log.Fatal(err1)
	}
}