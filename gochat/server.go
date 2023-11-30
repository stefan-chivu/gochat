package gochat

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/stefan-chivu/gochat/gochat/configuration"
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
	config *configuration.ServerConfig
}

func NewServer(config *configuration.ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

func (*Server) StartServer(opts *StartOpts) error {
	// http.HandleFunc("/", home)
	// http.HandleFunc("/ws", handleConnections)

	// go handleMsg()

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

	// m.Lock()
	// for conn := range userConnections {
	// 	conn.Close()
	// 	delete(userConnections, conn)
	// }
	// m.Unlock()

	server.Shutdown(nil)
	if signal == shutdown {
		return nil
	}
	return nil
}
