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

func (s *UserService) GetUserProfile(userID uint) (*UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}, nil
}

// GetUserByID возвращает данные любого пользователя по ID
func (s *UserService) GetUserByID(userID uint) (*UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}, nil
}

func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) error {
	// Валидация
	if req.Login != "" {
		if len(req.Login) < 3 || len(req.Login) > 25 {
			return fmt.Errorf("логин должен быть от 3 до 25 символов")
		}

		// Проверяем уникальность логина
		exists, err := s.repo.CheckLoginExists(req.Login, userID)
		if err != nil {
			return fmt.Errorf("ошибка проверки логина: %v", err)
		}
		if exists {
			return fmt.Errorf("пользователь с логином '%s' уже существует", req.Login)
		}
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			return fmt.Errorf("пароль должен быть не менее 6 символов")
		}
	}

	// Подготавливаем обновления
	updates := make(map[string]interface{})

	if req.Login != "" {
		updates["login"] = req.Login
	}
	if req.Password != "" {
		updates["password"] = req.Password // без хеширования для учебного проекта
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	// Выполняем обновление
	err := s.repo.UpdateUser(userID, updates)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Валидация
	if req.Login == "" {
		return nil, errors.New("логин обязателен")
	}
	if req.Password == "" {
		return nil, errors.New("пароль обязателен")
	}

	// Ищем пользователя по логину
	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("неверный логин или пароль")
	}

	// Проверяем пароль (без хеширования для учебного проекта)
	if user.Password != req.Password {
		return nil, errors.New("неверный логин или пароль")
	}

	return &LoginResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
		Message:     "Успешная аутентификация",
	}, nil
}

// Logout выполняет деавторизацию пользователя
func (s *UserService) Logout(userID uint) (*LogoutResponse, error) {
	// В реальном приложении здесь может быть:
	// - Добавление токена в blacklist
	// - Удаление сессии из базы
	// - Очистка кэша

	// Для учебного проекта просто возвращаем успех
	return &LogoutResponse{
		Message: "Успешный выход из системы",
	}, nil
}
