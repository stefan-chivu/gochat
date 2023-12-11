package server

import (
	"encoding/json"
	"fmt"
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

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{ \"msg\": \"Hello world from my server!\" }")
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
	// http.HandleFunc("/rooms/"+roomName, s.Rooms[roomName].ServeWs)
	http.HandleFunc("/rooms/"+roomName+"/messages", s.Rooms[roomName].GetRoomMessages)
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
