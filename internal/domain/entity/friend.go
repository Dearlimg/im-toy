package entity

import (
	"gorm.io/gorm"
)

type Friend struct {
	gorm.Model
	UserID   int64  `gorm:"not null" json:"user_id"`
	FriendID int64  `gorm:"not null" json:"friend_id"`
	Remark   string `gorm:"size:32;default:''" json:"remark"`
	Status   int    `gorm:"default:1;comment:'1正常 2拉黑'" json:"status"`
}
