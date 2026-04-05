package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
)

type FriendRepositoryImpl struct{}

func NewFriendRepository() repository.FriendRepository {
	return &FriendRepositoryImpl{}
}

func (r *FriendRepositoryImpl) Create(friend *entity.Friend) error {
	return database.DB.Create(friend).Error
}

func (r *FriendRepositoryImpl) FindByUID(uid string) ([]*entity.Friend, error) {
	var friends []*entity.Friend
	err := database.DB.Where("uid = ? AND status = ?", uid, 1).Find(&friends).Error
	return friends, err
}

func (r *FriendRepositoryImpl) FindByUIDAndFriendUID(uid, friendUID string) (*entity.Friend, error) {
	var friend entity.Friend
	err := database.DB.Where("uid = ? AND friend_uid = ?", uid, friendUID).First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *FriendRepositoryImpl) FindFriendsWithUserInfo(uid string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := database.DB.Table("friends").
		Select("friends.*, users.username, users.nickname, users.avatar").
		Joins("LEFT JOIN users ON users.uid = friends.friend_uid").
		Where("friends.uid = ? AND friends.status = ?", uid, 1).
		Find(&results).Error
	return results, err
}

func (r *FriendRepositoryImpl) FindFriendRequests(uid string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := database.DB.Table("friends").
		Select("friends.*, users.username, users.nickname, users.avatar").
		Joins("LEFT JOIN users ON users.uid = friends.uid").
		Where("friends.friend_uid = ? AND friends.status = ?", uid, 2).
		Find(&results).Error
	return results, err
}

func (r *FriendRepositoryImpl) Update(friend *entity.Friend) error {
	return database.DB.Save(friend).Error
}

func (r *FriendRepositoryImpl) Delete(uid, friendUID string) error {
	return database.DB.Where("(uid = ? AND friend_uid = ?) OR (uid = ? AND friend_uid = ?)", uid, friendUID, friendUID, uid).Delete(&entity.Friend{}).Error
}
