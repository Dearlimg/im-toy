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
	var dbCount int64
	err := database.DB.Model(&entity.MessageRead{}).Where("msg_id = ?", msgID).Count(&dbCount).Error
	if err != nil {
		return 0, err
	}
	return dbCount, nil
}

func (r *MessageReadRepositoryImpl) BatchGetReadCount(msgIDs []string) (map[string]int64, error) {
	if len(msgIDs) == 0 {
		return make(map[string]int64), nil
	}

	counts := make(map[string]int64, len(msgIDs))

	type Result struct {
		MsgID string
		Count int64
	}

	var results []Result
	err := database.DB.Model(&entity.MessageRead{}).
		Select("msg_id, COUNT(*) as count").
		Where("msg_id IN ?", msgIDs).
		Group("msg_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	for _, res := range results {
		counts[res.MsgID] = res.Count
	}

	for _, msgID := range msgIDs {
		if _, ok := counts[msgID]; !ok {
			counts[msgID] = 0
		}
	}

	return counts, nil
}

func (r *MessageReadRepositoryImpl) IsRead(msgID, uid string) (bool, error) {
	var count int64
	err := database.DB.Model(&entity.MessageRead{}).Where("msg_id = ? AND uid = ?", msgID, uid).Count(&count).Error
	return count > 0, err
}
