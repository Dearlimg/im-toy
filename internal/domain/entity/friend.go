package entity

import (
	"gorm.io/gorm"
)

type Friend struct {
	gorm.Model
	UID       string `gorm:"size:32;not null;index" json:"uid"`
	FriendUID string `gorm:"size:32;not null;index" json:"friend_uid"`
	Remark    string `gorm:"size:32;default:''" json:"remark"`
	Status    int    `gorm:"default:1;comment:'1正常 2待验证 3已拒绝'" json:"status"`
}

func (Friend) TableName() string {
	return "friends"
}
