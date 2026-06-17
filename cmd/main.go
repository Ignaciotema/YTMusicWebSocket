package main

import (
	"log"
	"net/http"

	"github.com/Ignaciotema/YTMusic-web-socket/internal/dispatcher"
	"github.com/Ignaciotema/YTMusic-web-socket/internal/player"
	"github.com/Ignaciotema/YTMusic-web-socket/internal/transport/httpHandler"
	"github.com/Ignaciotema/YTMusic-web-socket/internal/transport/websocketHandler"
)

func main() {
	manager := websocketHandler.NewManager()

	playerModule := player.New(manager)

	d := dispatcher.New(playerModule)

	manager.SetDispatcher(d)
	manager.SetPlayerModule(playerModule)

	wsHandler := websocketHandler.NewWebSocketHandler(manager)
	httpHandler := httpHandler.NewHTTPHandler(d)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.HandleFunc("/", httpHandler.ServeHTTP)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Server iniciado en http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
