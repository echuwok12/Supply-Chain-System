package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader config
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for dev
	},
}

type Handler struct {
	clients   map[*websocket.Conn]bool
	broadcast chan interface{}
}

func NewHandler() *Handler {
	return &Handler{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}),
	}
}

// HandleConnection upgrades HTTP to WS
func (h *Handler) HandleConnection(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS Upgrade Error:", err)
		return
	}
	// defer ws.Close() // Keep connection open

	h.clients[ws] = true

	// Keep-alive loop
	// If this loop breaks (connection closed), remove client
	go func() {
		defer func() {
			ws.Close()
			delete(h.clients, ws)
		}()
		for {
			var msg interface{}
			if err := ws.ReadJSON(&msg); err != nil {
				break
			}
		}
	}()
}

func (h *Handler) Run() {
	for {
		msg := <-h.broadcast
		for client := range h.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("WS Write Error:", err)
				client.Close()
				delete(h.clients, client)
			}
		}
	}
}

func (h *Handler) Broadcast(message interface{}) {
	// Non-blocking send (optional safety)
	go func() {
		h.broadcast <- message
	}()
}