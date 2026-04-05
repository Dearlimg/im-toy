package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"

	"gorm.io/gorm/clause"
)

type MessageReadRepositoryImpl struct{}

func NewMessageReadRepository() *MessageReadRepositoryImpl {
	return &MessageReadRepositoryImpl{}
}

func (r *MessageReadRepositoryImpl) BatchCreate(messageReads []*entity.MessageRead) error {
	if len(messageReads) == 0 {
		return nil
	}
	return database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(messageReads).Error
}

func (r *MessageReadRepositoryImpl) GetReadCount(msgID string) (int64, error) {
	var count int64
	err := database.DB.Model(&entity.MessageRead{}).Where("msg_id = ?", msgID).Count(&count).Error
	return count, err
}

func (r *MessageReadRepositoryImpl) IsRead(msgID, uid string) (bool, error) {
	var count int64
	err := database.DB.Model(&entity.MessageRead{}).Where("msg_id = ? AND uid = ?", msgID, uid).Count(&count).Error
	return count > 0, err
}
