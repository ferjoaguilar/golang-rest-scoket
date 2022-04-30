package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	id       string
	socket   websocket.Conn
	outbound chan []byte
}

func NewClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		socket:   *socket,
		outbound: make(chan []byte),
	}
}

func (c *Client) Write() {
	for {
		select {
		case message, ok := <-c.outbound:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.onConnect(client)
		case client := <-hub.unregister:
			hub.onDisconnect(client)
		}
	}
}

func (hub *Hub) onConnect(client *Client) {
	log.Print("Client connected", client.socket.RemoteAddr())

	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	client.id = client.socket.RemoteAddr().String()
}

func (hub *Hub) onDisconnect(client *Client) {
	log.Print("Client disconnected", client.socket.RemoteAddr())

	client.socket.Close()
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	i := -1
	for j, c := range hub.clients {
		if c.id == client.id {
			i = j
		}
	}

	copy(hub.clients[i:], hub.clients[i+1:])
	hub.clients[len(hub.clients)-1] = nil
	hub.clients = hub.clients[:len(hub.clients)-1]

}
