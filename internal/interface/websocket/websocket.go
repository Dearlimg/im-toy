package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/idgen"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/infra/redis"
	infra_repo "im-chat/internal/infra/repository"

	"github.com/gorilla/websocket"
)

const (
	HeartbeatInterval = 30 * time.Second
	MsgRateLimit      = 60
	ConnExpireSeconds = 60
	CleanupInterval   = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Conn struct {
	Conn         *websocket.Conn
	UserID       string
	Username     string
	RoomID       string
	SendChan     chan []byte
	Closed       bool
	LastBeatTime int64
}

var rooms = make(map[string]map[*Conn]bool)
var userConns = make(map[string]*Conn)
var mu sync.RWMutex

func init() {
	go cleanupExpiredConn()
}

func cleanupExpiredConn() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Unix()
		mu.Lock()
		for uid, conn := range userConns {
			if now-atomic.LoadInt64(&conn.LastBeatTime) > ConnExpireSeconds {
				if !conn.Closed {
					conn.Closed = true
					close(conn.SendChan)
					conn.Conn.Close()
				}
				delete(userConns, uid)
				if redis.Cache != nil {
					redis.Cache.SetUserOffline(uid)
					redis.Cache.RemoveConn(uid)
				}
				log.Printf("Cleaned up expired connection: %s", uid)
			}
		}
		mu.Unlock()
	}
}

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
	Type      string   `json:"type"`
	MsgID     string   `json:"msg_id,omitempty"`
	FromUID   string   `json:"from_uid,omitempty"`
	ToUID     string   `json:"to_uid,omitempty"`
	RoomID    string   `json:"room_id,omitempty"`
	GroupID   string   `json:"group_id,omitempty"`
	Content   string   `json:"content,omitempty"`
	MsgIDs    []string `json:"msg_ids,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	token := r.URL.Query().Get("token")

	if room == "" || token == "" {
		http.Error(w, "Room and token are required", http.StatusBadRequest)
		return
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var user entity.User
	userID := claims.UID

	if redis.Cache != nil {
		userInfo, _ := redis.Cache.GetUserInfo(userID)
		if userInfo != nil {
			user.UID = userInfo.UID
			user.Username = userInfo.Username
			user.Nickname = userInfo.Nickname
			user.Avatar = userInfo.Avatar
		}
	}

	if user.UID == "" {
		if err := database.DB.Where("uid = ?", userID).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		if redis.Cache != nil {
			redis.Cache.SetUserInfo(userID, &redis.UserBasicInfo{
				UID:      user.UID,
				Username: user.Username,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
			})
		}
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}

	conn := &Conn{
		Conn:         wsConn,
		UserID:       user.UID,
		Username:     user.Username,
		RoomID:       room,
		SendChan:     make(chan []byte, 256),
		Closed:       false,
		LastBeatTime: time.Now().Unix(),
	}

	mu.Lock()
	if _, exists := rooms[room]; !exists {
		rooms[room] = make(map[*Conn]bool)
	}
	rooms[room][conn] = true
	if oldConn, exists := userConns[user.UID]; exists {
		close(oldConn.SendChan)
		oldConn.Closed = true
		oldConn.Conn.Close()
	}
	userConns[user.UID] = conn
	mu.Unlock()

	if redis.Cache != nil {
		cache := redis.Cache
		cache.SetUserOnline(user.UID)
		cache.SetConn(user.UID, &redis.ConnInfo{
			UID:     user.UID,
			Server:  "single",
			RoomID:  room,
			LoginAt: time.Now().Unix(),
		})
		cache.SetUserInfo(user.UID, &redis.UserBasicInfo{
			UID:      user.UID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	log.Printf("User %s joined room: %s", user.Username, room)

	go conn.writePump()
	go conn.heartbeat()
	conn.readPump()
}

func (c *Conn) heartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for !c.Closed {
		<-ticker.C
		if c.Closed {
			break
		}
		atomic.StoreInt64(&c.LastBeatTime, time.Now().Unix())
		if redis.Cache != nil {
			redis.Cache.RefreshUserOnline(c.UserID)
		}
	}
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

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "private_message":
			handlePrivateMessage(c, &wsMsg)
		case "group_message":
			handleGroupMessage(c, &wsMsg)
		case "read_receipt":
			handleReadReceipt(c, &wsMsg)
		case "group_read_receipt":
			handleGroupReadReceipt(c, &wsMsg)
		case "heartbeat":
			atomic.StoreInt64(&c.LastBeatTime, time.Now().Unix())
			sendHeartbeatAck(c)
		default:
			handleRoomMessage(c, message)
		}
	}
}

func sendHeartbeatAck(c *Conn) {
	ack := map[string]interface{}{
		"type": "heartbeat_ack",
		"ts":   time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(ack)
	select {
	case c.SendChan <- data:
	default:
	}
}

func handlePrivateMessage(c *Conn, msg *WSMessage) {
	if msg.ToUID == "" || msg.Content == "" {
		return
	}

	if redis.Cache != nil {
		count, _ := redis.Cache.IncrMsgLimit(c.UserID)
		if count > MsgRateLimit {
			log.Printf("Rate limit exceeded for user: %s", c.UserID)
			return
		}
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

	wsMsg := map[string]interface{}{
		"type":      "private_message",
		"msg_id":    newMsg.MsgID,
		"from_uid":  c.UserID,
		"to_uid":    msg.ToUID,
		"content":   msg.Content,
		"send_time": sendTime,
	}

	if redis.Cache != nil {
		userInfo, _ := redis.Cache.GetUserInfo(c.UserID)
		if userInfo != nil {
			wsMsg["from_username"] = userInfo.Username
			wsMsg["from_nickname"] = userInfo.Nickname
		}
	}

	SendToUser(msg.ToUID, wsMsg)
	SendToUser(c.UserID, wsMsg)
}

func handleGroupMessage(c *Conn, msg *WSMessage) {
	if msg.GroupID == "" || msg.Content == "" {
		return
	}

	groupRepo := infra_repo.NewGroupRepository()
	if _, err := groupRepo.FindMember(msg.GroupID, c.UserID); err != nil {
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
		return
	}

	wsPayload := map[string]interface{}{
		"type":      "group_message",
		"msg_id":    newMsg.MsgID,
		"group_id":  msg.GroupID,
		"from_uid":  c.UserID,
		"content":   msg.Content,
		"send_time": sendTime,
	}

	if redis.Cache != nil {
		userInfo, _ := redis.Cache.GetUserInfo(c.UserID)
		if userInfo != nil {
			wsPayload["from_username"] = userInfo.Username
			wsPayload["from_nickname"] = userInfo.Nickname
		}
	}

	var members []string
	if redis.Cache != nil {
		members, _ = redis.Cache.GetGroupMembers(msg.GroupID)
	}

	if len(members) == 0 {
		dbMembers, _ := groupRepo.FindMembersByGroupID(msg.GroupID)
		members = make([]string, 0, len(dbMembers))
		for _, m := range dbMembers {
			members = append(members, m.UID)
		}
		if redis.Cache != nil && len(members) > 0 {
			redis.Cache.SetGroupMembers(msg.GroupID, members)
		}
	}

	wsPayload["total_count"] = len(members)

	msgBytes, _ := json.Marshal(wsPayload)
	for _, m := range members {
		SendBytesToUser(m, msgBytes)
	}
}

func handleReadReceipt(c *Conn, msg *WSMessage) {
	if msg.MsgID == "" {
		return
	}

	roomID := msg.RoomID
	if roomID == "" && msg.FromUID != "" {
		roomID = msg.FromUID
	}
	if roomID == "" {
		return
	}

	lastReadTime := time.Now().UnixMilli()

	readRepo := infra_repo.NewUserRoomReadRepository()
	readRepo.Upsert(c.UserID, roomID, msg.MsgID, lastReadTime)

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
		SendBytesToUser(targetMsg.FromUID, receiptBytes)
	}
}

func handleGroupReadReceipt(c *Conn, msg *WSMessage) {
	if msg.GroupID == "" || len(msg.MsgIDs) == 0 {
		return
	}

	readTime := time.Now().UnixMilli()
	readRepo := infra_repo.NewMessageReadRepository()
	groupRepo := infra_repo.NewGroupRepository()

	messageReads := make([]*entity.MessageRead, 0, len(msg.MsgIDs))
	for _, msgID := range msg.MsgIDs {
		messageReads = append(messageReads, &entity.MessageRead{
			MsgID:    msgID,
			UID:      c.UserID,
			ReadTime: readTime,
		})
	}

	readRepo.BatchCreate(messageReads)

	members, _ := groupRepo.FindMembersByGroupID(msg.GroupID)
	totalCount := len(members)

	readCountMap, _ := readRepo.BatchGetReadCount(msg.MsgIDs)
	msgMap := make(map[string]*entity.GroupMessage)
	if len(msg.MsgIDs) > 0 {
		messages, _ := groupRepo.FindMessagesByMsgIDs(msg.MsgIDs)
		for i := range messages {
			msgMap[messages[i].MsgID] = messages[i]
		}
	}

	for _, msgID := range msg.MsgIDs {
		groupMsg, exists := msgMap[msgID]
		if !exists || groupMsg.FromUID == c.UserID {
			continue
		}

		receipt := map[string]interface{}{
			"type":        "group_read_receipt",
			"group_id":    msg.GroupID,
			"msg_id":      msgID,
			"read_uid":    c.UserID,
			"read_count":  readCountMap[msgID],
			"total_count": totalCount,
			"read_time":   readTime,
		}
		receiptBytes, _ := json.Marshal(receipt)
		SendBytesToUser(groupMsg.FromUID, receiptBytes)
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

func SendToUser(uid string, msg map[string]interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	SendBytesToUser(uid, data)
}

func SendBytesToUser(uid string, msgBytes []byte) {
	mu.RLock()
	conn, exists := userConns[uid]
	mu.RUnlock()

	if exists && !conn.Closed {
		select {
		case conn.SendChan <- msgBytes:
		default:
			log.Printf("send queue full for uid=%s", uid)
		}
	} else {
		if !exists {
			log.Printf("user not connected: uid=%s", uid)
		} else {
			log.Printf("user connection closed: uid=%s", uid)
		}
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

		if redis.Cache != nil {
			cache := redis.Cache
			cache.SetUserOffline(c.UserID)
			cache.RemoveConn(c.UserID)
		}

		log.Printf("User %s left room: %s", c.Username, c.RoomID)
	}
}
