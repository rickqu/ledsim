package outputs

import (
	"ledsim"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

type udpOutput struct {
	outputAddr *net.UDPAddr
	outputBuff []byte
}

type TeensyNetwork struct {
	outputConn *net.UDPConn
	binConns   *sync.Map
	connsMutex *sync.Mutex
	serverPort int // probably already stored in outputConn but I don't know how to get it out
}

func (t *TeensyNetwork) Display(sys *ledsim.System) {
	for _, led := range sys.LEDs {
		r, g, b := led.Color.RGB255()
		*led.Red = r
		*led.Green = g
		*led.Blue = b
	}

	go func() {
		t.connsMutex.Lock()
		defer t.connsMutex.Unlock()

		// put write code here
		t.binConns.Range(func(key, value interface{}) bool {
			udpConnection := value.(*udpOutput)
			_, err := t.outputConn.WriteToUDP(udpConnection.outputBuff, udpConnection.outputAddr)

			if err != nil {
				panic("UDP write error to " + key.(string) + ": " + err.Error())
			}
			return true
		})
	}()
}

func NewTeensyNetwork(e *echo.Echo, sys *ledsim.System) *TeensyNetwork {

	outputConnection, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(10, 1, 2, 1),
		Port: 300,
	})
	if err != nil {
		panic("Cannot start UDP server: " + err.Error())
	}

	network := &TeensyNetwork{
		binConns:   new(sync.Map),
		connsMutex: new(sync.Mutex),
		outputConn: outputConnection,
		serverPort: 300,
	}

	for ip, teensy := range sys.Teensys {
		pins := make(map[int]int)
		lenPacket := 0
		for _, chain := range teensy.Chains {
			pins[chain.Pin] += chain.Length
			lenPacket += chain.Length
		}
		var ipArr []byte
		for _, substr := range strings.Split(ip, ".") {
			convResult, _ := strconv.Atoi(substr)
			ipArr = append(ipArr, byte(convResult))
		}

		// we use RGB (3 bytes) for each LED.
		network.binConns.Store(ip, udpOutput{
			outputAddr: &net.UDPAddr{IP: net.IPv4(ipArr[0], ipArr[1], ipArr[2], ipArr[3]), Port: network.serverPort},
			outputBuff: make([]byte, lenPacket*3)})
	}
	mapLedToOutputArray(sys, network)
	return network
}

func mapLedToOutputArray(sys *ledsim.System, teensyNetwork *TeensyNetwork) {
	for _, led := range sys.LEDs {
		teensy := sys.Teensys[led.TeensyIp]
		chain := teensy.Chains[led.Chain]

		ledsBeforeTarget := 0
		for i := 0; i < chain.Pin; i++ {
			for _, chain := range teensy.Chains {
				if chain.Pin < i {
					ledsBeforeTarget += chain.Length
				}
			}
		}
		if chain.Reversed {
			ledsBeforeTarget += chain.Length - (led.PositionOnChain + 1)
		} else {
			ledsBeforeTarget += led.PositionOnChain
		}
		outputArray, _ := teensyNetwork.binConns.Load(led.TeensyIp)
		led.Red = &outputArray.([]byte)[ledsBeforeTarget*3]
		led.Green = &outputArray.([]byte)[ledsBeforeTarget*3+1]
		led.Blue = &outputArray.([]byte)[ledsBeforeTarget*3+2]
	}
}
