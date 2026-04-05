package entity

import "gorm.io/gorm"

type MessageRead struct {
	gorm.Model
	MsgID    string `gorm:"size:64;not null;index:idx_msg_uid,unique;index:idx_msg" json:"msg_id"`
	UID      string `gorm:"size:32;not null;index:idx_msg_uid,unique;index:idx_uid" json:"uid"`
	ReadTime int64  `gorm:"index" json:"read_time"`
}
