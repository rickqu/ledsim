package outputs

import (
	"bytes"
	"ledsim"
	"log"
	"net"
)

type UDP struct {
	conn        *net.UDPConn
	sendMapping map[*net.UDPAddr][]int
}

func (u *UDP) Display(sys *ledsim.System) {
	binOut := new(bytes.Buffer)
	for _, led := range sys.LEDs {
		r, g, b := led.Color.RGB255()
		binOut.Write([]byte{r, g, b})
	}

	buf := binOut.Bytes()

	for target, mapping := range u.sendMapping {
		sendBuf := make([]byte, len(buf)*3)

		for i, led := range mapping {
			copy(sendBuf[i*3:], buf[led*3:led*3+3])
		}
		_, err := u.conn.WriteToUDP(sendBuf, target)
		if err != nil {
			log.Printf("ledsim/outputs/udp: error during write to %q: %v", target.String(), err)
		}
	}
}

func NewUDP(listenAddr string, sendMapping map[string][]int) (*UDP, error) {
	listen, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", listen)
	if err != nil {
		return nil, err
	}

	finishedMapping := make(map[*net.UDPAddr][]int)

	for targetAddr, mapping := range sendMapping {
		target, err := net.ResolveUDPAddr("udp", targetAddr)
		if err != nil {
			return nil, err
		}

		finishedMapping[target] = mapping
	}

	go func() {
		out := make([]byte, 1600)
		for {
			_, _, err := udpConn.ReadFromUDP(out)
			// fmt.Println("received packet with size:", n)
			if err != nil {
				panic(err)
			}
		}
	}()

	return &UDP{
		conn:        udpConn,
		sendMapping: finishedMapping,
	}, nil
}
