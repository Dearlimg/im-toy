package service

import (
	"im-chat/internal/application/dto"
)

type UserService interface {
	Register(req *dto.UserRegisterReq) (*dto.UserLoginResp, error)
	Login(req *dto.UserLoginReq) (*dto.UserLoginResp, error)
}
