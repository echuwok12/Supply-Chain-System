package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	clients   map[*websocket.Conn]bool
	broadcast chan interface{}
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}),
	}
}

func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	// Register client
	h.clients[ws] = true

	// Keep connection alive
	for {
		var msg interface{}
		// ReadJSON blocks until a message is received or connection is closed
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(h.clients, ws)
			break
		}
	}
}

// Run listens for messages on the broadcast channel and sends them to clients
func (h *WebSocketHandler) Run() {
	for {
		msg := <-h.broadcast
		for client := range h.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(h.clients, client)
			}
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHandler) Broadcast(message interface{}) {
	h.broadcast <- message
}
