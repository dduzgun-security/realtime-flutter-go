package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	websocket "github.com/gofiber/websocket/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Username  string    `gorm:"not null"`
	Text      string    `gorm:"not null"`
	CreatedAt time.Time `json:"timestamp"`
}

var (
	connections []*websocket.Conn
	mutex       sync.Mutex
)

func handleWebSocket(db *gorm.DB, c *websocket.Conn) {
	defer func() {
		// Remove connection from active connections list
		mutex.Lock()
		for i, conn := range connections {
			if conn == c {
				connections = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		mutex.Unlock()

		c.Close()
	}()

	// Add connection to active connections list
	mutex.Lock()
	connections = append(connections, c)
	mutex.Unlock()

	for {
		// Read incoming message from client
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Parse incoming message
		var data struct {
			Username string `json:"username"`
			Text     string `json:"text"`
		}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Println(err)
			continue
		}

		// Store message in database
		message := &Message{
			Username: data.Username,
			Text:     data.Text,
		}
		err = db.Create(message).Error
		if err != nil {
			log.Println(err)
			continue
		}

		// Broadcast message to all connected clients
		broadcastMessage(db, message)
	}
}

func broadcastMessage(db *gorm.DB, message *Message) {
	// Query all messages from database
	var messages []*Message
	err := db.Order("created_at desc").Limit(10).Find(&messages).Error
	if err != nil {
		log.Println(err)
		return
	}

	// Encode messages as JSON
	type messageData struct {
		Username  string    `json:"username"`
		Text      string    `json:"text"`
		Timestamp time.Time `json:"timestamp"`
	}
	var data []*messageData
	for _, message := range messages {
		data = append(data, &messageData{
			Username:  message.Username,
			Text:      message.Text,
			Timestamp: message.CreatedAt,
		})
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	// Broadcast JSON data to all connected clients
	mutex.Lock()
	for _, c := range connections {
		err = c.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	mutex.Unlock()
}

func main() {
	// Connect to PostgreSQL database
	dbURL := "postgres://pg:pass@localhost:5432/crud"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	// Auto-migrate the Message model
	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Fatal(err)
	}

	// Create Fiber app
	app := fiber.New()

	// Serve static files
	app.Static("/", "static")

	// Handle WebSocket connections
	app.Use("/chat", websocket.New(func(c *websocket.Conn) {
		handleWebSocket(db, c)
	}))

	// Start server
	log.Println("Server started on :8080")
	log.Fatal(app.Listen(":8080"))
}
