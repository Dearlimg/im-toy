package repository

import "im-chat/internal/domain/entity"

type MessageReadRepository interface {
	BatchCreate(messageReads []*entity.MessageRead) error
	GetReadCount(msgID string) (int64, error)
	IsRead(msgID, uid string) (bool, error)
}
