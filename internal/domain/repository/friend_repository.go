package repository

import "im-chat/internal/domain/entity"

type FriendRepository interface {
	Create(friend *entity.Friend) error
	FindByUserID(userID int64) ([]*entity.Friend, error)
	FindByUserAndFriendID(userID, friendID int64) (*entity.Friend, error)
	Update(friend *entity.Friend) error
	Delete(userID, friendID int64) error
}
