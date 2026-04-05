package handler

import (
	"im-chat/internal/apperrors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/interface/http/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService service.GroupService
}

func NewGroupHandler(groupService service.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req dto.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	resp, err := h.groupService.CreateGroup(currentUID, &req)
	if err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

func (h *GroupHandler) GetGroupUnreadMessages(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	groupID := c.Query("group_id")
	if groupID == "" {
		response.InvalidParams(c, "group_id is required")
		return
	}

	lastReadMsgID := c.Query("last_read_msg_id")

	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.groupService.GetGroupUnreadMessages(currentUID, groupID, lastReadMsgID, limit)
	if err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

func (h *GroupHandler) GetGroupList(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	resp, err := h.groupService.GetGroupList(currentUID)
	if err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	groupID := c.Query("group_id")
	if groupID == "" {
		response.InvalidParams(c, "group_id is required")
		return
	}

	resp, err := h.groupService.GetGroupMembers(currentUID, groupID)
	if err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

func (h *GroupHandler) GetGroupMessages(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	groupID := c.Query("group_id")
	if groupID == "" {
		response.InvalidParams(c, "group_id is required")
		return
	}

	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.groupService.GetGroupMessages(currentUID, groupID, limit)
	if err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, resp)
}
