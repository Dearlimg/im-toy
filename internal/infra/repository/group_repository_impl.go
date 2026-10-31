package repository

import (
	"gorm.io/gorm"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
)

type GroupRepositoryImpl struct{}

func (r *GroupRepositoryImpl) CreateWithMembers(group *entity.Group, members []*entity.GroupMember) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		return tx.Create(&members).Error
	})
}

func NewGroupRepository() repository.GroupRepository {
	return &GroupRepositoryImpl{}
}

func (r *GroupRepositoryImpl) CreateGroup(group *entity.Group) error {
	return database.DB.Create(group).Error
}

func (r *GroupRepositoryImpl) FindGroupByID(groupID string) (*entity.Group, error) {
	var group entity.Group
	err := database.DB.Where("group_id = ?", groupID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepositoryImpl) FindGroupsByUID(uid string) ([]*entity.Group, error) {
	var groups []*entity.Group
	err := database.DB.Table("groups").
		Joins("JOIN group_members ON groups.group_id = group_members.group_id").
		Where("group_members.uid = ? AND group_members.status = 1", uid).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepositoryImpl) UpdateGroup(group *entity.Group) error {
	return database.DB.Save(group).Error
}

func (r *GroupRepositoryImpl) DeleteGroup(groupID string) error {
	return database.DB.Where("group_id = ?", groupID).Delete(&entity.Group{}).Error
}

func (r *GroupRepositoryImpl) AddMember(member *entity.GroupMember) error {
	return database.DB.Create(member).Error
}

func (r *GroupRepositoryImpl) RemoveMember(groupID, uid string) error {
	return database.DB.Where("group_id = ? AND uid = ?", groupID, uid).Delete(&entity.GroupMember{}).Error
}

func (r *GroupRepositoryImpl) FindMembersByGroupID(groupID string) ([]*entity.GroupMember, error) {
	var members []*entity.GroupMember
	err := database.DB.Where("group_id = ? AND status = 1", groupID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *GroupRepositoryImpl) FindMember(groupID, uid string) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := database.DB.Where("group_id = ? AND uid = ? AND status = ?", groupID, uid, 1).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *GroupRepositoryImpl) CreateMessage(msg *entity.GroupMessage) error {
	return database.DB.Create(msg).Error
}

func (r *GroupRepositoryImpl) FindMessagesByGroupID(groupID string, limit int) ([]*entity.GroupMessage, error) {
	var messages []*entity.GroupMessage
	err := database.DB.Where("group_id = ?", groupID).
		Order("seq DESC, id DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *GroupRepositoryImpl) FindUnreadMessages(groupID, lastReadMsgID string, limit int) ([]*entity.GroupMessage, error) {
	var messages []*entity.GroupMessage
	query := database.DB.Where("group_id = ?", groupID)

	if lastReadMsgID != "" {
		var last entity.GroupMessage
		if err := database.DB.Where("group_id = ? AND msg_id = ?", groupID, lastReadMsgID).First(&last).Error; err != nil {
			return nil, err
		}
		query = query.Where("seq > ? OR (seq = ? AND id > ?)", last.Seq, last.Seq, last.ID)
	}

	err := query.Order("seq ASC, id ASC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *GroupRepositoryImpl) FindMessagesByMsgIDs(msgIDs []string) ([]*entity.GroupMessage, error) {
	if len(msgIDs) == 0 {
		return []*entity.GroupMessage{}, nil
	}
	var messages []*entity.GroupMessage
	err := database.DB.Where("msg_id IN ?", msgIDs).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
