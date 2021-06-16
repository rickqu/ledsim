package outputs

import (
	"bytes"
	"ledsim"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Mirage struct {
	binConns   *sync.Map
	connsMutex *sync.Mutex
}

func (m *Mirage) Display(sys *ledsim.System) {
	binOut := new(bytes.Buffer)
	for _, led := range sys.LEDs {
		r, g, b := led.Color.RGB255()
		binOut.Write([]byte{r, g, b})
	}

	go func() {
		m.connsMutex.Lock()
		defer m.connsMutex.Unlock()

		m.binConns.Range(func(key, value interface{}) bool {
			conn := value.(*websocket.Conn)
			conn.WriteMessage(websocket.BinaryMessage, binOut.Bytes())
			return true
		})
	}()
}

func NewMirage(e *echo.Echo) *Mirage {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  65535,
		WriteBufferSize: 65535,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	m := &Mirage{
		binConns:   new(sync.Map),
		connsMutex: new(sync.Mutex),
	}

	e.GET("/wsbin", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		c.Response().Committed = true

		key := conn.RemoteAddr().String()
		m.binConns.Store(key, conn)
		defer m.binConns.Delete(key)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
		}
	})

	return m
}
