package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/stefan-chivu/gochat/gochat/configuration"
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
	Config *configuration.ServerConfig
	// Rooms represent the rooms currently available on the server
	Rooms   map[string]*room.Room
	Clients map[*websocket.Conn]string
}

func NewServer(config *configuration.ServerConfig) *Server {
	// TODO: Implement database fetch for existing rooms
	// rooms := db.getRooms()
	rooms := make(map[string]*room.Room)
	return &Server{
		Config: config,
		Rooms:  rooms,
	}
}

func (s *Server) StartServer(opts *StartOpts) error {
	crt, _ := os.ReadFile(s.Config.ServerTLSCert)
	if string(crt) != "" {
		// TODO: Enable TLS
		s.Config.Log.Info().Msg("Received TLS Certs")
	}
	http.HandleFunc("/", s.home)
	http.HandleFunc("/rooms/create", s.createRoom)
	http.HandleFunc("/rooms/get-rooms", s.getRooms)
	http.HandleFunc("/get-users", s.getActiveUsers)

	server := &http.Server{Addr: ":8080"}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("Starting server on %s\n", server.Addr)
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
