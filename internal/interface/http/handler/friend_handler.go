package handler

import (
	"im-chat/internal/apperrors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendService service.FriendService
}

func NewFriendHandler(friendService service.FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

func (h *FriendHandler) SearchUser(c *gin.Context) {
	var req dto.SearchUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	resp, err := h.friendService.SearchUser(currentUID, req.Keyword)
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

func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	var req dto.DeleteFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	if err := h.friendService.DeleteFriend(currentUID, &req); err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

func (h *FriendHandler) AddFriend(c *gin.Context) {
	var req dto.AddFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	if err := h.friendService.AddFriend(currentUID, &req); err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

func (h *FriendHandler) AcceptFriend(c *gin.Context) {
	var req dto.AcceptFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	if err := h.friendService.AcceptFriend(currentUID, &req); err != nil {
		if appErr, ok := apperrors.FromError(err); ok {
			response.AppError(c, appErr)
		} else {
			response.Failed(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

func (h *FriendHandler) GetFriendList(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	resp, err := h.friendService.GetFriendList(currentUID)
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

func (h *FriendHandler) GetFriendRequests(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	resp, err := h.friendService.GetFriendRequests(currentUID)
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
