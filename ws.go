package main

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Client struct {
	conn *websocket.Conn
	mu sync.Mutex
}

var clients = make(map[*Client]struct{})

func echoHandler(conn *websocket.Conn) {
	fmt.Println("Client connected")

	client := &Client{conn: conn}

	clients[client] = struct{}{}

	handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func ()  {
		client.mu.Lock()
		delete(clients, client)
		client.mu.Unlock()
	}()

	for {
		var msg string
		if err := websocket.Message.Receive(client.conn, &msg); err != nil {
			fmt.Println("Error receiving message:", err)
			break
		}

		fmt.Println("Received message:", msg)
		fmt.Println(clients)

		for client := range clients {
			// Echo the message back to the client
			client.mu.Lock()
			err := websocket.Message.Send(client.conn, msg); 
			client.mu.Unlock()
			if err != nil {
				fmt.Println("Error sending message:", err)
				break
			}
		}

	}

	defer client.conn.Close()

	fmt.Println("Client disconnected")
}

func main() {
	http.Handle("/ws", websocket.Handler(echoHandler))
	port := 8080
	fmt.Printf("Server started on :%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println(err)
	}
}