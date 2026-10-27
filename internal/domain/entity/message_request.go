package entity

// MessageRequest persists the idempotency decision together with the message.
type MessageRequest struct {
	SenderUID   string `gorm:"primaryKey;size:32"`
	ClientMsgID string `gorm:"primaryKey;size:64"`
	Fingerprint string `gorm:"size:64;not null"`
	MsgID       string `gorm:"size:40;not null"`
	Seq         uint64 `gorm:"not null"`
	SendTime    int64  `gorm:"not null"`
}

// MessageConversation serializes sequence allocation for one conversation.
type MessageConversation struct {
	ConversationKey string `gorm:"primaryKey;size:64"`
	LastSeq         uint64 `gorm:"not null"`
}
