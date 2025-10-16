package service

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository"
	"errors"
	"fmt"
)

type UserService struct {
	repo *repository.Repository
}

func NewUserService(repo *repository.Repository) *UserService {
	return &UserService{repo: repo}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// Валидация логина
	if len(req.Login) < 3 || len(req.Login) > 25 {
		return nil, errors.New("логин должен быть от 3 до 25 символов")
	}

	// Валидация пароля
	if len(req.Password) < 6 {
		return nil, errors.New("пароль должен быть не менее 6 символов")
	}

	// Проверяем уникальность логина
	exists, err := s.repo.CheckUserExists(req.Login)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки пользователя: %v", err)
	}
	if exists {
		return nil, errors.New("пользователь с таким логином уже существует")
	}

	// Создаем пользователя (пароль хранится как есть)
	user := &ds.Users{
		Login:       req.Login,
		Password:    req.Password, // без хеширования
		IsModerator: req.IsModerator,
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %v", err)
	}

	return &RegisterResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
		Message:     "Пользователь успешно зарегистрирован",
	}, nil
}
