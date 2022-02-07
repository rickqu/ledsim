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

const TARGET_PORT = 5151
const SERVER_PORT = 900

type TeensyNetwork struct {
	outputConn *net.UDPConn
	binConns   *sync.Map
	connsMutex *sync.Mutex
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
		Port: SERVER_PORT,
	})
	if err != nil {
		panic("Cannot start UDP server: " + err.Error())
	}

	network := &TeensyNetwork{
		binConns:   new(sync.Map),
		connsMutex: new(sync.Mutex),
		outputConn: outputConnection,
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

		// we use RGB (3 bytes) for each LED. Each Teensy is aware of how long the chains are.
		outputArray := make([]byte, lenPacket*3)

		network.binConns.Store(ip, &udpOutput{
			outputAddr: &net.UDPAddr{IP: net.IPv4(ipArr[0], ipArr[1], ipArr[2], ipArr[3]), Port: TARGET_PORT},
			outputBuff: outputArray})
	}
	mapLedToOutputArray(sys, network)
	return network
}

func mapLedToOutputArray(sys *ledsim.System, teensyNetwork *TeensyNetwork) {
	for _, led := range sys.LEDs {
		teensy := sys.Teensys[led.TeensyIp]
		chain := teensy.Chains[led.Chain]

		ledsBeforeTarget := 0
		for _, specificChain := range teensy.Chains {
			if specificChain.Pin < chain.Pin {
				ledsBeforeTarget += specificChain.Length
			} else if specificChain.Pin == chain.Pin && specificChain.PosOnPin < chain.PosOnPin {
				ledsBeforeTarget += specificChain.Length
			}
		}
		if chain.Reversed {
			ledsBeforeTarget += chain.Length - (led.PositionOnChain + 1)
		} else {
			ledsBeforeTarget += led.PositionOnChain
		}
		outputArrayFromMap, _ := teensyNetwork.binConns.Load(led.TeensyIp)
		outputArray := outputArrayFromMap.(*udpOutput).outputBuff

		led.Red = &outputArray[ledsBeforeTarget*3]
		led.Green = &outputArray[ledsBeforeTarget*3+1]
		led.Blue = &outputArray[ledsBeforeTarget*3+2]

		// uncomment the following line to see which slots in the output array an
		// LED maps to.
		// if *led.Red != 0 {
		// 	panic("We've already visited this array slot!")
		// }
		// *led.Red = 255
		// *led.Green = 255
		// *led.Blue = 255
	}
	// debugging view what the output buffer looks like.
	// test, _ := teensyNetwork.binConns.Load("10.1.2.1")
	// testRead := test.(*udpOutput).outputBuff
	// print(testRead)
}
