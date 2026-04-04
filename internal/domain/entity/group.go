package entity

import (
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Name    string `gorm:"size:32;not null" json:"name"`
	OwnerID int64  `gorm:"not null" json:"owner_id"`
}

type GroupMember struct {
	gorm.Model
	GroupID int64 `gorm:"not null" json:"group_id"`
	UserID  int64 `gorm:"not null" json:"user_id"`
}
