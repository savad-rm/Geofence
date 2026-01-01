package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub *AlertHub

type Client struct {
	hub  *AlertHub
	conn *websocket.Conn
	send chan interface{}
}

type AlertHub struct {
	clients    map[*Client]bool
	broadcast chan interface{}
	register   chan *Client
	unregister chan *Client
}

func NewAlertHub() *AlertHub {
	return &AlertHub{
		clients:    make(map[*Client]bool),
		broadcast: make(chan interface{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *AlertHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("Client connected")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Client disconnected")
			}

		case alert := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- alert:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func handleWebSocket(alertHub *AlertHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		hub:  alertHub,
		conn: conn,
		send: make(chan interface{}, 256),
	}

	alertHub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("WebSocket error:", err)
			}
			return
		}

		log.Println("WebSocket message:", string(message))
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case alert, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			alertJSON, _ := json.Marshal(alert)
			w.Write(alertJSON)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
