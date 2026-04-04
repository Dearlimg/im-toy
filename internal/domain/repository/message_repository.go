package repository

import "im-chat/internal/domain/entity"

type MessageRepository interface {
	Create(message *entity.Message) error
	FindBySendID(sendID int64) ([]*entity.Message, error)
	FindByRecvID(recvID int64) ([]*entity.Message, error)
	FindByUserID(userID int64) ([]*entity.Message, error)
}
