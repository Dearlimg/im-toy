package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UID       string    `gorm:"size:32;uniqueIndex;not null" json:"uid"`
	Username  string    `gorm:"size:32;not null;uniqueIndex" json:"username"`
	Password  string    `gorm:"size:64;not null" json:"-"`
	Nickname  string    `gorm:"size:32;not null" json:"nickname"`
	Avatar    string    `gorm:"size:255;default:''" json:"avatar"`
	PubKey    string    `gorm:"size:256;default:''" json:"pubkey"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
