package player

import "github.com/gorilla/websocket"

type ClientMessenger interface {
	SendMessage(msgType string, data any, target *websocket.Conn)
	Event(msgType string, data any)
}

type Module struct {
	Conn      *websocket.Conn
	Type      string
	messenger ClientMessenger
}

func New(messenger ClientMessenger) *Module {
	return &Module{messenger: messenger}
}

func (m *Module) SetConn(conn *websocket.Conn) {
	m.Conn = conn
}

func (m *Module) PlayPause() {
	m.messenger.SendMessage("command", map[string]string{"action": "playPause"}, m.Conn)
}

func (m *Module) Previous() {
	m.messenger.SendMessage("command", map[string]string{"action": "previous"}, m.Conn)
}

func (m *Module) Next() {
	m.messenger.SendMessage("command", map[string]string{"action": "next"}, m.Conn)
}
