package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
)

type PrivateMessageRepositoryImpl struct{}

func NewPrivateMessageRepository() repository.PrivateMessageRepository {
	return &PrivateMessageRepositoryImpl{}
}

func (r *PrivateMessageRepositoryImpl) Create(message *entity.PrivateMessage) error {
	return database.DB.Create(message).Error
}

func (r *PrivateMessageRepositoryImpl) FindConversation(uid1, uid2 string, limit, offset int) ([]*entity.PrivateMessage, error) {
	var messages []*entity.PrivateMessage
	err := database.DB.Where("(from_uid = ? AND to_uid = ?) OR (from_uid = ? AND to_uid = ?)", uid1, uid2, uid2, uid1).
		Order("send_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, err
}

func (r *PrivateMessageRepositoryImpl) MarkAsRead(fromUID, toUID string) error {
	return database.DB.Model(&entity.PrivateMessage{}).
		Where("from_uid = ? AND to_uid = ? AND `read` = ?", fromUID, toUID, 0).
		Update("read", 1).Error
}

func (r *PrivateMessageRepositoryImpl) GetUnreadCount(uid string) (int64, error) {
	var count int64
	err := database.DB.Model(&entity.PrivateMessage{}).
		Where("to_uid = ? AND `read` = ?", uid, 0).
		Count(&count).Error
	return count, err
}
