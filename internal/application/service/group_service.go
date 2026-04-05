package service

import (
	"im-chat/internal/application/dto"
)

type GroupService interface {
	CreateGroup(currentUID string, req *dto.CreateGroupReq) (*dto.GroupResp, error)
	GetGroupList(currentUID string) ([]*dto.GroupResp, error)
	GetGroupMembers(currentUID, groupID string) ([]*dto.GroupMemberResp, error)
	GetGroupMessages(currentUID, groupID string, limit int) ([]*dto.GroupMessageResp, error)
	GetGroupUnreadMessages(currentUID, groupID, lastReadMsgID string, limit int) ([]*dto.GroupMessageResp, error)
}
