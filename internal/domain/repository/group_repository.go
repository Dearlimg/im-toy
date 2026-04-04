package repository

import "im-chat/internal/domain/entity"

type GroupRepository interface {
	Create(group *entity.Group) error
	FindByID(id int64) (*entity.Group, error)
	FindByOwnerID(ownerID int64) ([]*entity.Group, error)
	Update(group *entity.Group) error
	Delete(id int64) error
}

type GroupMemberRepository interface {
	Create(member *entity.GroupMember) error
	FindByGroupID(groupID int64) ([]*entity.GroupMember, error)
	FindByUserID(userID int64) ([]*entity.GroupMember, error)
	FindByGroupAndUserID(groupID, userID int64) (*entity.GroupMember, error)
	Delete(groupID, userID int64) error
}
