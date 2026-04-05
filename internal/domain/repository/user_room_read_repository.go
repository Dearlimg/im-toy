package repository

import "im-chat/internal/domain/entity"

type UserRoomReadRepository interface {
	Upsert(uid, roomID, lastReadMsgID string, lastReadTime int64) error
	FindByUIDAndRoomID(uid, roomID string) (*entity.UserRoomRead, error)
	GetUnreadCount(uid, roomID string, lastMsgID string) (int64, error)
}
