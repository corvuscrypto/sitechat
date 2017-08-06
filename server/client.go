package main

import (
	"encoding/json"
	"net/http"
	"strings"
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
	Host            string
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
		mType, rawMessage, err := c.Socket.ReadMessage()
		if err != nil {
			c.cleanup()
			return
		}
		if mType == websocket.TextMessage {
			evt := new(event)
			err := json.Unmarshal(rawMessage, evt)
			if err != nil {
				continue
			}
			switch evt.Type {
			case "location":
				location := new(locationUpdate)
				if err := json.Unmarshal(evt.Data, location); err != nil {
					continue
				}
				normalizedHost := strings.ToLower(location.Host)
				if normalizedHost != c.Host {
					if !isAllowedHost(normalizedHost) {
						c.cleanup()
						return
					}
					c.Host = normalizedHost
					switchClientList(c, normalizedHost)
				}
				c.Path = location.Path

			case "message":
				message := &messageEvent{
					Username: c.Username,
				}
				err := json.Unmarshal(evt.Data, message)
				if err != nil {
					continue
				}
				c.clientList.BroadcastMessage(message, c.Path)

			default:
				continue
			}
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

func switchClientList(c *Client, host string) {
	currentList := c.clientList
	newList, ok := siteBuckets[host]
	if !ok {
		newList = NewClientList()
		siteBuckets[host] = newList
	}
	currentList.RemoveClient(c)
	newList.AddClient(c)
	c.clientList = newList
}

// NewClient is a constructor for creating a new client based on a connect
// request.
func NewClient(w http.ResponseWriter, r *http.Request) *Client {
	path := r.URL.Query().Get("path")
	client := new(Client)
	client.Username = generateUsername()
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
func (c *ClientList) BroadcastMessage(message *messageEvent, path string) {
	if messageBytes, err := json.Marshal(message); err == nil {
		for _, client := range c.clients {
			if client.Path == path {
				client.sendMessage(messageBytes)
			}
		}
	}
}
