package room

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

type Room struct {
	mu sync.Mutex

	name string

	// clients holds all current clients in this room.
	clients map[*websocket.Conn]string

	// the maximum capacity of a room
	capacity int
}

func NewRoom(name string, capacity int) *Room {
	return &Room{
		name:     name,
		capacity: capacity,
		clients:  make(map[*websocket.Conn]string),
	}
}

func (r *Room) broadcast(msg []byte) {
	for client := range r.clients {
		go func(ws *websocket.Conn) {
			r.mu.Lock()
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Default().Println("Websocket write error: ", err)
			}
			r.mu.Unlock()
		}(client)
	}
}

func (r *Room) readLoop(ws *websocket.Conn) {
	for {
		_, buff, err := ws.ReadMessage()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Default().Println("Websocket read error", err)
		}

		r.broadcast(buff)
	}
}

func (r *Room) GetClients() map[*websocket.Conn]string {
	return r.clients
}

func (r *Room) RemoveClient(ws *websocket.Conn) {
	delete(r.clients, ws)
}

func (r *Room) HandleConnection(w http.ResponseWriter, req *http.Request) {
	if len(r.clients) >= r.capacity {
		http.Error(w, fmt.Sprintf("Room '%s' is full; Max capacity: %d", r.name, r.capacity), http.StatusNotAcceptable)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	username := req.Form.Get("username")

	// TODO better valid username check
	if username == "" {
		// error case
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Default().Print("Websocket upgrade failed:", err)
		return
	}

	r.clients[socket] = username

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username + "; Room: " + r.name)

	r.readLoop(socket)
}
