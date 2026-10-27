package handler

import (
	"im-chat/internal/application/readstate"
	"im-chat/internal/domain/entity"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	"im-chat/internal/interface/http/response"
	ws "im-chat/internal/interface/websocket"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	readRepo domain_repo.UserRoomReadRepository
}

func NewMessageHandler(readRepo domain_repo.UserRoomReadRepository) *MessageHandler {
	return &MessageHandler{
		readRepo: readRepo,
	}
}

type ReportReadReq struct {
	RoomID        string `json:"room_id" binding:"required"`
	LastReadMsgID string `json:"last_read_msg_id" binding:"required"`
}

func (h *MessageHandler) ReportRead(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	var req ReportReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	msg, err := readstate.MarkPrivate(database.DB, currentUID, req.RoomID, req.LastReadMsgID)
	if err != nil {
		response.Failed(c, "上报已读失败")
		return
	}
	ws.SendToUser(msg.FromUID, map[string]interface{}{
		"type": "read_receipt", "room_id": req.RoomID, "read_uid": currentUID,
		"msg_id": req.LastReadMsgID, "read_time": time.Now().UnixMilli(),
	})

	response.Success(c, nil)
}

type GetUnreadCountReq struct {
	RoomID string `form:"room_id" binding:"required"`
}

func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	roomID := c.Query("room_id")
	if roomID == "" {
		response.Failed(c, "缺少room_id参数")
		return
	}

	lastMsgID := c.Query("last_msg_id")
	count, err := h.readRepo.GetUnreadCount(currentUID, roomID, lastMsgID)
	if err != nil {
		response.Failed(c, "获取未读数失败")
		return
	}

	response.Success(c, gin.H{"count": count})
}

type GetRoomReadStatusReq struct {
	RoomID string `form:"room_id" binding:"required"`
}

func (h *MessageHandler) GetRoomReadStatus(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	roomID := c.Query("room_id")
	if roomID == "" {
		response.Failed(c, "缺少room_id参数")
		return
	}

	var record entity.PrivateMessage
	err := database.DB.Where("from_uid = ? AND to_uid = ? AND `read` = ?", roomID, currentUID, 1).
		Order("seq DESC, id DESC").First(&record).Error
	if err != nil {
		response.Success(c, gin.H{"last_read_msg_id": "", "last_read_time": 0})
		return
	}

	response.Success(c, gin.H{
		"last_read_msg_id": record.MsgID,
		"last_read_time":   record.UpdatedAt.UnixMilli(),
	})
}
