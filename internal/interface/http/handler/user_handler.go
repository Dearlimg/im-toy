package handler

import (
	"im-chat/internal/apperrors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	resp, err := h.userService.Register(&req)
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

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	resp, err := h.userService.Login(&req)
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
