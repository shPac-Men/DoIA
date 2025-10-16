package repository

import (
	"AwsProj/internal/app/ds"
	"errors"

	"gorm.io/gorm"
)

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(user *ds.Users) error {
	return r.db.Create(user).Error
}

// CheckUserExists проверяет существование пользователя по логину
func (r *Repository) CheckUserExists(login string) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Users{}).
		Where("login = ?", login).
		Count(&count).Error
	return count > 0, err
}

// GetUserByLogin возвращает пользователя по логину
func (r *Repository) GetUserByLogin(login string) (*ds.Users, error) {
	var user ds.Users
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
