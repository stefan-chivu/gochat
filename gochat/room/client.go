package room

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stefan-chivu/gochat/pkg/models"
)

type client struct {

	// Socket is the web Socket for this client.
	socket *websocket.Conn

	// Receive is a channel to Receive messages from other clients.
	receive chan []byte

	room *Room
}

func RunClient() {
	url := "ws://localhost:8080/ws"
	randId := rand.Intn(10)
	message := models.Message{Message: fmt.Sprintf("Hello world from my client %d !", randId), Username: fmt.Sprintf("client %d", randId)}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("error dialing %s\n", err)
	}
	defer c.Close()

	done := make(chan bool)
	// reading server messages
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("error reading: %s\n", err)
				return
			}
			fmt.Printf("Got message: %s\n", message)
		}
	}()

	// writing messages to server
	go func() {
		for {
			err := c.WriteJSON(message)
			if err != nil {
				log.Printf("error writing %s\n", err)
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()

	<-done
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
