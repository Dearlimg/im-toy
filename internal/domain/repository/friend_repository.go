package repository

import "im-chat/internal/domain/entity"

type FriendRepository interface {
	AcceptRequest(uid, friendUID string) error
	Create(friend *entity.Friend) error
	FindByUID(uid string) ([]*entity.Friend, error)
	FindByUIDAndFriendUID(uid, friendUID string) (*entity.Friend, error)
	FindFriendsWithUserInfo(uid string) ([]map[string]interface{}, error)
	FindFriendRequests(uid string) ([]map[string]interface{}, error)
	Update(friend *entity.Friend) error
	Delete(uid, friendUID string) error
}
