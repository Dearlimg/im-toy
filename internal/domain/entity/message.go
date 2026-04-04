package entity

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SendID   int64  `gorm:"not null" json:"send_id"`
	RecvID   int64  `gorm:"not null" json:"recv_id"`
	Type     int    `gorm:"not null;comment:'1私聊 2群聊'" json:"type"`
	Content  string `gorm:"type:text;not null" json:"content"`
	SendTime int64  `gorm:"not null" json:"send_time"`
}
