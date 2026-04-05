package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/jwt"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Conn struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
	RoomID   string
	SendChan chan []byte
	Closed   bool
	LastBeat time.Time
}

var rooms = make(map[string]map[*Conn]bool)
var userConns = make(map[string]*Conn)
var mu sync.Mutex

type Message struct {
	Type         int    `json:"type"`
	MsgID        string `json:"msg_id"`
	RoomID       string `json:"room_id"`
	FromUID      string `json:"from_uid"`
	FromUsername string `json:"from_username"`
	Content      string `json:"content"`
	SendTime     int64  `json:"send_time"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	token := r.URL.Query().Get("token")

	if room == "" {
		http.Error(w, "Room parameter is required", http.StatusBadRequest)
		return
	}

	if token == "" {
		http.Error(w, "Token parameter is required", http.StatusUnauthorized)
		return
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var user entity.User
	if err := database.DB.Where("uid = ?", claims.UID).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}

	conn := &Conn{
		Conn:     wsConn,
		UserID:   user.UID,
		Username: user.Username,
		RoomID:   room,
		SendChan: make(chan []byte, 256),
		Closed:   false,
		LastBeat: time.Now(),
	}

	mu.Lock()
	if _, exists := rooms[room]; !exists {
		rooms[room] = make(map[*Conn]bool)
	}
	rooms[room][conn] = true
	userConns[user.UID] = conn
	mu.Unlock()

	log.Printf("User %s joined room: %s, total users: %d", user.Username, room, len(rooms[room]))

	go conn.writePump()
	conn.readPump()
}

func (c *Conn) readPump() {
	defer func() {
		c.cleanup()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}

		log.Printf("Received message from %s in room %s: %s", c.Username, c.RoomID, message)

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			msg = Message{
				Type:         1,
				RoomID:       c.RoomID,
				FromUID:      c.UserID,
				FromUsername: c.Username,
				Content:      string(message),
				SendTime:     time.Now().Unix(),
			}
		} else {
			msg.FromUID = c.UserID
			msg.FromUsername = c.Username
			msg.SendTime = time.Now().Unix()
		}

		msgBytes, _ := json.Marshal(msg)

		mu.Lock()
		for conn := range rooms[c.RoomID] {
			select {
			case conn.SendChan <- msgBytes:
			default:
				close(conn.SendChan)
				delete(rooms[c.RoomID], conn)
			}
		}
		mu.Unlock()
	}
}

func (c *Conn) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.SendChan {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		n := len(c.SendChan)
		for i := 0; i < n; i++ {
			w.Write([]byte{'\n'})
			w.Write(<-c.SendChan)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

func (c *Conn) cleanup() {
	mu.Lock()
	defer mu.Unlock()

	if !c.Closed {
		c.Closed = true
		close(c.SendChan)
		delete(rooms[c.RoomID], c)
		delete(userConns, c.UserID)
		log.Printf("User %s left room: %s", c.Username, c.RoomID)
	}
}
