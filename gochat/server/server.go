package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/stefan-chivu/gochat/gochat/auth"
	"github.com/stefan-chivu/gochat/gochat/configuration"
	"github.com/stefan-chivu/gochat/gochat/models"
	"github.com/stefan-chivu/gochat/gochat/room"
)

var shutdown os.Signal = syscall.SIGUSR1

var (
	// Buildtime is set to the current time during the build process by GOLDFLAGS
	Buildtime string
	// Version is set to the current git tag during the build process by GOLDFLAGS
	Version string
)

var (
	CPUProfile   string
	PrintVersion bool
	PProf        bool
)

// StartOpts is passed to StartServer() and is used to set the running configuration
type StartOpts struct {
}

type ServerOpts struct {
	Config *configuration.ServerConfig
}

type Server struct {
	mu sync.Mutex

	Config *configuration.ServerConfig

	Mux *http.ServeMux
	// Rooms represent the rooms currently available on the server
	Rooms map[string]*room.Room

	Clients map[*websocket.Conn]string
	// TODO Replace string with User at some point
	Messages map[string]([]*models.Message)
}

func NewServer(config *configuration.ServerConfig) *Server {
	// TODO: Implement database fetch for existing rooms
	// rooms := db.getRooms()

	auth.NewCookieStore()

	rooms := make(map[string]*room.Room)

	if len(rooms) == 0 {
		rooms["Global"] = room.NewRoom("Global", 50)
		http.HandleFunc("/rooms/Global", rooms["Global"].HandleRoomConnection)
		// http.HandleFunc("/rooms/Global", rooms["Global"].ServeWs)
		http.HandleFunc("/rooms/Global/messages", rooms["Global"].GetRoomMessages)
	}

	messages := make(map[string][]*models.Message)
	return &Server{
		Config:   config,
		Rooms:    rooms,
		Messages: messages,
		Clients:  make(map[*websocket.Conn]string),
	}
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	http.HandleFunc("/chat/create", s.createPrivateChat)
	http.HandleFunc("/rooms/create", s.createRoom)
	http.HandleFunc("/rooms", s.getRooms)
	http.HandleFunc("/users/login", auth.Login)
	http.HandleFunc("/users", s.getUsers)
	// http.HandleFunc("/users/register", auth.)
	http.HandleFunc("/messages", s.getUserMessages)

	http.HandleFunc("/", s.home)
	// http.HandleFunc("/ws", serveWs)
}

func (s *Server) StartServer(opts *StartOpts) error {
	crt, _ := os.ReadFile(s.Config.ServerTLSCert)
	if string(crt) != "" {
		// TODO: Enable TLS
		s.Config.Log.Info().Msg("Received TLS Certs")
	}
	s.Mux = http.NewServeMux()

	s.setupRoutes(s.Mux)

	c := cors.Default()

	// Wrap the mux with the CORS middleware
	handler := c.Handler(http.DefaultServeMux)

	server := &http.Server{Addr: s.Config.ServerListenAddress, Handler: handler}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("Starting server on %s\n", s.Config.ServerListenAddress)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("error starting server: %s", err)
			stop <- shutdown
		}
	}()

	signal := <-stop
	log.Printf("Shutting down server ... ")

	for _, room := range s.Rooms {
		clients := room.GetClients()
		for ws, username := range clients {
			ws.Close()
			room.RemoveClient(ws)
			s.Config.Log.Info().Msgf("Disconnected user %s", username)
		}
	}

	server.Shutdown(context.TODO())
	if signal == shutdown {
		return nil
	}
	return nil
}
