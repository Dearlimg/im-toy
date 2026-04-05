package handler

import (
	"im-chat/internal/domain/entity"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/idgen"
	infra_repo "im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/response"
	"im-chat/internal/interface/websocket"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PrivateMessageHandler struct {
	msgRepo domain_repo.PrivateMessageRepository
}

func NewPrivateMessageHandler() *PrivateMessageHandler {
	return &PrivateMessageHandler{
		msgRepo: infra_repo.NewPrivateMessageRepository(),
	}
}

type SendPrivateMessageReq struct {
	ToUID   string `json:"to_uid" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *PrivateMessageHandler) SendMessage(c *gin.Context) {
	uid, _ := c.Get("uid")
	fromUID := uid.(string)

	var req SendPrivateMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	if fromUID == req.ToUID {
		response.Failed(c, "不能给自己发消息")
		return
	}

	msgID := idgen.GenerateMsgID()
	sendTime := time.Now().UnixMilli()

	msg := &entity.PrivateMessage{
		MsgID:    msgID,
		FromUID:  fromUID,
		ToUID:    req.ToUID,
		Content:  req.Content,
		Type:     1,
		SendTime: sendTime,
		Read:     0,
	}

	if err := h.msgRepo.Create(msg); err != nil {
		response.Failed(c, "发送消息失败")
		return
	}

	var fromUser entity.User
	if err := database.DB.Where("uid = ?", fromUID).First(&fromUser).Error; err != nil {
		fromUser.Username = "未知用户"
		fromUser.Nickname = "未知用户"
	}

	wsMsg := map[string]interface{}{
		"type":          "private_message",
		"msg_id":        msgID,
		"from_uid":      fromUID,
		"from_username": fromUser.Username,
		"from_nickname": fromUser.Nickname,
		"to_uid":        req.ToUID,
		"content":       req.Content,
		"send_time":     sendTime,
	}

	log.Printf("About to send WebSocket message to %s: %s", req.ToUID, req.Content)
	websocket.SendPrivateMessage(req.ToUID, wsMsg)

	response.Success(c, msg)
}

type GetConversationReq struct {
	FriendUID string `form:"friend_uid" binding:"required"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

func (h *PrivateMessageHandler) GetConversation(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	friendUID := c.Query("friend_uid")
	if friendUID == "" {
		response.Failed(c, "缺少friend_uid参数")
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if li, err := strconv.Atoi(l); err == nil && li > 0 {
			limit = li
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if oi, err := strconv.Atoi(o); err == nil && oi >= 0 {
			offset = oi
		}
	}

	messages, err := h.msgRepo.FindConversation(currentUID, friendUID, limit, offset)
	if err != nil {
		response.Failed(c, "获取消息失败")
		return
	}

	_ = h.msgRepo.MarkAsRead(friendUID, currentUID)

	response.Success(c, messages)
}

func (h *PrivateMessageHandler) GetUnreadCount(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	count, err := h.msgRepo.GetUnreadCount(currentUID)
	if err != nil {
		response.Failed(c, "获取未读消息数失败")
		return
	}

	response.Success(c, gin.H{"count": count})
}

func (h *PrivateMessageHandler) MarkAsRead(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	fromUID := c.Query("from_uid")
	if fromUID == "" {
		response.Failed(c, "缺少from_uid参数")
		return
	}

	if err := h.msgRepo.MarkAsRead(fromUID, currentUID); err != nil {
		response.Failed(c, "标记已读失败")
		return
	}

	response.Success(c, nil)
}
