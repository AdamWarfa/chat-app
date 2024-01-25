package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// Registered clients.
	clients map[*client]bool

	// Register requests from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client

	// Inbound messages from the clients.
	forward chan []byte
}

// newRoom creates a new room
func newRoom() *room {
	return &room{
		forward:    make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[*client]bool),
	}
}

// run runs the room in a goroutine
func (r *room) run() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
		case client := <-r.unregister:
			delete(r.clients, client)
			close(client.send)
		case message := <-r.forward:
			for client := range r.clients {
				client.send <- message
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

// ServeHTTP handles the websocket requests from the peer.
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.register <- client
	defer func() {
		r.unregister <- client
	}()
	go client.writePump()
	client.readPump()
}
