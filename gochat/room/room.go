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

	Name string

	// clients holds all current clients in this room.
	Clients map[*websocket.Conn]string

	// the maximum capacity of a room
	Capacity int

	Messages []*models.Message

	forward chan *models.Message
}

type RoomInfo struct {
	Capacity    int
	ClientCount int
}

func NewRoom(name string, capacity int) *Room {
	return &Room{
		Name:     name,
		Capacity: capacity,
		Clients:  make(map[*websocket.Conn]string),
		Messages: make([]*models.Message, 0),
		forward:  make(chan *models.Message),
	}
}

// func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return ws, err
// 	}
// 	return ws, nil
// }

// func (r *Room) reader(conn *websocket.Conn) {
// 	for {
// 		messageType, buff, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		// fmt.Println(string(p))
// 		msg := &models.Message{
// 			Username:  r.Clients[conn],
// 			Content:   string(buff),
// 			Timestamp: time.Now().UTC(),
// 		}

// 		msgData, err := json.Marshal(msg)

// 		if err := conn.WriteMessage(messageType, msgData); err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 	}
// }

// func (r *Room) writer() {
// 	for {
// 		fmt.Println("Sending")
// 		for conn, _ := range r.Clients {
// 			messageType, buff, err := conn.NextReader()
// 			if err != nil {
// 				fmt.Println(err)
// 				continue
// 			}
// 			w, err := conn.NextWriter(messageType)
// 			if err != nil {
// 				fmt.Println(err)
// 				continue
// 			}
// 			if _, err := io.Copy(w, buff); err != nil {
// 				fmt.Println(err)
// 				continue
// 			}
// 			if err := w.Close(); err != nil {
// 				fmt.Println(err)
// 				continue
// 			}
// 		}
// 	}
// }

// func (r *Room) ServeWs(w http.ResponseWriter, req *http.Request) {
// 	ws, err := Upgrade(w, req)

// 	if err != nil {
// 		fmt.Fprintf(w, "%+V\n", err)
// 	}

// 	if len(r.Clients) >= r.Capacity {
// 		http.Error(w, fmt.Sprintf("Room '%s' is full; Max capacity: %d", r.Name, r.Capacity), http.StatusNotAcceptable)
// 		return
// 	}

// 	if err := req.ParseForm(); err != nil {
// 		http.Error(w, "Parse form failed", http.StatusBadRequest)
// 		return
// 	}

// 	username := req.Form.Get("username")

// 	// TODO better valid username check
// 	if username == "" {
// 		// error case
// 		http.Error(w, "Invalid username", http.StatusBadRequest)
// 		return
// 	}

// 	r.mu.Lock()
// 	r.Clients[ws] = username
// 	r.mu.Unlock()

// 	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username + "; Room: " + r.Name)

// 	go r.writer()
// 	r.reader(ws)
// }

func (r *Room) broadcast(msg []byte) {
	for ws := range r.Clients {
		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Default().Println("Websocket write error: ", err)
		}
	}
}

func (r *Room) readLoop(ws *websocket.Conn) {
	layout := "02/01/2006 15:04"
	for {
		_, buff, err := ws.ReadMessage()
		if err != nil {
			if err == io.EOF {
				continue
			}

			if _, ok := r.Clients[ws]; ok {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					r.forward <- (&models.Message{
						Username:  "Server",
						Content:   r.Clients[ws] + " disconnected",
						Timestamp: time.Now().Format(layout),
					})

					r.handleClose(ws, fmt.Sprintf("[ %s ] %s is going away", r.Name, r.Clients[ws]))
					break
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					r.handleClose(ws, fmt.Sprintf("[ %s ] %s closed unexpectedly", r.Name, r.Clients[ws]))
					break
				}
			}

			log.Default().Println("Websocket read error", err)
			break
		}

		log.Default().Printf("[ %s ] received message: [ %s : %s ]", r.Name, r.Clients[ws], string(buff))

		r.forward <- (&models.Message{
			Username:  r.Clients[ws],
			Content:   string(buff),
			Timestamp: time.Now().Format(layout),
		})
	}
}

func (r *Room) GetClients() map[*websocket.Conn]string {
	return r.Clients
}

func (r *Room) RemoveClient(ws *websocket.Conn) {
	delete(r.Clients, ws)
}

func (r *Room) HandleRoomConnection(w http.ResponseWriter, req *http.Request) {
	if len(r.Clients) >= r.Capacity {
		http.Error(w, fmt.Sprintf("Room '%s' is full; Max capacity: %d", r.Name, r.Capacity), http.StatusNotAcceptable)
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
	r.Clients[socket] = username
	r.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username + "; Room: " + r.Name)

	go r.handleRoomMsg()
	r.readLoop(socket)
}

func (r *Room) handleRoomMsg() {
	for {
		msg := <-r.forward
		log.Default().Printf("[ %s ] %s : %s", r.Name, msg.Username, msg.Content)
		r.mu.Lock()
		r.Messages = append(r.Messages, msg)
		r.mu.Unlock()
		//TODO db.syncMsg()
		msgData, err := json.Marshal(msg)
		if err != nil {
			log.Default().Printf("Failed marshalling message into JSON")
			continue
		}
		r.broadcast(msgData)
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
	delete(r.Clients, ws)
}
