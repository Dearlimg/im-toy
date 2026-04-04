package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}

	clients   = make(map[*websocket.Conn]string) // conn -> userID
	clientsMu sync.Mutex
	userIDSeq = 0
)

type Message struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Content string `json:"content"`
}

func main() {
	// Serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	// WebSocket endpoint
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Assign user ID
	clientsMu.Lock()
	userIDSeq++
	userID := fmt.Sprintf("user%d", userIDSeq)
	clients[conn] = userID
	clientsMu.Unlock()

	log.Printf("User %s connected", userID)

	// Send welcome message with user ID
	welcomeMsg := fmt.Sprintf("Welcome! Your user ID is: %s", userID)
	conn.WriteJSON(Message{From: "server", Content: welcomeMsg})

	// Broadcast new user joined
	broadcastMessage(Message{
		From:    "server",
		Content: fmt.Sprintf("User %s joined the chat", userID),
	})

	// Read messages from client
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error for user %s: %v", userID, err)
			break
		}

		msg.From = userID // Override from field with actual user ID

		if msg.To == "" {
			// Broadcast to all if no recipient specified
			broadcastMessage(msg)
		} else {
			// Send to specific user
			sendToUser(msg.To, msg)
		}
	}

	// Clean up
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()

	log.Printf("User %s disconnected", userID)
	broadcastMessage(Message{
		From:    "server",
		Content: fmt.Sprintf("User %s left the chat", userID),
	})
}

func broadcastMessage(msg Message) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn := range clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

func sendToUser(userID string, msg Message) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn, id := range clients {
		if id == userID {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Println("Write error:", err)
				conn.Close()
				delete(clients, conn)
			}
			return
		}
	}

	// If user not found, send error back to sender
	for conn, id := range clients {
		if id == msg.From {
			err := conn.WriteJSON(Message{
				From:    "server",
				Content: fmt.Sprintf("User %s not found", userID),
			})
			if err != nil {
				log.Println("Write error:", err)
			}
			return
		}
	}
}
