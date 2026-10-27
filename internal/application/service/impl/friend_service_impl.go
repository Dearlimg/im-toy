package impl

import (
	"im-chat/internal/apperrors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/redis"
	"im-chat/internal/interface/http/response"
)

type FriendServiceImpl struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
}

func NewFriendServiceImpl(friendRepo repository.FriendRepository, userRepo repository.UserRepository) service.FriendService {
	return &FriendServiceImpl{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

func (s *FriendServiceImpl) AddFriend(currentUID string, req *dto.AddFriendReq) error {
	if currentUID == req.FriendUID {
		return apperrors.New(response.CodeCannotAddSelfAsFriend, response.GetMsg(response.CodeCannotAddSelfAsFriend))
	}

	_, err := s.userRepo.FindByUID(req.FriendUID)
	if err != nil {
		return apperrors.New(response.CodeUserNotFound, response.GetMsg(response.CodeUserNotFound))
	}

	existing, _ := s.friendRepo.FindByUIDAndFriendUID(currentUID, req.FriendUID)
	if existing != nil {
		if existing.Status == 1 {
			return apperrors.New(response.CodeFriendAlreadyExist, response.GetMsg(response.CodeFriendAlreadyExist))
		} else if existing.Status == 2 {
			return apperrors.New(response.CodeFriendRequestSent, response.GetMsg(response.CodeFriendRequestSent))
		}
	}

	friend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Remark:    req.Remark,
		Status:    2,
	}

	if err := s.friendRepo.Create(friend); err != nil {
		return apperrors.New(response.CodeFriendAddFailed, response.GetMsg(response.CodeFriendAddFailed))
	}

	return nil
}

func (s *FriendServiceImpl) AcceptFriend(currentUID string, req *dto.AcceptFriendReq) error {
	friend, err := s.friendRepo.FindByUIDAndFriendUID(req.FriendUID, currentUID)
	if err != nil {
		return apperrors.New(response.CodeFriendRequestNotFound, response.GetMsg(response.CodeFriendRequestNotFound))
	}

	if friend.Status != 2 {
		return apperrors.New(response.CodeFriendRequestInvalid, response.GetMsg(response.CodeFriendRequestInvalid))
	}

	friend.Status = 1
	if err := s.friendRepo.Update(friend); err != nil {
		return apperrors.New(response.CodeFriendAcceptFailed, response.GetMsg(response.CodeFriendAcceptFailed))
	}

	reverseFriend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Status:    1,
	}
	if err := s.friendRepo.Create(reverseFriend); err != nil {
		return apperrors.New(response.CodeFriendCreateFailed, response.GetMsg(response.CodeFriendCreateFailed))
	}

	return nil
}

func (s *FriendServiceImpl) DeleteFriend(currentUID string, req *dto.DeleteFriendReq) error {
	if err := s.friendRepo.Delete(currentUID, req.FriendUID); err != nil {
		return apperrors.New(response.CodeFriendDeleteFailed, response.GetMsg(response.CodeFriendDeleteFailed))
	}
	if err := s.friendRepo.Delete(req.FriendUID, currentUID); err != nil {
		return apperrors.New(response.CodeFriendDeleteFailed, response.GetMsg(response.CodeFriendDeleteFailed))
	}
	return nil
}

func (s *FriendServiceImpl) GetFriendList(currentUID string) ([]*dto.FriendResp, error) {
	friends, err := s.friendRepo.FindFriendsWithUserInfo(currentUID)
	if err != nil {
		return nil, apperrors.New(response.CodeFriendListGetFailed, response.GetMsg(response.CodeFriendListGetFailed))
	}

	var resp []*dto.FriendResp
	for _, f := range friends {
		friendUID := getString(f, "friend_uid")
		online := false
		if redis.Cache != nil {
			online, _ = redis.Cache.IsUserOnline(friendUID)
		}
		resp = append(resp, &dto.FriendResp{
			UID:       getString(f, "uid"),
			FriendUID: friendUID,
			Username:  getString(f, "username"),
			Nickname:  getString(f, "nickname"),
			Avatar:    getString(f, "avatar"),
			Remark:    getString(f, "remark"),
			Status:    getInt(f, "status"),
			IsOnline:  online,
			PubKey:    getString(f, "pub_key"),
		})
	}

	return resp, nil
}

func (s *FriendServiceImpl) GetFriendRequests(currentUID string) ([]*dto.FriendResp, error) {
	requests, err := s.friendRepo.FindFriendRequests(currentUID)
	if err != nil {
		return nil, apperrors.New(response.CodeFriendRequestsGetFailed, response.GetMsg(response.CodeFriendRequestsGetFailed))
	}

	var resp []*dto.FriendResp
	for _, f := range requests {
		resp = append(resp, &dto.FriendResp{
			UID:       getString(f, "uid"),
			FriendUID: getString(f, "friend_uid"),
			Username:  getString(f, "username"),
			Nickname:  getString(f, "nickname"),
			Avatar:    getString(f, "avatar"),
			Remark:    getString(f, "remark"),
			Status:    getInt(f, "status"),
		})
	}

	return resp, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func (s *FriendServiceImpl) SearchUser(currentUID string, keyword string) ([]*dto.UserResp, error) {
	users, err := s.userRepo.SearchUsers(keyword, currentUID, 20)
	if err != nil {
		return nil, apperrors.New(response.CodeUserSearchFailed, response.GetMsg(response.CodeUserSearchFailed))
	}

	var resp []*dto.UserResp
	for _, u := range users {
		resp = append(resp, &dto.UserResp{
			UID:      u.UID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
		})
	}

	return resp, nil
}
