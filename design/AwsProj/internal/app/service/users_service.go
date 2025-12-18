package service

import (
	"AwsProj/internal/app/config"
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserService struct {
	repo   *repository.Repository
	config *config.Config
}

// NewUserService - конструктор с двумя параметрами
func NewUserService(repo *repository.Repository, cfg *config.Config) *UserService {
	return &UserService{
		repo:   repo,
		config: cfg,
	}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	if req.Login == "" {
		return nil, errors.New("логин обязателен")
	}
	if req.Password == "" {
		return nil, errors.New("пароль обязателен")
	}

	_, err := s.repo.GetUserByLogin(req.Login)
	if err == nil {
		return nil, errors.New("пользователь с таким логином уже существует")
	}

	user := &ds.Users{
		Login:       req.Login,
		Password:    req.Password,
		IsModerator: false,
	}

	createdUser, err := s.repo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return &RegisterResponse{
		ID:    createdUser.ID,
		Login: createdUser.Login,
	}, nil
}

// Login выполняет аутентификацию пользователя
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	if req.Login == "" {
		return nil, errors.New("логин обязателен")
	}
	if req.Password == "" {
		return nil, errors.New("пароль обязателен")
	}

	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return nil, errors.New("неверный логин или пароль")
	}

	// Сравниваем пароли напрямую (без bcrypt)
	if user.Password != req.Password {
		return nil, errors.New("неверный логин или пароль")
	}

	role := "client"
	if user.IsModerator {
		role = "admin"
	}

	token, err := s.generateJWT(user.ID, user.Login, role, user.IsModerator)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return &LoginResponse{
		Token:       token,
		ID:          user.ID,
		Login:       user.Login,
		Role:        role,
		IsModerator: user.IsModerator,
		Message:     "Успешная аутентификация",
	}, nil
}

// generateJWT генерирует JWT токен
func (s *UserService) generateJWT(userID uint, login string, role string, isModerator bool) (string, error) {
	expiresIn := s.config.JWT.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 24 * time.Hour
	}

	claims := &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "AwsProj",
			Subject:   fmt.Sprintf("%d", userID),
		},
		UserID:      userID,
		UserUUID:    uuid.New(),
		Login:       login,
		Role:        role,
		IsModerator: isModerator,
		Scopes:      []string{"user"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Token))
}

// GetUserProfile возвращает профиль пользователя
func (s *UserService) GetUserProfile(userID uint) (*UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	role := "client"
	if user.IsModerator {
		role = "admin"
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Login:       user.Login,
		Role:        role,
		IsModerator: user.IsModerator,
	}, nil
}

// GetUserByID возвращает данные пользователя по ID
func (s *UserService) GetUserByID(userID uint) (*UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	role := "client"
	if user.IsModerator {
		role = "admin"
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Login:       user.Login,
		Role:        role,
		IsModerator: user.IsModerator,
	}, nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *UserService) GetUserByLogin(login string) (*ds.Users, error) {
	return s.repo.GetUserByLogin(login)
}

// GetAllUsers возвращает всех пользователей
func (s *UserService) GetAllUsers() ([]*UserProfileResponse, error) {
	users, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	result := make([]*UserProfileResponse, len(users))
	for i, user := range users {
		role := "client"
		if user.IsModerator {
			role = "admin"
		}

		result[i] = &UserProfileResponse{
			ID:          user.ID,
			Login:       user.Login,
			Role:        role,
			IsModerator: user.IsModerator,
		}
	}

	return result, nil
}

// UpdateUser обновляет данные пользователя
// service/user_service.go

func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) error {
	// Валидация логина (без изменений)
	if req.Login != "" {
		if len(req.Login) < 3 || len(req.Login) > 25 {
			return fmt.Errorf("логин должен быть от 3 до 25 символов")
		}
		exists, err := s.repo.CheckLoginExists(req.Login, userID)
		if err != nil {
			return fmt.Errorf("ошибка проверки логина: %v", err)
		}
		if exists {
			return fmt.Errorf("пользователь с логином '%s' уже существует", req.Login)
		}
	}

	/*
	   // ---- ПРОВЕРКА ДЛИНЫ ПАРОЛЯ УДАЛЕНА ----
	   if req.Password != "" {
	       if len(req.Password) < 6 {
	           return fmt.Errorf("пароль должен быть не менее 6 символов")
	       }
	   }
	*/

	// Подготавливаем обновления
	updates := make(map[string]interface{})

	if req.Login != "" {
		updates["login"] = req.Login
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}

	if len(updates) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	// Выполняем обновление в БД
	err := s.repo.UpdateUser(userID, updates)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserRole обновляет роль пользователя
func (s *UserService) UpdateUserRole(userID uint, isModerator bool) error {
	return s.repo.UpdateUserRole(userID, isModerator)
}

// Logout выполняет деавторизацию пользователя
func (s *UserService) Logout(userID uint) (*LogoutResponse, error) {
	// В JWT-based системе logout происходит на клиенте
	// Здесь можно добавить blacklist токенов в Redis
	return &LogoutResponse{
		Message: "Успешный выход из системы",
	}, nil
}
