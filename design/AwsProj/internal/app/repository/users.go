package repository

import (
	"AwsProj/internal/app/ds"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CreateUser создает нового пользователя
// repository/user.go
func (r *Repository) CreateUser(user *ds.Users) (*ds.Users, error) {
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
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
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(userID uint) (*ds.Users, error) {
	var user ds.Users
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

// GetAllUsers возвращает всех пользователей
func (r *Repository) GetAllUsers() ([]*ds.Users, error) {
	var users []*ds.Users
	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(userID uint, updates map[string]interface{}) error {
	// Сначала проверяем существование пользователя
	var user ds.Users
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("пользователь не найден")
		}
		return err
	}

	// Обновляем пользователя
	err = r.db.Model(&ds.Users{}).
		Where("id = ?", userID).
		Updates(updates).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserRole обновляет роль пользователя (is_moderator)
func (r *Repository) UpdateUserRole(userID uint, isModerator bool) error {
	return r.db.Model(&ds.Users{}).
		Where("id = ?", userID).
		Update("is_moderator", isModerator).Error
}

// CheckLoginExists проверяет существование логина у других пользователей
func (r *Repository) CheckLoginExists(login string, excludeUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Users{}).
		Where("login = ? AND id != ?", login, excludeUserID).
		Count(&count).Error
	return count > 0, err
}
