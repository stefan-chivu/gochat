package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/stefan-chivu/gochat/gochat/room"
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

func (s *Server) home(w http.ResponseWriter, req *http.Request) {
	ws, err := Upgrade(w, req)

	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
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

	s.mu.Lock()
	s.Clients[ws] = username
	s.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username)
	// go s.writer(ws)
	s.reader(ws)
}

func (s *Server) getUserMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Default().Print("Websocket upgrade failed:", err)
		return
	}
	user := s.Clients[socket]
	messageList := s.Messages[user]
	responseData, err := json.Marshal(messageList)

	if err != nil {
		http.Error(w, "Message list JSON marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Write(responseData)
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return ws, err
	}
	return ws, nil
}

func (s *Server) reader(conn *websocket.Conn) {
	for {
		messageType, buff, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				continue
			}

			if _, ok := s.Clients[conn]; ok {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					s.handleClose(conn, fmt.Sprintf("%s is going away", s.Clients[conn]))
					break
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					s.handleClose(conn, fmt.Sprintf("%s closed unexpectedly", s.Clients[conn]))
					break
				}
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					s.handleClose(conn, fmt.Sprintf("%s closed abnormally", s.Clients[conn]))
					break
				}
			}

			log.Default().Println("Websocket read error", err)
			break
		}
		if err := conn.WriteMessage(messageType, buff); err != nil {
			log.Println(err)
			continue
		}
	}
}

// func (s *Server) writer(conn *websocket.Conn) {
// 	for {
// 		messageType, buff, err := conn.NextReader()
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		w, err := conn.NextWriter(messageType)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		if _, err := io.Copy(w, buff); err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		if err := w.Close(); err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 	}
// }

func (s *Server) ServeWs(w http.ResponseWriter, req *http.Request) {
	ws, err := Upgrade(w, req)

	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
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

	s.mu.Lock()
	s.Clients[ws] = username
	s.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userList := []string{}

	for _, v := range s.Clients {
		userList = append(userList, v)
	}

	responseData, err := json.Marshal(userList)

	if err != nil {
		http.Error(w, "User list JSON marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Write(responseData)
}

func (s *Server) createPrivateChat(w http.ResponseWriter, r *http.Request) {
	s.Config.Log.Info().Msg(httpReqLogMsg(r, "Create Room Request received"))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data from the POST request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	username1 := r.Form.Get("username1")

	if ok, reason := s.isValidUsername(username1); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}

	username2 := r.Form.Get("username2")

	if ok, reason := s.isValidUsername(username2); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}

	chat := room.NewPrivateChat(username1, username2)

	s.Rooms[chat.Name] = chat
}

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {

	s.Config.Log.Info().Msg(httpReqLogMsg(r, "Create Room Request received"))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data from the POST request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	roomName := r.Form.Get("roomName")

	if ok, reason := isValidRoomName(roomName); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}

	capacity, err := strconv.Atoi(r.Form.Get("capacity"))

	if err != nil {
		http.Error(w, "Invalid room capacity parameter", http.StatusBadRequest)
		return
	}

	if capacity < 5 || capacity > 20 {
		http.Error(w, "Room capacity must be a value between 5 and 20", http.StatusBadRequest)
		return
	}

	if _, ok := s.Rooms[roomName]; ok {
		http.Error(w, "A room named "+roomName+" already exists", http.StatusNotAcceptable)
		s.Config.Log.Error().Msgf("Room '%s' creation failed. Already exists.", roomName)
		return
	}

	s.Rooms[roomName] = room.NewRoom(roomName, capacity)
	http.HandleFunc("/rooms/"+roomName, s.Rooms[roomName].HandleRoomConnection)
	http.HandleFunc("/rooms/"+roomName+"/messages", s.Rooms[roomName].GetRoomMessages)
	http.HandleFunc("/rooms/"+roomName+"/users", s.Rooms[roomName].GetRoomUsers)
	s.Config.Log.Info().Msgf("Room '" + roomName + "' has been created")

	// TODO: save room data to DB
}

func (s *Server) getRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomData := map[string]*room.RoomInfo{}
	for name, r := range s.Rooms {
		roomData[name] = &room.RoomInfo{
			Capacity:    r.Capacity,
			ClientCount: len(r.Clients),
		}
	}
	responseData, err := json.Marshal(roomData)

	if err != nil {
		http.Error(w, "Room data JSON marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Write(responseData)
}

func httpReqLogMsg(r *http.Request, message string) string {
	return "Src: [ " + r.RemoteAddr + " ] ; Dst: [ " + r.Host + " ] ; Method: " + r.Method + "; " + message
}

func isValidRoomName(roomName string) (bool, string) {
	if roomName == "" {
		return false, "Room name cannot be empty"
	}

	if len(roomName) > 20 {
		return false, "Room name should not exceed 20 characters"
	}

	// TODO: think of more constraints

	return true, ""
}

func (s *Server) isValidUsername(username string) (bool, string) {
	if username == "" {
		return false, "Invalid username"
	}

	// if _, ok := s.Clients ; !ok {
	// 	return false, "User is not connected"
	// }

	// TODO: think of more constraints

	return true, ""
}

func (s *Server) handleClose(ws *websocket.Conn, message string) {
	log.Default().Print(message)
	delete(s.Clients, ws)
}
