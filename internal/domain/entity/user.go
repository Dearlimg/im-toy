package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string    `gorm:"size:32;not null;unique" json:"username"`
	Password  string    `gorm:"size:64;not null" json:"password"`
	Nickname  string    `gorm:"size:32;not null" json:"nickname"`
	Avatar    string    `gorm:"size:255;default:''" json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
