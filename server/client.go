package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	defer func() {
		log.Println("client disconnected")
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WSServe(hub *hub, w http.ResponseWriter, r *http.Request) {
	log.Println("client connected")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte)}
	client.hub.register <- client

	go client.writePump()
}
