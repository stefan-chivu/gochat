package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/stefan-chivu/gochat/gochat/room"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world from my server!")
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
	http.HandleFunc("/room/"+roomName, s.Rooms[roomName].HandleConnection)
	s.Config.Log.Info().Msgf("Room '" + roomName + "' has been created")

	// TODO: save room data to DB
}

func (s *Server) getRooms(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	responseData, err := json.Marshal(s.Rooms)

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
