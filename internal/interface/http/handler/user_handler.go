package handler

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/idgen"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/interface/http/response"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=2,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname" binding:"required,min=2,max=20"`
}

type UserRegisterResp struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	var existingUser entity.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		response.Failed(c, "用户名已存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Failed(c, "密码加密失败")
		return
	}

	uid := idgen.Gen.NextUID()
	user := &entity.User{
		UID:      uid,
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := database.DB.Create(user).Error; err != nil {
		response.Failed(c, "注册失败")
		return
	}

	token, err := jwt.GenerateToken(uid)
	if err != nil {
		response.Failed(c, "生成Token失败")
		return
	}

	response.Success(c, UserRegisterResp{
		UID:      uid,
		Username: user.Username,
		Nickname: user.Nickname,
		Token:    token,
	})
}

type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResp struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	var user entity.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Failed(c, "用户不存在")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Failed(c, "密码错误")
		return
	}

	if user.Status != 1 {
		response.Failed(c, "账号已被禁用")
		return
	}

	token, err := jwt.GenerateToken(user.UID)
	if err != nil {
		response.Failed(c, "生成Token失败")
		return
	}

	response.Success(c, UserLoginResp{
		UID:      user.UID,
		Username: user.Username,
		Nickname: user.Nickname,
		Token:    token,
	})
}
