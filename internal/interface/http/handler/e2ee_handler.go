package handler

import (
	"log"

	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
)

type E2EEHandler struct {
	userRepo domain_repo.UserRepository
}

func NewE2EEHandler(userRepo domain_repo.UserRepository) *E2EEHandler {
	return &E2EEHandler{userRepo: userRepo}
}

type RegisterPubKeyReq struct {
	PubKey string `json:"pubkey" binding:"required"`
}

// RegisterPubKey 注册当前用户的 E2EE 公钥（公开信息，用于好友间密钥交换）
func (h *E2EEHandler) RegisterPubKey(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	var req RegisterPubKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	user, err := h.userRepo.FindByUID(currentUID)
	if err != nil {
		response.Failed(c, "用户不存在")
		return
	}

	user.PubKey = req.PubKey
	if err := h.userRepo.Update(user); err != nil {
		log.Printf("Failed to register pubkey for %s: %v", currentUID, err)
		response.Failed(c, "注册失败")
		return
	}

	response.Success(c, nil)
}
