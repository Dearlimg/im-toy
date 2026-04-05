package impl

import (
	"errors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
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
		return errors.New("不能添加自己为好友")
	}

	_, err := s.userRepo.FindByUID(req.FriendUID)
	if err != nil {
		return errors.New("用户不存在")
	}

	existing, _ := s.friendRepo.FindByUIDAndFriendUID(currentUID, req.FriendUID)
	if existing != nil {
		if existing.Status == 1 {
			return errors.New("已经是好友了")
		} else if existing.Status == 2 {
			return errors.New("已发送好友请求，等待对方验证")
		}
	}

	friend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Remark:    req.Remark,
		Status:    2,
	}

	if err := s.friendRepo.Create(friend); err != nil {
		return errors.New("添加好友失败")
	}

	return nil
}

func (s *FriendServiceImpl) AcceptFriend(currentUID string, req *dto.AcceptFriendReq) error {
	friend, err := s.friendRepo.FindByUIDAndFriendUID(req.FriendUID, currentUID)
	if err != nil {
		return errors.New("好友请求不存在")
	}

	if friend.Status != 2 {
		return errors.New("好友请求状态异常")
	}

	friend.Status = 1
	if err := s.friendRepo.Update(friend); err != nil {
		return errors.New("接受好友请求失败")
	}

	reverseFriend := &entity.Friend{
		UID:       currentUID,
		FriendUID: req.FriendUID,
		Status:    1,
	}
	if err := s.friendRepo.Create(reverseFriend); err != nil {
		return errors.New("创建好友关系失败")
	}

	return nil
}

func (s *FriendServiceImpl) DeleteFriend(currentUID string, req *dto.DeleteFriendReq) error {
	if err := s.friendRepo.Delete(currentUID, req.FriendUID); err != nil {
		return errors.New("删除好友失败")
	}
	if err := s.friendRepo.Delete(req.FriendUID, currentUID); err != nil {
		return errors.New("删除好友失败")
	}
	return nil
}

func (s *FriendServiceImpl) GetFriendList(currentUID string) ([]*dto.FriendResp, error) {
	friends, err := s.friendRepo.FindFriendsWithUserInfo(currentUID)
	if err != nil {
		return nil, errors.New("获取好友列表失败")
	}

	var resp []*dto.FriendResp
	for _, f := range friends {
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

func (s *FriendServiceImpl) GetFriendRequests(currentUID string) ([]*dto.FriendResp, error) {
	requests, err := s.friendRepo.FindFriendRequests(currentUID)
	if err != nil {
		return nil, errors.New("获取好友请求失败")
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
		return nil, errors.New("搜索用户失败")
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
