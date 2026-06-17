package websocketHandler

import (
	"encoding/json"
	"log"

	"github.com/Ignaciotema/YTMusic-web-socket/internal/dispatcher"
	"github.com/Ignaciotema/YTMusic-web-socket/internal/player"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ConnType int

const ()

type Client struct {
	Id   string
	Conn *websocket.Conn
}

type IncomingMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Id:   uuid.New().String(),
		Conn: conn,
	}
}

type ConnectionManager struct {
	PlayerConn      *websocket.Conn
	playerModule    *player.Module
	controllerConns []*Client
	dispatcher      *dispatcher.Dispatcher
}

func (c *ConnectionManager) HandleMsg(msg []byte) {
	var incoming IncomingMessage
	if err := json.Unmarshal(msg, &incoming); err != nil {
		log.Println("error parsing message:", err)
		return
	}
	if incoming.Type == "command" {
		c.dispatcher.DispatchCommand(string(incoming.Data))
	}
}

func (c *ConnectionManager) Register(msg []byte, conn *websocket.Conn) bool {
	var incoming IncomingMessage
	if err := json.Unmarshal(msg, &incoming); err != nil {
		log.Println("error parsing register message:", err)
		return false
	}

	if incoming.Type != "register" {
		return false
	}

	if incoming.Data == "player" {
		c.PlayerConn = conn
		if c.playerModule != nil {
			c.playerModule.SetConn(conn)
		}
		log.Printf("player registrado")
	} else {
		c.controllerConns = append(c.controllerConns, NewClient(conn))
		log.Printf("controller registrado")
	}

	return true
}

func NewManager() *ConnectionManager {
	return &ConnectionManager{
		controllerConns: make([]*Client, 0),
	}
}

func (c *ConnectionManager) SetDispatcher(d *dispatcher.Dispatcher) {
	c.dispatcher = d
}

func (c *ConnectionManager) SetPlayerModule(p *player.Module) {
	c.playerModule = p
}

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (c *ConnectionManager) SendMessage(msgType string, data any, target *websocket.Conn) {
	if target == nil {
		return
	}

	payload, err := json.Marshal(Message{Type: msgType, Data: data})
	if err != nil {
		log.Println("SendMessage marshal error:", err)
		return
	}

	if err := target.WriteMessage(websocket.TextMessage, payload); err != nil {
		log.Println("SendMessage write error:", err)
	}
}

func (c *ConnectionManager) Event(msgType string, data any) {
	for _, client := range c.controllerConns {
		c.SendMessage(msgType, data, client.Conn)
	}
}
