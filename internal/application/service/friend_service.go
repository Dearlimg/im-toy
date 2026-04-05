package service

import (
	"im-chat/internal/application/dto"
)

type FriendService interface {
	AddFriend(currentUID string, req *dto.AddFriendReq) error
	AcceptFriend(currentUID string, req *dto.AcceptFriendReq) error
	DeleteFriend(currentUID string, req *dto.DeleteFriendReq) error
	GetFriendList(currentUID string) ([]*dto.FriendResp, error)
	GetFriendRequests(currentUID string) ([]*dto.FriendResp, error)
	SearchUser(currentUID string, keyword string) ([]*dto.UserResp, error)
}
