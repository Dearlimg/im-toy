package entity

import (
	"gorm.io/gorm"
)

type PrivateMessage struct {
	gorm.Model
	MsgID    string `gorm:"size:40;uniqueIndex;not null" json:"msg_id"`
	Seq      uint64 `gorm:"index" json:"seq"`
	FromUID  string `gorm:"size:32;index;not null" json:"from_uid"`
	ToUID    string `gorm:"size:32;index;not null" json:"to_uid"`
	Content  string `gorm:"size:2000;not null" json:"content"`
	Type     int    `gorm:"default:1" json:"type"`
	SendTime int64  `gorm:"index;index:idx_private_conversation,sort:desc" json:"send_time"`
	Read     int    `gorm:"default:0" json:"read"`
}

func (PrivateMessage) TableName() string {
	return "private_messages"
}
