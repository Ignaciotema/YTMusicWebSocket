package websocketHandler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	manager  *ConnectionManager
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(manager *ConnectionManager) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		manager: manager,
	}
}

func CloseConnection(conn *websocket.Conn) {
	conn.Close()

}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer CloseConnection(conn)

	log.Printf("Conexion con WS establecida")

	// primer mensaje debe ser de registro
	_, firstMsg, err := conn.ReadMessage()
	if err != nil {
		return
	}
	if !h.manager.Register(firstMsg, conn) {
		log.Printf("Registro fallido, cerrando conexion")
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		log.Printf("Mensaje recibido: %s", msg)
		h.manager.HandleMsg(msg)
	}
}
