package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
)

type UserRoomReadRepositoryImpl struct{}

func NewUserRoomReadRepository() repository.UserRoomReadRepository {
	return &UserRoomReadRepositoryImpl{}
}

func (r *UserRoomReadRepositoryImpl) Upsert(uid, roomID, lastReadMsgID string, lastReadTime int64) error {
	var existing entity.UserRoomRead
	err := database.DB.Where("uid = ? AND room_id = ?", uid, roomID).First(&existing).Error

	if err != nil {
		record := &entity.UserRoomRead{
			UID:           uid,
			RoomID:        roomID,
			LastReadMsgID: lastReadMsgID,
			LastReadTime:  lastReadTime,
		}
		return database.DB.Create(record).Error
	}

	return database.DB.Model(&existing).Updates(map[string]interface{}{
		"last_read_msg_id": lastReadMsgID,
		"last_read_time":   lastReadTime,
	}).Error
}

func (r *UserRoomReadRepositoryImpl) FindByUIDAndRoomID(uid, roomID string) (*entity.UserRoomRead, error) {
	var record entity.UserRoomRead
	err := database.DB.Where("uid = ? AND room_id = ?", uid, roomID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *UserRoomReadRepositoryImpl) GetUnreadCount(uid, roomID string, lastMsgID string) (int64, error) {
	var count int64
	record, _ := r.FindByUIDAndRoomID(uid, roomID)
	if record == nil || record.LastReadMsgID == "" {
		return 0, nil
	}

	err := database.DB.Model(&entity.PrivateMessage{}).
		Where("msg_id > ? AND ((from_uid = ? AND to_uid = ?) OR (from_uid = ? AND to_uid = ?))",
			record.LastReadMsgID, uid, roomID, roomID, uid).
		Count(&count).Error
	return count, err
}
