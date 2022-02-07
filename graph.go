package ledsim

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// thread safe struct
type undirectedGraph struct {
	vertices []*LED
	edges    map[*LED][]*LED
}

func newGraph() *undirectedGraph {
	return &undirectedGraph{
		vertices: make([]*LED, 0),
		edges:    make(map[*LED][]*LED),
	}
}

// func (g *undirectedGraph) getEdges() map[*LED][]*LED {
// 	return g.edges
// }

// func (g *undirectedGraph) getVertices() []*LED {
// 	return g.vertices
// }

func (g *undirectedGraph) addEdge(u *LED, v *LED) {
	if u == v { //same vertex
		return
		// g.edges[u] = append(g.edges[u], v)
	} else { // general case
		g.edges[u] = append(g.edges[u], v)
		g.edges[v] = append(g.edges[v], u)
	}
}

// move item to be removed to last on slice and then return slice without last element
func removeVertexFromSlice(index int, s []*LED) []*LED {
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

func getDistance(x0 float64, y0 float64, z0 float64, x1 float64, y1 float64, z1 float64) float64 {
	return math.Sqrt(math.Pow(x1-x0, 2) + math.Pow(y1-y0, 2) + math.Pow(z1-z0, 2))
}

func (g *undirectedGraph) getVertexByCoord(X float64, Y float64, Z float64) *LED {
	min_dist := 10000.0
	var curr_vertex *LED
	for _, v := range g.vertices {
		dist := getDistance(X, Y, Z, v.X, v.Y, v.Z)
		if dist < min_dist {
			min_dist = dist
			curr_vertex = v
		}
	}
	if curr_vertex == nil {
		panic("get vertex by coord failed")
	}
	return curr_vertex
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

// // basic representation function, {vertex} -> {connected} {to} {these}
// func (g *undirectedGraph) toString() {
// 	for key, value := range g.getEdges() {
// 		fmt.Printf("{%v} -> ", key.toString())
// 		for i := 0; i < len(value); i++ {
// 			fmt.Printf("%v ", *value[i])
// 		}
// 		fmt.Println()
// 	}
// }

//go:embed ledpos.txt
var ledposFile []byte

//go:embed mapping.txt
var mappingFile []byte

//go:embed teensy.txt
var teensyFile []byte

func (g *undirectedGraph) loadTeensys() (map[string]*Teensy, map[int]string) {
	ip := regexp.MustCompile(`(?:[0-9]{1,3}\.){3}[0-9]{1,3}`)
	teensyScanner := bufio.NewScanner(bytes.NewReader(teensyFile))

	teensys := make(map[string]*Teensy)
	chainToIpMap := make(map[int]string)
	var currentIp string
	var currentPin int
	var currentChainPosOnPin int
	for teensyScanner.Scan() {
		if ip.MatchString(teensyScanner.Text()) {
			newTeensy := Teensy{
				IP:     teensyScanner.Text(),
				Chains: make(map[int]*Chain, 0),
			}
			teensys[newTeensy.IP] = &newTeensy
			currentIp = newTeensy.IP
			currentPin = 0
		} else {
			chainIds := strings.Split(teensyScanner.Text(), ",")
			for _, chainId := range chainIds {
				isReversed := chainId[len(chainId)-1] == '\''
				var chainIdNum int
				if isReversed {
					chainIdNum, _ = strconv.Atoi(chainId[0 : len(chainId)-1])
				} else {
					chainIdNum, _ = strconv.Atoi(chainId)
				}
				teensys[currentIp].Chains[chainIdNum] = &Chain{Id: chainIdNum, Pin: currentPin, PosOnPin: currentChainPosOnPin, Length: 0, Reversed: isReversed}
				currentChainPosOnPin += 1
				chainToIpMap[chainIdNum] = currentIp
			}
			currentChainPosOnPin = 0
			currentPin += 1
		}
	}

	return teensys, chainToIpMap
}

// uses the ledpos and mapping text to build static graph
func (g *undirectedGraph) populateGraph(sys *System) {

	// aka chainId
	crack := regexp.MustCompile(`\s{12}\{\d*\}`)

	// Golang does not support lookaheads and so we will need to remove the ". {" later.
	coordPos := regexp.MustCompile(`\d*\.\s*\{`)
	coordSection := regexp.MustCompile(`\{-?\d*\.\d*,\s-?\d*\.\d*,\s-?\d*\.\d*\}`)
	coord := regexp.MustCompile(`-?\d*\.\d*`)

	vertexPair := regexp.MustCompile(`-?\d*\.\d*`)

	ledposScanner := bufio.NewScanner(bytes.NewReader(ledposFile))
	mappingScanner := bufio.NewScanner(bytes.NewReader(mappingFile))

	teensys, ledToIpMap := g.loadTeensys()
	sys.Teensys = teensys

	currLedRun := make([]*LED, 0)
	var chainIdNum int
	for ledposScanner.Scan() {
		if crack.MatchString(ledposScanner.Text()) {
			chainIdWithBrackets := strings.TrimSpace(ledposScanner.Text())
			chainIdNum, _ = strconv.Atoi(strings.Trim(chainIdWithBrackets, "{}"))

			// parse through list and make edges
			// make new list
			g.vertices = append(g.vertices, currLedRun...)
			for i := 1; i < len(currLedRun); i++ {
				g.addEdge(currLedRun[i-1], currLedRun[i])
			}
			currLedRun = make([]*LED, 0)
		} else if coordSection.MatchString(ledposScanner.Text()) {
			coordStr := coordSection.FindAllString(ledposScanner.Text(), 1)
			coords := coord.FindAllStringSubmatch(coordStr[0], 3)
			X, _ := strconv.ParseFloat(coords[0][0], 64)
			Y, _ := strconv.ParseFloat(coords[1][0], 64)
			Z, _ := strconv.ParseFloat(coords[2][0], 64)
			IP, _ := ledToIpMap[chainIdNum]
			coordPosId, _ := strconv.Atoi(strings.Split(coordPos.FindAllString(ledposScanner.Text(), 1)[0], ".")[0])

			led := &LED{
				X: X,
				Y: Y,
				Z: Z,
				PhysicalLEDPosition: PhysicalLEDPosition{
					TeensyIp:        IP,
					Chain:           chainIdNum,
					PositionOnChain: coordPosId,
				},
				RawLine: ledposScanner.Text(),
			}
			teensys[led.TeensyIp].Chains[chainIdNum].Length += 1
			sys.AddLED(led)

			currLedRun = append(currLedRun, led)
		} else {
			fmt.Fprintln(os.Stdout, "no coord found")
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
			g.addEdge(g.getVertexByCoord(X0, Y0, Z0), g.getVertexByCoord(X1, Y1, Z1))
		}
	}

	if err0 := ledposScanner.Err(); err0 != nil {
		log.Fatal(err0)
	} else if err1 := mappingScanner.Err(); err1 != nil {
		log.Fatal(err1)
	}

}
