package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

var clients = make(map[*Client]struct{})

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &Client{conn: conn}

	// Add the client to the clients map
	clients[client] = struct{}{}

	go handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func() {
		// Remove the client when the connection is closed
		client.mu.Lock()
		delete(clients, client)
		client.mu.Unlock()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		// Broadcast the message to all clients
		broadcast(message)
	}
}

func broadcast(message []byte) {
	for client := range clients {
		client.mu.Lock()
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		client.mu.Unlock()
		if err != nil {
			fmt.Println(err)
		}
	}
}
