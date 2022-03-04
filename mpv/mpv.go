package mpv

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

type ipcResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

type ipcCommand struct {
	Command []interface{} `json:"command"`
}

// Player represents a MPV player.
type Player struct {
	conn          net.Conn
	eventHandlers map[string]func()
	commandQueue  chan chan<- *ipcResponse
	cmd           *exec.Cmd
}

// NewPlayer returns a new player with the given path.
func NewPlayer(pathToFile string, arg string, debug bool) (*Player, error) {
	cmd, err := runMPV(pathToFile, arg, debug)
	if err != nil {
		return nil, err
	}

	stoppedChan := make(chan error, 1)
	go func() {
		stoppedChan <- cmd.Wait()
	}()

	time.Sleep(3 * time.Second)

	select {
	case err := <-stoppedChan:
		return nil, fmt.Errorf("mpv: failed to start mpv: %w", err)
	default:
	}

	conn, err := connect()
	if err != nil {
		return nil, err
	}

	p := &Player{
		conn:          conn,
		eventHandlers: make(map[string]func()),
		commandQueue:  make(chan chan<- *ipcResponse, 5),
		cmd:           cmd,
	}

	go p.processIPC()

	return p, nil
}

func (p *Player) processIPC() error {
	defer p.conn.Close()

	rd := bufio.NewReader(p.conn)

	for {
		data, err := rd.ReadBytes('\n')
		if err != nil {
			return err
		}

		var resp ipcResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			log.Println("mpv: json unmarshal error:", err)
			continue
		}

		if resp.Event != "" {
			if handler, found := p.eventHandlers[resp.Event]; found {
				handler()
			}
			continue
		}

		(<-p.commandQueue) <- &resp
	}
}

// Command executes a command.
func (p *Player) Command(ctx context.Context, args ...interface{}) (interface{}, error) {
	data, err := json.Marshal(ipcCommand{args})
	if err != nil {
		return nil, err
	}

	recv := make(chan *ipcResponse, 1)
	p.commandQueue <- recv

	_, err = p.conn.Write(append(data, '\n'))
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-recv:
		if resp.Error != "" && resp.Error != "success" {
			return nil, errors.New("mpv: command error: " + resp.Error)
		}

		return resp.Data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CommandString executes a command and expects a string response.
func (p *Player) CommandString(ctx context.Context, args ...interface{}) (string, error) {
	val, err := p.Command(ctx, args...)
	if err != nil {
		return "", err
	}

	str, ok := val.(string)
	if !ok {
		return "", errors.New("mpv: unexpected type")
	}

	return str, nil
}

// CommandFloat64 executes a command and expects a float64 response.
func (p *Player) CommandFloat64(ctx context.Context, args ...interface{}) (float64, error) {
	val, err := p.Command(ctx, args...)
	if err != nil {
		return 0, err
	}

	f, ok := val.(float64)
	if !ok {
		return 0, errors.New("mpv: unexpected type")
	}

	return f, nil
}

// CommandBool executes a command and expects a bool response.
func (p *Player) CommandBool(ctx context.Context, args ...interface{}) (bool, error) {
	val, err := p.Command(ctx, args...)
	if err != nil {
		return false, err
	}

	b, ok := val.(bool)
	if !ok {
		return false, errors.New("mpv: unexpected type")
	}

	return b, nil
}

// Close closes a connection.
func (p *Player) Close() error {
	p.conn.Close()
	return p.cmd.Process.Kill()
}

func (p *Player) Play() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := p.Command(ctx, "set_property", "pause", false)
	return err
}

func (p *Player) Pause() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := p.Command(ctx, "set_property", "pause", true)
	return err
}

func (p *Player) SeekTo(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := p.Command(ctx, "set_property", "playback-time", t.Seconds())
	return err
}

func (p *Player) GetTimestamp() (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	current, err := p.CommandFloat64(ctx, "get_property", "playback-time")
	if err != nil {
		return 0, err
	}

	return time.Duration(current*1000.0) * time.Millisecond, nil
}

// GetPlayerState returns the state of the player.
// func (p *Player) GetPlayerState() (*aru.PlayerState, error) {
// 	current, _ := p.CommandFloat64("get_property", "playback-time")
// 	duration, _ := p.CommandFloat64("get_property", "duration")
// 	mediaPath, _ := p.CommandString("get_property", "path")
// 	coreIdle, _ := p.CommandBool("get_property", "core-idle")
// 	idleActive, err := p.CommandBool("get_property", "idle-active")
// 	if err != nil {
// 		return nil, err
// 	}

// 	state := aru.StatePlaying
// 	if idleActive {
// 		state = aru.StateStopped
// 	} else if coreIdle {
// 		state = aru.StatePaused
// 	}

// 	un, err := url.PathUnescape(mediaPath)
// 	if err == nil {
// 		mediaPath = un
// 	}

// 	return &aru.PlayerState{
// 		Path: mediaPath,
// 		Start: time.Now().Add(-1 * time.Duration(current*1000) *
// 			time.Millisecond),
// 		Current:  time.Duration(current*1000) * time.Millisecond,
// 		Duration: time.Duration(duration*1000) * time.Millisecond,
// 		State:    state,
// 	}, nil
// }
