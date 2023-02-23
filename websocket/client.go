package websocket

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		// just listening for socket close message
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			// assuming close. this will invoke the deferred unregister of this client
			log.Println(err)
			return
		}
		log.Printf("Unexpected message of type %v received: %+v\n", messageType, string(p))
	}
}
