package repository

import "im-chat/internal/domain/entity"

type PrivateMessageRepository interface {
	Create(message *entity.PrivateMessage) error
	FindConversation(uid1, uid2 string, limit, offset int) ([]*entity.PrivateMessage, error)
	MarkAsRead(fromUID, toUID string) error
	GetUnreadCount(uid string) (int64, error)
}
