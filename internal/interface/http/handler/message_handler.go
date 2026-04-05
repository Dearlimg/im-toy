package handler

import (
	"encoding/json"
	"im-chat/internal/domain/entity"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	infra_repo "im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/response"
	ws "im-chat/internal/interface/websocket"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	readRepo domain_repo.UserRoomReadRepository
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		readRepo: infra_repo.NewUserRoomReadRepository(),
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

	lastReadTime := time.Now().UnixMilli()
	if err := h.readRepo.Upsert(currentUID, req.RoomID, req.LastReadMsgID, lastReadTime); err != nil {
		response.Failed(c, "上报已读失败")
		return
	}

	go h.sendReadReceiptToSender(currentUID, req.RoomID, req.LastReadMsgID)

	response.Success(c, nil)
}

func (h *MessageHandler) sendReadReceiptToSender(readerUID, roomID, lastReadMsgID string) {
	var msg entity.PrivateMessage
	err := database.DB.Where("msg_id = ?", lastReadMsgID).First(&msg).Error
	if err != nil || msg.FromUID == readerUID {
		return
	}

	receipt := map[string]interface{}{
		"type":       "read_receipt",
		"room_id":    roomID,
		"reader_uid": readerUID,
		"msg_id":     lastReadMsgID,
		"read_time":  time.Now().UnixMilli(),
	}

	receiptBytes, _ := json.Marshal(receipt)
	ws.BroadcastToUser(msg.FromUID, receiptBytes)
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

	record, err := h.readRepo.FindByUIDAndRoomID(currentUID, roomID)
	if err != nil {
		response.Success(c, gin.H{"last_read_msg_id": "", "last_read_time": 0})
		return
	}

	response.Success(c, gin.H{
		"last_read_msg_id": record.LastReadMsgID,
		"last_read_time":   record.LastReadTime,
	})
}
