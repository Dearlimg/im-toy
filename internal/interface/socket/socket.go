package socket

import (
	"errors"
	"im-chat/internal/infra/jwt"
	"log"
	"strconv"

	socketio "github.com/googollee/go-socket.io"
)

func NewSocketServer() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		// Get token from query parameter
		url := s.URL()
		token := url.Query().Get("token")
		if token == "" {
			return errors.New("missing token")
		}

		// Parse token
		claims, err := jwt.ParseAccessToken(token)
		if err != nil {
			return err
		}

		// Store user ID in socket context
		s.SetContext(claims.UserID)

		// Join personal room for private chat
		s.Join("user_" + strconv.FormatInt(claims.UserID, 10))

		log.Printf("User %d connected", claims.UserID)
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		userID := s.Context().(int64)
		log.Printf("User %d disconnected: %s", userID, reason)
	})

	// Private chat event
	server.OnEvent("/", "private_chat", func(s socketio.Conn, msg map[string]interface{}) {
		userID := s.Context().(int64)
		recvID := int64(msg["recv_id"].(float64))
		content := msg["content"].(string)

		// Send message to target user
		server.BroadcastToRoom("/", "user_"+strconv.FormatInt(recvID, 10), "new_msg", map[string]interface{}{
			"send_id": userID,
			"recv_id": recvID,
			"content": content,
			"type":    1, // 1 for private chat
		})
	})

	// Group chat event
	server.OnEvent("/", "group_chat", func(s socketio.Conn, msg map[string]interface{}) {
		userID := s.Context().(int64)
		groupID := int64(msg["group_id"].(float64))
		content := msg["content"].(string)

		// Send message to all members in group room
		server.BroadcastToRoom("/", "group_"+strconv.FormatInt(groupID, 10), "new_msg", map[string]interface{}{
			"send_id":  userID,
			"group_id": groupID,
			"content":  content,
			"type":     2, // 2 for group chat
		})
	})

	return server
}
