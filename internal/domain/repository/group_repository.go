package repository

import "im-chat/internal/domain/entity"

type GroupRepository interface {
	CreateWithMembers(group *entity.Group, members []*entity.GroupMember) error
	CreateGroup(group *entity.Group) error
	FindGroupByID(groupID string) (*entity.Group, error)
	FindGroupsByUID(uid string) ([]*entity.Group, error)
	UpdateGroup(group *entity.Group) error
	DeleteGroup(groupID string) error

	AddMember(member *entity.GroupMember) error
	RemoveMember(groupID, uid string) error
	FindMembersByGroupID(groupID string) ([]*entity.GroupMember, error)
	FindMember(groupID, uid string) (*entity.GroupMember, error)

	CreateMessage(msg *entity.GroupMessage) error
	FindMessagesByGroupID(groupID string, limit int) ([]*entity.GroupMessage, error)
	FindUnreadMessages(groupID, lastReadMsgID string, limit int) ([]*entity.GroupMessage, error)
	FindMessagesByMsgIDs(msgIDs []string) ([]*entity.GroupMessage, error)
}
