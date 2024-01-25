package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	// websocket connection
	socket *websocket.Conn
	// buffered channel of outbound messages
	send chan []byte

	// room is the room this client is chatting in
	room *room
}

// readPump pumps messages from the websocket connection to the hub.
func (c *client) readPump() {
	defer c.socket.Close()
	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *client) writePump() {
	defer c.socket.Close()

	for message := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}
