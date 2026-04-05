package repository

import "im-chat/internal/domain/entity"

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id int64) (*entity.User, error)
	FindByUID(uid string) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	FindByUIDs(uids []string) ([]*entity.User, error)
	SearchUsers(keyword string, excludeUID string, limit int) ([]*entity.User, error)
	Update(user *entity.User) error
	Delete(id int64) error
}
