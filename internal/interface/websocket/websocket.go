package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/idgen"
	"im-chat/internal/infra/jwt"
	infra_repo "im-chat/internal/infra/repository"

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

type WSMessage struct {
	Type      string `json:"type"`
	MsgID     string `json:"msg_id,omitempty"`
	FromUID   string `json:"from_uid,omitempty"`
	ToUID     string `json:"to_uid,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
	Content   string `json:"content,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	token := r.URL.Query().Get("token")

	log.Printf("WebSocket connection attempt - room: %s, token: %s", room, func() string {
		if len(token) > 20 {
			return token[:20] + "..."
		}
		return token
	}())

	if room == "" {
		log.Println("WebSocket error: room parameter missing")
		http.Error(w, "Room parameter is required", http.StatusBadRequest)
		return
	}

	if token == "" {
		log.Println("WebSocket error: token parameter missing")
		http.Error(w, "Token parameter is required", http.StatusUnauthorized)
		return
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		log.Printf("WebSocket error: invalid token - %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("WebSocket token valid, uid: %s", claims.UID)

	var user entity.User
	if err := database.DB.Where("uid = ?", claims.UID).First(&user).Error; err != nil {
		log.Printf("WebSocket error: user not found - %v", err)
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

		log.Printf("Received message from %s: %s", c.Username, message)

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "private_message":
			handlePrivateMessage(c, &wsMsg)
		case "group_message":
			handleGroupMessage(c, &wsMsg)
		case "read_receipt":
			handleReadReceipt(c, &wsMsg)
		default:
			handleRoomMessage(c, message)
		}
	}
}

func handleReadReceipt(c *Conn, msg *WSMessage) {
	if msg.MsgID == "" {
		log.Printf("Invalid read receipt: %+v", msg)
		return
	}

	log.Printf("Received read receipt from %s for msg: %s", c.Username, msg.MsgID)

	readRepo := infra_repo.NewUserRoomReadRepository()
	lastReadTime := time.Now().UnixMilli()

	roomID := msg.RoomID
	if roomID == "" && msg.FromUID != "" {
		roomID = msg.FromUID
	}
	if roomID == "" {
		log.Printf("Read receipt missing room_id or from_uid")
		return
	}

	if err := readRepo.Upsert(c.UserID, roomID, msg.MsgID, lastReadTime); err != nil {
		log.Printf("Failed to save read receipt: %v", err)
	}

	receipt := map[string]interface{}{
		"type":      "read_receipt",
		"msg_id":    msg.MsgID,
		"read_uid":  c.UserID,
		"room_id":   roomID,
		"read_time": lastReadTime,
	}

	var targetMsg entity.PrivateMessage
	if err := database.DB.Where("msg_id = ?", msg.MsgID).First(&targetMsg).Error; err == nil && targetMsg.FromUID != c.UserID {
		receiptBytes, _ := json.Marshal(receipt)
		log.Printf("Pushing read receipt to sender: %s", targetMsg.FromUID)
		BroadcastToUser(targetMsg.FromUID, receiptBytes)
	}
}

func handleRoomMessage(c *Conn, message []byte) {
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

func handlePrivateMessage(c *Conn, msg *WSMessage) {
	if msg.ToUID == "" || msg.Content == "" {
		log.Printf("Invalid private message: %+v", msg)
		return
	}

	msgID := idgen.Gen.NextUID()
	sendTime := time.Now().UnixMilli()

	newMsg := &entity.PrivateMessage{
		MsgID:    msgID,
		FromUID:  c.UserID,
		ToUID:    msg.ToUID,
		Content:  msg.Content,
		Type:     1,
		SendTime: sendTime,
		Read:     0,
	}

	msgRepo := infra_repo.NewPrivateMessageRepository()
	if err := msgRepo.Create(newMsg); err != nil {
		log.Printf("Failed to save private message: %v", err)
		return
	}

	var fromUser entity.User
	if err := database.DB.Where("uid = ?", c.UserID).First(&fromUser).Error; err != nil {
		fromUser.Username = "未知用户"
		fromUser.Nickname = "未知用户"
	}

	wsMsg := map[string]interface{}{
		"type":          "private_message",
		"msg_id":        newMsg.MsgID,
		"from_uid":      c.UserID,
		"from_username": fromUser.Username,
		"from_nickname": fromUser.Nickname,
		"to_uid":        msg.ToUID,
		"content":       msg.Content,
		"send_time":     sendTime,
	}

	SendPrivateMessage(msg.ToUID, wsMsg)
	SendPrivateMessage(c.UserID, wsMsg)
	log.Printf("Private message saved and sent: from=%s to=%s", c.UserID, msg.ToUID)
}

func handleGroupMessage(c *Conn, msg *WSMessage) {
	if msg.GroupID == "" || msg.Content == "" {
		log.Printf("Invalid group message: %+v", msg)
		return
	}

	groupRepo := infra_repo.NewGroupRepository()
	if _, err := groupRepo.FindMember(msg.GroupID, c.UserID); err != nil {
		log.Printf("group message denied: not a member group=%s uid=%s", msg.GroupID, c.UserID)
		return
	}

	msgID := idgen.GenerateMsgID()
	sendTime := time.Now().UnixMilli()

	newMsg := &entity.GroupMessage{
		MsgID:    msgID,
		GroupID:  msg.GroupID,
		FromUID:  c.UserID,
		Content:  msg.Content,
		Type:     2,
		SendTime: sendTime,
	}

	if err := groupRepo.CreateMessage(newMsg); err != nil {
		log.Printf("Failed to save group message: %v", err)
		return
	}

	var fromUser entity.User
	if err := database.DB.Where("uid = ?", c.UserID).First(&fromUser).Error; err != nil {
		fromUser.Username = "未知用户"
		fromUser.Nickname = "未知用户"
	}

	wsPayload := map[string]interface{}{
		"type":          "group_message",
		"msg_id":        newMsg.MsgID,
		"group_id":      msg.GroupID,
		"from_uid":      c.UserID,
		"from_username": fromUser.Username,
		"from_nickname": fromUser.Nickname,
		"content":       msg.Content,
		"send_time":     sendTime,
	}

	msgBytes, err := json.Marshal(wsPayload)
	if err != nil {
		log.Printf("Failed to marshal group message: %v", err)
		return
	}

	members, err := groupRepo.FindMembersByGroupID(msg.GroupID)
	if err != nil {
		log.Printf("Failed to list group members: %v", err)
		return
	}

	for _, m := range members {
		if m.UID == c.UserID {
			continue
		}
		sendBytesToUser(m.UID, msgBytes)
	}
	log.Printf("Group message broadcast: group=%s from=%s members=%d", msg.GroupID, c.UserID, len(members))
}

func sendBytesToUser(uid string, msgBytes []byte) {
	mu.Lock()
	conn, exists := userConns[uid]
	mu.Unlock()

	if !exists || conn == nil || conn.Closed {
		return
	}
	select {
	case conn.SendChan <- msgBytes:
	default:
		log.Printf("send queue full for uid=%s", uid)
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

func SendPrivateMessage(toUID string, msg map[string]interface{}) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal private message: %v", err)
		return
	}

	log.Printf("Attempting to send private message to user: %s", toUID)

	mu.Lock()
	conn, exists := userConns[toUID]
	log.Printf("User %s exists in connections: %v, closed: %v", toUID, exists, exists && conn != nil && conn.Closed)
	mu.Unlock()

	if exists && conn != nil && !conn.Closed {
		select {
		case conn.SendChan <- msgBytes:
			log.Printf("✅ Private message sent to user: %s", toUID)
		default:
			log.Printf("❌ Failed to send private message to %s: send chan full", toUID)
		}
	} else {
		log.Printf("❌ User %s is not connected, cannot send private message", toUID)
	}
}

func BroadcastToUser(uid string, msg []byte) {
	mu.Lock()
	conn, exists := userConns[uid]
	mu.Unlock()

	if exists && !conn.Closed {
		select {
		case conn.SendChan <- msg:
		default:
		}
	}
}
