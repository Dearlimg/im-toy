package handler

import (
	"im-chat/internal/domain/entity"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	infra_repo "im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendRepo domain_repo.FriendRepository
}

func NewFriendHandler() *FriendHandler {
	return &FriendHandler{
		friendRepo: infra_repo.NewFriendRepository(),
	}
}

type AddFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
	Remark    string `json:"remark"`
}

func (h *FriendHandler) AddFriend(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	var req AddFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	if currentUID == req.FriendUID {
		response.Failed(c, "不能添加自己为好友")
		return
	}

	var friendUser entity.User
	if err := database.DB.Where("uid = ?", req.FriendUID).First(&friendUser).Error; err != nil {
		response.Failed(c, "用户不存在")
		return
	}

	existing, _ := h.friendRepo.FindByUIDAndFriendUID(currentUID, req.FriendUID)
	if existing != nil {
		if existing.Status == 1 {
			response.Failed(c, "已经是好友了")
			return
		} else if existing.Status == 2 {
			response.Failed(c, "已发送好友请求，等待对方验证")
			return
		}
	}

	friend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Remark:    req.Remark,
		Status:    2,
	}

	if err := h.friendRepo.Create(friend); err != nil {
		response.Failed(c, "添加好友失败")
		return
	}

	response.Success(c, nil)
}

type AcceptFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
}

func (h *FriendHandler) AcceptFriend(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	var req AcceptFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	friendRequest, err := h.friendRepo.FindByUIDAndFriendUID(req.FriendUID, currentUID)
	if err != nil || friendRequest.Status != 2 {
		response.Failed(c, "好友请求不存在")
		return
	}

	friendRequest.Status = 1
	if err := h.friendRepo.Update(friendRequest); err != nil {
		response.Failed(c, "接受好友失败")
		return
	}

	reverseFriend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Status:    1,
	}
	if err := h.friendRepo.Create(reverseFriend); err != nil {
		response.Failed(c, "添加好友失败")
		return
	}

	response.Success(c, nil)
}

func (h *FriendHandler) GetFriendList(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	friends, err := h.friendRepo.FindFriendsWithUserInfo(currentUID)
	if err != nil {
		response.Failed(c, "获取好友列表失败")
		return
	}

	response.Success(c, friends)
}

func (h *FriendHandler) GetFriendRequests(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	requests, err := h.friendRepo.FindFriendRequests(currentUID)
	if err != nil {
		response.Failed(c, "获取好友请求失败")
		return
	}

	response.Success(c, requests)
}

type DeleteFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
}

func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	uid, _ := c.Get("uid")
	currentUID := uid.(string)

	var req DeleteFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	if err := h.friendRepo.Delete(currentUID, req.FriendUID); err != nil {
		response.Failed(c, "删除好友失败")
		return
	}

	response.Success(c, nil)
}

type SearchUserReq struct {
	Username string `json:"username" binding:"required"`
}

func (h *FriendHandler) SearchUser(c *gin.Context) {
	var req SearchUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, "参数错误："+err.Error())
		return
	}

	var users []entity.User
	if err := database.DB.Where("username LIKE ?", "%"+req.Username+"%").Limit(10).Find(&users).Error; err != nil {
		response.Failed(c, "搜索用户失败")
		return
	}

	response.Success(c, users)
}
