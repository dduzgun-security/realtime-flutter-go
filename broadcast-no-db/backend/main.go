package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func main() {
	// Create a new Fiber app
	app := fiber.New()

	// WebSocket endpoint
	app.Get("/ws", websocket.New(handleWebSocket))

	// Start listening for incoming messages to broadcast
	go handleMessages()

	// Start the server
	log.Println("Server starting on port 8080...")
	log.Fatal(app.Listen(":8080"))
}

func handleWebSocket(c *websocket.Conn) {
	// Register the client connection
	clients[c] = true

	// Close the connection and remove the client when needed
	defer func() {
		delete(clients, c)
		c.Close()
	}()

	// Continuously listen for incoming messages
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Broadcast the message to all connected clients
		broadcast <- message
	}
}

func handleMessages() {
	for {
		message := <-broadcast

		// Send the message to all connected clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
