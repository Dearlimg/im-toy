package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FriendRepositoryImpl struct{}

func NewFriendRepository() repository.FriendRepository {
	return &FriendRepositoryImpl{}
}

func (r *FriendRepositoryImpl) Create(friend *entity.Friend) error {
	// Restore a deleted relationship without violating the pair's unique index.
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uid"}, {Name: "friend_uid"}},
		DoUpdates: clause.AssignmentColumns([]string{"deleted_at", "status", "remark", "updated_at"}),
	}).Create(friend).Error
}

func (r *FriendRepositoryImpl) AcceptRequest(uid, friendUID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var request entity.Friend
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("uid = ? AND friend_uid = ?", friendUID, uid).First(&request).Error; err != nil {
			return err
		}
		if err := tx.Model(&request).Update("status", 1).Error; err != nil {
			return err
		}
		reverse := &entity.Friend{UID: uid, FriendUID: friendUID, Status: 1}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "friend_uid"}},
			DoUpdates: clause.AssignmentColumns([]string{"deleted_at", "status", "updated_at"}),
		}).Create(reverse).Error
	})
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
		Select("friends.*, users.username, users.nickname, users.avatar, users.pub_key").
		Joins("LEFT JOIN users ON users.uid = friends.friend_uid").
		Where("friends.uid = ? AND friends.status = ? AND friends.deleted_at IS NULL", uid, 1).
		Find(&results).Error
	return results, err
}

func (r *FriendRepositoryImpl) FindFriendRequests(uid string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := database.DB.Table("friends").
		Select("friends.*, users.username, users.nickname, users.avatar").
		Joins("LEFT JOIN users ON users.uid = friends.uid").
		Where("friends.friend_uid = ? AND friends.status = ? AND friends.deleted_at IS NULL", uid, 2).
		Find(&results).Error
	return results, err
}

func (r *FriendRepositoryImpl) Update(friend *entity.Friend) error {
	return database.DB.Save(friend).Error
}

func (r *FriendRepositoryImpl) Delete(uid, friendUID string) error {
	return database.DB.Where("(uid = ? AND friend_uid = ?) OR (uid = ? AND friend_uid = ?)", uid, friendUID, friendUID, uid).Delete(&entity.Friend{}).Error
}
