package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}

// Client represents a connected client which can send messages
type Client struct {
	Username        string
	LastMessageSent time.Time
	Socket          *websocket.Conn // change this to websocket type later
	Path            string
	clientList      *ClientList
}

func (c *Client) cleanup() {
	if err := c.Socket.Close(); err != nil {
		c.Socket.Close()
	}
	c.clientList.RemoveClient(c)
	c.clientList = nil
}

func (c *Client) waitForMessages() {
	for {
		mType, message, err := c.Socket.ReadMessage()
		if err != nil {
			c.cleanup()
			return
		}
		if mType == websocket.TextMessage {
			c.clientList.BroadcastMessage(messageToJSON(c.Username, message))
		}
	}
}

func (c *Client) sendMessage(message []byte) {
	for i := 512; i > 0; i >>= 1 {
		err := c.Socket.WriteMessage(websocket.TextMessage, message)
		if err == nil {
			return
		}
	}
}

// NewClient is a constructor for creating a new client based on a connect
// request.
func NewClient(w http.ResponseWriter, r *http.Request) *Client {
	username := r.URL.Query().Get("username")
	path := r.URL.Query().Get("path")
	if username == "" {
		return nil
	}
	client := new(Client)
	client.Username = username
	client.Path = path
	client.Socket, _ = upgrader.Upgrade(w, r, nil)
	client.LastMessageSent = time.Now()
	return client
}

// ClientList holds a slice of clients and contains a lock for self-management
type ClientList struct {
	sync.Mutex
	clients []*Client
}

// NewClientList is the constructor for the ClientList type
func NewClientList() *ClientList {
	return &ClientList{
		clients: make([]*Client, 0),
	}
}

// AddClient adds a client to the internal slice. Locks to prevent races.
func (c *ClientList) AddClient(client *Client) {
	c.Lock()
	defer c.Unlock()
	c.clients = append(c.clients, client)
}

// RemoveClient removes a client given. Locks the list to prevent races.
func (c *ClientList) RemoveClient(client *Client) {
	var toRemove = -1
	for i, v := range c.clients {
		if v == client {
			toRemove = i
		}
	}
	if toRemove >= 0 {
		c.Lock()
		defer c.Unlock()
		c.clients = append(c.clients[:toRemove], c.clients[toRemove+1:]...)
	}
}

// BroadcastMessage takes a message and Broadcasts it to all clients in the
// internal slice
func (c *ClientList) BroadcastMessage(message []byte) {
	for _, client := range c.clients {
		client.sendMessage(message)
	}
}
