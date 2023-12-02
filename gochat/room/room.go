package room

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	models "github.com/stefan-chivu/gochat/gochat/models"
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
	Capacity int `json:"capacity"`

	Messages []*models.Message

	forward chan *models.Message
}

func NewRoom(name string, capacity int) *Room {
	return &Room{
		name:     name,
		Capacity: capacity,
		clients:  make(map[*websocket.Conn]string),
		Messages: make([]*models.Message, 0),
		forward:  make(chan *models.Message),
	}
}

func (r *Room) broadcast(msg *models.Message) {
	for client := range r.clients {
		go func(ws *websocket.Conn) {
			r.mu.Lock()
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg.Content)); err != nil {
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
				continue
			}

			if _, ok := r.clients[ws]; ok {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					r.handleClose(ws, fmt.Sprintf("[ %s ] %s is going away", r.name, r.clients[ws]))
					continue
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					r.handleClose(ws, fmt.Sprintf("[ %s ] %s closed unexpectedly", r.name, r.clients[ws]))
					continue
				}
			}

			log.Default().Println("Websocket read error", err)
			continue // break????
		}

		log.Default().Printf("[ %s ] received message: [ %s : %s ]", r.name, r.clients[ws], string(buff))

		r.forward <- (&models.Message{
			Username:  r.clients[ws],
			Content:   string(buff),
			Timestamp: time.Now().UTC(),
		})
	}
}

func (r *Room) GetClients() map[*websocket.Conn]string {
	return r.clients
}

func (r *Room) RemoveClient(ws *websocket.Conn) {
	delete(r.clients, ws)
}

func (r *Room) HandleRoomConnection(w http.ResponseWriter, req *http.Request) {
	if len(r.clients) >= r.Capacity {
		http.Error(w, fmt.Sprintf("Room '%s' is full; Max capacity: %d", r.name, r.Capacity), http.StatusNotAcceptable)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Parse form failed", http.StatusBadRequest)
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

	r.mu.Lock()
	r.clients[socket] = username
	r.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username + "; Room: " + r.name)

	go r.handleRoomMsg()
	r.readLoop(socket)
}

func (r *Room) handleRoomMsg() {
	for {
		r.mu.Lock()
		msg := <-r.forward
		log.Default().Printf("[ %s ] %s : %s", r.name, msg.Username, msg.Content)
		r.Messages = append(r.Messages, msg)
		//TOTO db.syncMsg()
		r.broadcast(msg)
		r.mu.Unlock()
	}
}

func (r *Room) GetRoomMessages(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	responseData, err := json.Marshal(r.Messages)

	if err != nil {
		http.Error(w, "Room messages JSON marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Write(responseData)
}

func (r *Room) handleClose(ws *websocket.Conn, message string) {
	log.Default().Print(message)
	delete(r.clients, ws)
}
