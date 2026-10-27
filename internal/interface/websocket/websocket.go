package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	message_service "im-chat/internal/application/message"
	"im-chat/internal/application/readstate"
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/infra/kafka"
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
		expired := []*Conn{}
		mu.Lock()
		for _, conn := range userConns {
			if now-atomic.LoadInt64(&conn.LastBeatTime) > ConnExpireSeconds {
				expired = append(expired, conn)
			}
		}
		mu.Unlock()
		for _, conn := range expired {
			conn.cleanup()
		}
	}
}

// notifyFriendsPresence 向该用户的所有好友推送在线状态变更事件
func notifyFriendsPresence(uid string, online bool) {
	friendRepo := infra_repo.NewFriendRepository()
	friends, err := friendRepo.FindByUID(uid)
	if err != nil {
		return
	}
	presence := map[string]interface{}{
		"type":   "presence",
		"uid":    uid,
		"online": online,
	}
	for _, f := range friends {
		SendToUser(f.FriendUID, presence)
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
	Type        string `json:"type"`
	MsgID       string `json:"msg_id,omitempty"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
	FromUID     string `json:"from_uid,omitempty"`
	ToUID       string `json:"to_uid,omitempty"`
	RoomID      string `json:"room_id,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	Content     string `json:"content,omitempty"`

	MsgIDs []string `json:"msg_ids,omitempty"`

	Timestamp int64 `json:"timestamp,omitempty"`
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
		closeConnLocked(oldConn)
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

	notifyFriendsPresence(user.UID, true)

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
			if redis.Cache != nil {
				redis.Cache.RefreshUserOnline(c.UserID)
			}
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
	sendToConn(c, data)
}

func handlePrivateMessage(c *Conn, msg *WSMessage) {
	if redis.Cache != nil {
		count, err := redis.Cache.IncrMsgLimit(c.UserID)
		if err == nil && count > MsgRateLimit {
			sendFailure(c, msg.ClientMsgID, "发送过于频繁")
			return
		}
	}
	sendMessage(c, msg, "private", msg.ToUID)
}

func handleGroupMessage(c *Conn, msg *WSMessage) {
	sendMessage(c, msg, "group", msg.GroupID)
}

func sendMessage(c *Conn, msg *WSMessage, kind, target string) {
	in := message_service.Input{
		Kind: kind, SenderUID: c.UserID, TargetID: target,
		ClientMsgID: msg.ClientMsgID, Content: msg.Content,
	}
	result, err := message_service.Send(context.Background(), database.DB, in)
	if err != nil {
		log.Printf("message acceptance failed for uid=%s: %v", c.UserID, err)
		sendFailure(c, msg.ClientMsgID, "消息未保存，请重试")
		return
	}
	PushAcceptedMessage(in, result, true)
}

func sendFailure(c *Conn, clientMsgID, reason string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"type": "send_failed", "client_msg_id": clientMsgID, "error": reason,
	})
	sendToConn(c, payload)
}

// PushAcceptedMessage delivers a committed message to currently connected users.
// History remains the source of truth when a user is offline or delivery fails.
func PushAcceptedMessage(in message_service.Input, result message_service.Result, notifySender bool) {
	payload := map[string]interface{}{
		"type": in.Kind + "_message", "client_msg_id": in.ClientMsgID,
		"msg_id": result.MsgID, "seq": result.Seq, "from_uid": in.SenderUID,
		"content": in.Content, "send_time": result.SendTime, "acked": true,
	}
	if in.Kind == "private" {
		payload["to_uid"] = in.TargetID
	} else {
		payload["group_id"] = in.TargetID
	}
	if redis.Cache != nil {
		if info, _ := redis.Cache.GetUserInfo(in.SenderUID); info != nil {
			payload["from_username"] = info.Username
			payload["from_nickname"] = info.Nickname
		}
	}
	if notifySender {
		SendToUser(in.SenderUID, payload)
	}
	if result.Duplicate {
		return
	}
	if in.Kind == "private" {
		SendToUser(in.TargetID, payload)
		return
	}
	members, err := infra_repo.NewGroupRepository().FindMembersByGroupID(in.TargetID)
	if err != nil {
		log.Printf("failed to find group recipients for group=%s: %v", in.TargetID, err)
		return
	}
	for _, member := range members {
		if member.UID != in.SenderUID {
			SendToUser(member.UID, payload)
		}
	}
}

func handleReadReceipt(c *Conn, msg *WSMessage) {
	roomID := msg.RoomID
	if roomID == "" {
		roomID = msg.FromUID
	}
	target, err := readstate.MarkPrivate(database.DB, c.UserID, roomID, msg.MsgID)
	if err != nil {
		log.Printf("read receipt rejected: %v", err)
		return
	}
	receipt := map[string]interface{}{
		"type":      "read_receipt",
		"msg_id":    msg.MsgID,
		"read_uid":  c.UserID,
		"room_id":   roomID,
		"read_time": time.Now().UnixMilli(),
	}
	SendToUser(target.FromUID, receipt)
}

func handleGroupReadReceipt(c *Conn, msg *WSMessage) {
	messages, err := readstate.MarkGroup(database.DB, c.UserID, msg.GroupID, msg.MsgIDs)
	if err != nil {
		log.Printf("group read receipt rejected: %v", err)
		return
	}
	readTime := time.Now().UnixMilli()
	groupRepo := infra_repo.NewGroupRepository()
	members, err := groupRepo.FindMembersByGroupID(msg.GroupID)
	if err != nil {
		return
	}
	totalCount := len(members) - 1
	for _, groupMsg := range messages {
		if groupMsg.FromUID == c.UserID {
			continue
		}
		var readCount int64
		if err := database.DB.Model(&entity.MessageRead{}).Where("msg_id = ?", groupMsg.MsgID).Count(&readCount).Error; err != nil {
			continue
		}
		receipt := map[string]interface{}{
			"type":        "group_read_receipt",
			"group_id":    msg.GroupID,
			"msg_id":      groupMsg.MsgID,
			"read_uid":    c.UserID,
			"read_count":  readCount,
			"total_count": totalCount,
			"read_time":   readTime,
		}
		SendToUser(groupMsg.FromUID, receipt)
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

	mu.RLock()
	for conn := range rooms[c.RoomID] {
		select {
		case conn.SendChan <- msgBytes:
		default:
			log.Printf("send queue full for uid=%s", conn.UserID)
			// Force a reconnect so the client can repair the gap through sync.
			if conn.Conn != nil {
				conn.Conn.Close()
			}
		}
	}
	mu.RUnlock()
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
	defer mu.RUnlock()
	conn, exists := userConns[uid]

	if exists && !conn.Closed {
		select {
		case conn.SendChan <- msgBytes:
		default:
			log.Printf("send queue full for uid=%s", uid)
			if conn.Conn != nil {
				conn.Conn.Close()
			}
		}
	} else {
		if !exists {
			log.Printf("user not connected: uid=%s", uid)
		} else {
			log.Printf("user connection closed: uid=%s", uid)
		}
	}
}

// sendToConn holds the same lock used when closing a send channel.
func sendToConn(conn *Conn, data []byte) bool {
	mu.RLock()
	defer mu.RUnlock()
	if conn.Closed {
		return false
	}
	select {
	case conn.SendChan <- data:
		return true
	default:
		if conn.Conn != nil {
			conn.Conn.Close()
		}
		return false
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
	if c.Closed {
		mu.Unlock()
		return
	}
	current := userConns[c.UserID] == c
	closeConnLocked(c)
	mu.Unlock()
	if !current {
		return
	}

	if redis.Cache != nil {
		cache := redis.Cache
		cache.SetUserOffline(c.UserID)
		cache.RemoveConn(c.UserID)
	}

	notifyFriendsPresence(c.UserID, false)
	log.Printf("User %s left room: %s", c.Username, c.RoomID)
}

// closeConnLocked detaches and closes a connection while holding mu.
func closeConnLocked(c *Conn) {
	if c.Closed {
		return
	}
	c.Closed = true
	close(c.SendChan)
	delete(rooms[c.RoomID], c)
	if len(rooms[c.RoomID]) == 0 {
		delete(rooms, c.RoomID)
	}
	if userConns[c.UserID] == c {
		delete(userConns, c.UserID)
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
}

type WSPusher struct{}

func (p *WSPusher) PushToUser(uid string, msg map[string]interface{}) {
	SendToUser(uid, msg)
}

var wsPusherInstance = &WSPusher{}

func init() {
	kafka.RegisterPusher(wsPusherInstance)
}
