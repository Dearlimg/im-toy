package entity

import (
	"gorm.io/gorm"
)

type PrivateMessage struct {
	gorm.Model
	MsgID    string `gorm:"size:40;uniqueIndex;not null" json:"msg_id"`
	FromUID  string `gorm:"size:32;index;not null" json:"from_uid"`
	ToUID    string `gorm:"size:32;index;not null" json:"to_uid"`
	Content  string `gorm:"size:2000;not null" json:"content"`
	Type     int    `gorm:"default:1" json:"type"`
	SendTime int64  `gorm:"index" json:"send_time"`
	Read     int    `gorm:"default:0" json:"read"`
}

func (PrivateMessage) TableName() string {
	return "private_messages"
}
