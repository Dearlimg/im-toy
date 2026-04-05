package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms = make(map[string]map[*websocket.Conn]bool)
var mu sync.Mutex

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	if room == "" {
		http.Error(w, "Room parameter is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	if _, exists := rooms[room]; !exists {
		rooms[room] = make(map[*websocket.Conn]bool)
	}
	rooms[room][conn] = true
	mu.Unlock()

	log.Printf("User joined room: %s, total users: %d", room, len(rooms[room]))

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			mu.Lock()
			delete(rooms[room], conn)
			mu.Unlock()
			break
		}

		log.Printf("Received message in room %s: %s", room, p)

		mu.Lock()
		for otherConn := range rooms[room] {
			if otherConn != conn {
				err = otherConn.WriteMessage(messageType, p)
				if err != nil {
					log.Println("Error writing message:", err)
					delete(rooms[room], otherConn)
					otherConn.Close()
				}
			}
		}
		mu.Unlock()
	}
}
