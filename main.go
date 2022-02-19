package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"ledsim/effects"
	"ledsim/internal"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/1lann/dissonance/ffmpeg"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65535,
	WriteBufferSize: 65535,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	LEDs []*internal.LED `json:"leds"`
}

var (
	scale  = 0.0005 * 2.056422
	origin = [...]float64{
		(1 / (0.0005 * 2.056422)) * -18.04,
		(1 / (0.0005 * 2.056422)) * 9.58,
		(1 / (0.0005 * 2.056422)) * 1,
	}
)

var pattern = regexp.MustCompile(`(?m)^.*{([-\.0-9]+), ([-\.0-9]+), ([-\.0-9]+)}\s*$`)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify the name of the input device you want to use in quotes as the second argument.")
		fmt.Println("Available devices:")
		devices, err := ffmpeg.GetDshowDevices()
		if err != nil {
			panic(err)
		}

		for _, dev := range devices {
			fmt.Println(dev)
		}

		return
	}

	// For writing bytes to a file
	//f, err := os.Create("bytes_data")
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()

	// For reading bytes to a file
	// # REPLAY CODE
	f, err := os.Open("rgb_bytes")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := ioutil.ReadFile("./resources/crack_leds.txt")
	if err != nil {
		panic(err)
	}

	effects := []internal.Effect{
		//effects.NewVolumeAdjust(os.Args[1], false),
		&effects.DiagonalRainbow{},
	}

	// make a ring of 80 LEDs with radius 1m
	sys := &internal.System{
		NormalizeOnce: new(sync.Once),
	}

	groups := pattern.FindAllStringSubmatch(string(data), -1)
	for _, group := range groups {
		x, _ := strconv.ParseFloat(group[1], 64)
		y, _ := strconv.ParseFloat(group[2], 64)
		z, _ := strconv.ParseFloat(group[3], 64)

		sys.AddLED(&internal.LED{
			X: -(x - origin[0]) * scale,
			Y: (y - origin[1]) * scale,
			Z: (z - origin[2]) * scale,
		})
	}

	conns := new(sync.Map)
	binConns := new(sync.Map)
	connMutex := new(sync.Mutex)

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 5050,
	})
	if err != nil {
		panic(err)
	}

	targetUDPAddr := &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 100),
		Port: 8888,
	}

	go func() {
		out := make([]byte, 1600)
		for {
			_, _, err := udpConn.ReadFromUDP(out)
			// fmt.Println("received packet with size:", n)
			if err != nil {
				// close(recvChan)
				panic(err)
			}
		}
	}()

	sys.AfterFrame(func(s *internal.System, t time.Time) {
		// msg := &Message{
		// 	LEDs: s.LEDs,
		// }
		out := new(bytes.Buffer)
		for i, led := range s.LEDs {
			out.WriteString(strconv.FormatInt(int64(led.RGB()), 10))
			if i != len(s.LEDs)-1 {
				out.WriteByte(',')
			}
		}

		binOut := new(bytes.Buffer)

		bytes6300 := make([]byte, 6300) // array of 6300 bytes being read in from file # REPLAY CODE

		led_count := len(bytes6300) / 3 // 2100 # REPLAY CODE

		_, err = f.Read(bytes6300) // # REPLAY CODE

		//for _, led := range s.LEDs {
		for i := 0; i < led_count; i++ {

			//byt = append(byt, led.R)
			//byt = append(byt, led.G)
			//byt = append(byt, led.B)
			R := bytes6300[i*3]   // # REPLAY CODE
			G := bytes6300[i*3+1] // # REPLAY CODE
			B := bytes6300[i*3+2] // # REPLAY CODE

			//fmt.Printf("R: %d G: %d B: %d\n", R, G, B)
			binOut.Write([]byte{R, G, B})
		}
		//f.Write(byt)
		bytes6300 = nil

		go func() {
			connMutex.Lock()
			defer connMutex.Unlock()

			conns.Range(func(key, value interface{}) bool {
				conn := value.(*websocket.Conn)
				conn.WriteMessage(websocket.TextMessage, out.Bytes())
				return true
			})

			binConns.Range(func(key, value interface{}) bool {
				conn := value.(*websocket.Conn)
				conn.WriteMessage(websocket.BinaryMessage, binOut.Bytes())
				return true
			})

			_, err := udpConn.WriteToUDP(binOut.Bytes()[:300*3], targetUDPAddr)
			if err != nil {
				fmt.Println("error during write:", err)
			}
		}()
	})

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Static("/", "./resources/index.html")
	e.Static("/script.js", "./resources/script.js")

	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		c.Response().Committed = true

		key := conn.RemoteAddr().String()
		conns.Store(key, conn)
		defer conns.Delete(key)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
		}
	})

	e.GET("/wsbin", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		c.Response().Committed = true

		key := conn.RemoteAddr().String()
		binConns.Store(key, conn)
		defer binConns.Delete(key)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
		}
	})

	go sys.Run(effects)

	log.Fatalln(e.Start(":9000"))
}
