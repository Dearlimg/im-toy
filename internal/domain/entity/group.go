package entity

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	GroupID   string `gorm:"index;unique;size:64" json:"group_id"`
	Name      string `gorm:"size:100" json:"name"`
	OwnerUID  string `gorm:"size:64;index" json:"owner_uid"`
	Avatar    string `gorm:"size:500" json:"avatar"`
	MaxMember int    `json:"max_member"`
	Status    int    `gorm:"default:1" json:"status"`
}

type GroupMember struct {
	gorm.Model
	GroupID  string `gorm:"size:64;index;uniqueIndex:idx_group_user" json:"group_id"`
	UID      string `gorm:"size:64;index;uniqueIndex:idx_group_user" json:"uid"`
	Nickname string `gorm:"size:100" json:"nickname"`
	Role     int    `gorm:"default:2" json:"role"`
	Status   int    `gorm:"default:1" json:"status"`
}

type GroupMessage struct {
	gorm.Model
	MsgID    string `gorm:"index;unique;size:64" json:"msg_id"`
	GroupID  string `gorm:"size:64;index:idx_group_id_send_time" json:"group_id"`
	FromUID  string `gorm:"size:64;index" json:"from_uid"`
	Content  string `gorm:"type:text" json:"content"`
	Type     int    `json:"type"`
	SendTime int64  `gorm:"index:idx_group_id_send_time,sort:desc" json:"send_time"`
}

func (GroupMessage) TableName() string {
	return "group_messages"
}
