package repository

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/database"
)

type UserRepositoryImpl struct{}

func NewUserRepository() repository.UserRepository {
	return &UserRepositoryImpl{}
}

func (r *UserRepositoryImpl) Create(user *entity.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepositoryImpl) FindByID(id int64) (*entity.User, error) {
	var user entity.User
	err := database.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByUID(uid string) (*entity.User, error) {
	var user entity.User
	err := database.DB.Where("uid = ?", uid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) SearchUsers(keyword string, excludeUID string, limit int) ([]*entity.User, error) {
	var users []*entity.User
	err := database.DB.Where("(username LIKE ? OR nickname LIKE ?) AND uid != ?",
		"%"+keyword+"%", "%"+keyword+"%", excludeUID).
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepositoryImpl) FindByUIDs(uids []string) ([]*entity.User, error) {
	var users []*entity.User
	if len(uids) == 0 {
		return users, nil
	}
	err := database.DB.Where("uid IN ?", uids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepositoryImpl) Update(user *entity.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepositoryImpl) Delete(id int64) error {
	return database.DB.Delete(&entity.User{}, id).Error
}
