package chat

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Subscription contains the Client Connection and Room
type Subscription struct {
	Client *Client
	Room   string
}

// Close will terminate the client websocket connection
func (s *Subscription) Close() {
	s.Client.Conn.Close()
}

// Read will proceed to read the messages published by the client
func (s *Subscription) Read(pubsub *PubSub, room *Room) {
	c := s.Client

	c.Conn.SetReadLimit(maxMessageSize)
	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Println(errors.Wrap(err, "websocket closed"))
			break
		}

		if err := pubsub.Publish(msg); err != nil {
			log.Println(errors.Wrap(err, "error publishing"))
			break
		}
	}

	room.Unsubscribe <- s

	s.Close()
}

// Write will write new messages to the client
func (s *Subscription) Write() {
	c := s.Client
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteJSON(message)
		}
	}
}

// NewSubscription will return a new pointer to the subscription
func NewSubscription(room string, client *Client) *Subscription {
	return &Subscription{
		Client: client,
		Room:   room,
	}
}
