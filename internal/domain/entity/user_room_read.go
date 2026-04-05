package entity

import (
	"gorm.io/gorm"
)

type UserRoomRead struct {
	gorm.Model
	UID           string `gorm:"size:32;not null;index:idx_uid_room,unique" json:"uid"`
	RoomID        string `gorm:"size:32;not null;index:idx_uid_room,unique" json:"room_id"`
	LastReadMsgID string `gorm:"size:64" json:"last_read_msg_id"`
	LastReadTime  int64  `gorm:"index" json:"last_read_time"`
}

func (UserRoomRead) TableName() string {
	return "user_room_read"
}
