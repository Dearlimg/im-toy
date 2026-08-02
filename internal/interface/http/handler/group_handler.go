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

// GetGroupList 获取当前用户的群组列表
//
// 从上下文中获取当前用户ID，调用群组服务获取该用户加入的所有群组信息
//
// 参数:
//   - c: Gin上下文对象，包含请求信息和当前用户UID
//
// 返回:
//   - 成功: 返回群组列表数据
//   - 失败: 返回应用错误或系统错误信息
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

// GetGroupMessages 获取群组消息记录
// 根据group_id查询该群组的历史消息，支持自定义返回条数限制（默认50条）
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
