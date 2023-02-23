package websocket

import (
  "fmt"
  "example/golang-api-news/models"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan models.Album
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan models.Album),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Pool.Register: Size of Connection Pool: ", len(pool.Clients))
			// for client, _ := range pool.Clients {
			// 	fmt.Println(client)
				// client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			// }
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Pool.Unregister: Size of Connection Pool: ", len(pool.Clients))
			// for client, _ := range pool.Clients {
			// 	client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			// }
		case album := <-pool.Broadcast:
			fmt.Printf("Pool.Broadcast album id %d, price %f", album.ID, album.Price)
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(album); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
