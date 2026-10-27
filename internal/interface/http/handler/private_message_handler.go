package handler

import (
	message_service "im-chat/internal/application/message"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	"im-chat/internal/interface/http/response"
	ws "im-chat/internal/interface/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PrivateMessageHandler struct {
	msgRepo  domain_repo.PrivateMessageRepository
	readRepo domain_repo.UserRoomReadRepository
}

func NewPrivateMessageHandler(msgRepo domain_repo.PrivateMessageRepository, readRepo domain_repo.UserRoomReadRepository) *PrivateMessageHandler {
	return &PrivateMessageHandler{
		msgRepo:  msgRepo,
		readRepo: readRepo,
	}
}

type SendPrivateMessageReq struct {
	ToUID       string `json:"to_uid" binding:"required"`
	Content     string `json:"content" binding:"required"`
	ClientMsgID string `json:"client_msg_id" binding:"required"`
}

func (h *PrivateMessageHandler) SendMessage(c *gin.Context) {
	uid, _ := c.Get("uid")
	var req SendPrivateMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "参数错误："+err.Error())
		return
	}
	in := message_service.Input{
		Kind: "private", SenderUID: uid.(string), TargetID: req.ToUID,
		ClientMsgID: req.ClientMsgID, Content: req.Content,
	}
	result, err := message_service.Send(c.Request.Context(), database.DB, in)
	if err != nil {
		response.Failed(c, "消息未保存，请检查参数后重试")
		return
	}
	ws.PushAcceptedMessage(in, result, false)
	response.Success(c, gin.H{
		"msg_id": result.MsgID, "seq": result.Seq, "send_time": result.SendTime,
		"client_msg_id": req.ClientMsgID, "from_uid": in.SenderUID, "to_uid": req.ToUID,
		"content": req.Content, "acked": true,
	})
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

	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		isRead := msg.Read == 1
		isAcked := true
		result = append(result, map[string]interface{}{
			"msg_id":    msg.MsgID,
			"from_uid":  msg.FromUID,
			"to_uid":    msg.ToUID,
			"content":   msg.Content,
			"type":      msg.Type,
			"send_time": msg.SendTime,
			"is_read":   isRead,
			"acked":     isAcked,
		})
	}

	response.Success(c, result)
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

func (h *PrivateMessageHandler) GetUnreadCounts(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	countMap, err := h.msgRepo.GetUnreadCountPerFriend(currentUID)
	if err != nil {
		response.Failed(c, "获取未读消息数失败")
		return
	}

	response.Success(c, countMap)
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
