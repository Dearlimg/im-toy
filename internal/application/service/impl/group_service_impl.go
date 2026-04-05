package impl

import (
	"errors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/idgen"
	infra_repo "im-chat/internal/infra/repository"
)

type GroupServiceImpl struct {
	groupRepo repository.GroupRepository
}

func NewGroupServiceImpl(groupRepo repository.GroupRepository) service.GroupService {
	return &GroupServiceImpl{
		groupRepo: groupRepo,
	}
}

func (s *GroupServiceImpl) CreateGroup(currentUID string, req *dto.CreateGroupReq) (*dto.GroupResp, error) {
	groupID := idgen.Gen.NextUID()

	group := &entity.Group{
		GroupID:   groupID,
		Name:      req.Name,
		OwnerUID:  currentUID,
		MaxMember: 500,
		Status:    1,
	}

	if err := s.groupRepo.CreateGroup(group); err != nil {
		return nil, errors.New("创建群聊失败")
	}

	owner := &entity.GroupMember{
		GroupID: groupID,
		UID:     currentUID,
		Role:    1,
		Status:  1,
	}
	if err := s.groupRepo.AddMember(owner); err != nil {
		return nil, errors.New("添加群主失败")
	}

	for _, uid := range req.MemberUIDs {
		if uid == "" || uid == currentUID {
			continue
		}
		member := &entity.GroupMember{
			GroupID: groupID,
			UID:     uid,
			Role:    2,
			Status:  1,
		}
		if err := s.groupRepo.AddMember(member); err != nil {
			continue
		}
	}

	return &dto.GroupResp{
		GroupID:   groupID,
		Name:      group.Name,
		OwnerUID:  group.OwnerUID,
		Avatar:    group.Avatar,
		MaxMember: group.MaxMember,
	}, nil
}

func (s *GroupServiceImpl) GetGroupList(currentUID string) ([]*dto.GroupResp, error) {
	groups, err := s.groupRepo.FindGroupsByUID(currentUID)
	if err != nil {
		return nil, errors.New("获取群聊列表失败")
	}

	var resp []*dto.GroupResp
	for _, g := range groups {
		resp = append(resp, &dto.GroupResp{
			GroupID:   g.GroupID,
			Name:      g.Name,
			OwnerUID:  g.OwnerUID,
			Avatar:    g.Avatar,
			MaxMember: g.MaxMember,
		})
	}

	return resp, nil
}

func (s *GroupServiceImpl) GetGroupMembers(currentUID, groupID string) ([]*dto.GroupMemberResp, error) {
	if _, err := s.groupRepo.FindMember(groupID, currentUID); err != nil {
		return nil, errors.New("无权查看该群或您不是群成员")
	}

	var members []*entity.GroupMember
	if err := database.DB.Where("group_id = ? AND status = 1", groupID).Find(&members).Error; err != nil {
		return nil, errors.New("获取群成员失败")
	}

	var uids []string
	for _, m := range members {
		uids = append(uids, m.UID)
	}

	var users []*entity.User
	if err := database.DB.Where("uid IN ?", uids).Find(&users).Error; err != nil {
		return nil, errors.New("获取用户信息失败")
	}

	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.UID] = u
	}

	var resp []*dto.GroupMemberResp
	for _, m := range members {
		if user, ok := userMap[m.UID]; ok {
			resp = append(resp, &dto.GroupMemberResp{
				UID:      user.UID,
				Username: user.Username,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
				Role:     m.Role,
			})
		}
	}

	return resp, nil
}

func (s *GroupServiceImpl) GetGroupMessages(currentUID, groupID string, limit int) ([]*dto.GroupMessageResp, error) {
	if _, err := s.groupRepo.FindMember(groupID, currentUID); err != nil {
		return nil, errors.New("无权查看该群消息或您不是群成员")
	}

	messages, err := s.groupRepo.FindMessagesByGroupID(groupID, limit)
	if err != nil {
		return nil, errors.New("获取群消息失败")
	}

	members, err := s.groupRepo.FindMembersByGroupID(groupID)
	if err != nil {
		return nil, errors.New("获取群成员失败")
	}
	totalCount := len(members)

	var uids []string
	for _, m := range messages {
		uids = append(uids, m.FromUID)
	}

	var users []*entity.User
	if err := database.DB.Where("uid IN ?", uids).Find(&users).Error; err != nil {
		return nil, errors.New("获取用户信息失败")
	}

	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.UID] = u
	}

	readRepo := infra_repo.NewMessageReadRepository()

	var resp []*dto.GroupMessageResp
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		if user, ok := userMap[m.FromUID]; ok {
			readCount, _ := readRepo.GetReadCount(m.MsgID)
			resp = append(resp, &dto.GroupMessageResp{
				MsgID:        m.MsgID,
				GroupID:      m.GroupID,
				FromUID:      m.FromUID,
				FromUsername: user.Username,
				FromNickname: user.Nickname,
				Content:      m.Content,
				SendTime:     m.SendTime,
				ReadCount:    readCount,
				TotalCount:   totalCount,
			})
		}
	}

	return resp, nil
}
