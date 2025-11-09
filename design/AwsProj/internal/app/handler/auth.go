package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"AwsProj/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Хардкодные credentials (временные)
const (
	hardcodedLogin    = "login"
	hardcodedPassword = "check123"
)

// Login обработчик для аутентификации
// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(gCtx *gin.Context) {
	logrus.Info("=== LOGIN HANDLER CALLED ===")

	// Проверяем что конфиг инициализирован
	if h.config == nil {
		logrus.Error("Config is nil in Login handler")
		h.errorHandler(gCtx, http.StatusInternalServerError, fmt.Errorf("configuration error"))
		return
	}

	// Получаем настройки JWT из конфига с защитой от nil
	var signingMethod jwt.SigningMethod
	var jwtSecret string
	var expiresIn time.Duration

	// Используем настройки из конфига
	if h.config.JWT.SigningMethod != nil {
		signingMethod = h.config.JWT.SigningMethod
	} else {
		// Fallback на HS256 если не настроено
		signingMethod = jwt.SigningMethodHS256
		logrus.Warn("JWT signing method not configured, using HS256 as fallback")
	}

	if h.config.JWT.Token != "" {
		jwtSecret = h.config.JWT.Token
	} else {
		// Fallback если токен не настроен
		jwtSecret = "fallback-secret-key-change-in-production"
		logrus.Warn("JWT token not configured, using fallback")
	}

	if h.config.JWT.ExpiresIn != 0 {
		expiresIn = h.config.JWT.ExpiresIn
	} else {
		// Fallback если время не настроено
		expiresIn = 24 * time.Hour
		logrus.Warn("JWT expires_in not configured, using 24h as fallback")
	}

	logrus.Infof("JWT settings from config: method=%v, expires=%v", signingMethod, expiresIn)

	req := &LoginRequest{}
	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		logrus.Errorf("Failed to parse request: %v", err)
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid request format"))
		return
	}

	logrus.Infof("Login attempt: %s", req.Login)

	if req.Login == hardcodedLogin && req.Password == hardcodedPassword {
		logrus.Info("Credentials valid, generating JWT token")

		token := jwt.NewWithClaims(signingMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(expiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "AwsProj",
			},
			UserUUID: uuid.New(),
			Scopes:   []string{"user"},
		})

		strToken, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			logrus.Errorf("Failed to sign token: %v", err)
			h.errorHandler(gCtx, http.StatusInternalServerError, fmt.Errorf("failed to generate token"))
			return
		}

		logrus.Info("JWT token generated successfully")
		gCtx.JSON(http.StatusOK, LoginResponse{
			ExpiresIn:   expiresIn.Milliseconds(),
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
		return
	}

	logrus.Warn("Invalid credentials provided")
	h.errorHandler(gCtx, http.StatusForbidden, fmt.Errorf("invalid credentials"))
}

// WithAuthCheck middleware для проверки JWT токена
func (h *Handler) WithAuthCheck() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// Проверяем что конфиг инициализирован
		if h.config == nil {
			logrus.Error("Config is nil in WithAuthCheck middleware")
			gCtx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		jwtStr := gCtx.GetHeader("Authorization")
		const jwtPrefix = "Bearer "

		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			logrus.Warn("No Bearer token in Authorization header")
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]

		// Используем секрет из конфига
		jwtSecret := h.config.JWT.Token
		if jwtSecret == "" {
			logrus.Error("JWT token not configured in middleware")
			gCtx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logrus.Warnf("JWT token validation failed: %v", err)
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !token.Valid {
			logrus.Warn("JWT token is invalid")
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Сохраняем claims в контекст для использования в handlers
		if claims, ok := token.Claims.(*ds.JWTClaims); ok {
			gCtx.Set("userUUID", claims.UserUUID)
			gCtx.Set("userScopes", claims.Scopes)
			logrus.Infof("User authenticated: %s", claims.UserUUID)
		}

		gCtx.Next()
	}
}

// Register обработчик регистрации
func (h *Handler) Register(gCtx *gin.Context) {
	req := &RegisterRequest{}
	if err := gCtx.ShouldBindJSON(req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid request: %v", err))
		return
	}

	// TODO: Реализовать регистрацию через UserService
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Registration endpoint - to be implemented",
		Data: RegisterResponse{
			ID:    1,
			Login: req.Login,
			Email: req.Email,
		},
	})
}

// GetUserProfile получение профиля пользователя
func (h *Handler) GetUserProfile(gCtx *gin.Context) {
	userUUID, exists := gCtx.Get("userUUID")
	if !exists {
		h.errorHandler(gCtx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	// TODO: Получить данные пользователя из UserService по userUUID
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User profile retrieved successfully",
		Data: UserProfileResponse{
			ID:       1,
			Login:    "current_user",
			Email:    "user@example.com",
			UserUUID: userUUID.(string),
		},
	})
}

// UpdateUser обновление данных пользователя
func (h *Handler) UpdateUser(gCtx *gin.Context) {
	userUUID, exists := gCtx.Get("userUUID")
	if !exists {
		h.errorHandler(gCtx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	req := &UpdateUserRequest{}
	if err := gCtx.ShouldBindJSON(req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid request: %v", err))
		return
	}

	// TODO: Реализовать обновление через UserService
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data: UserProfileResponse{
			ID:       1,
			Login:    "updated_user",
			Email:    "updated@example.com",
			UserUUID: userUUID.(string),
		},
	})
}

// Logout выход из системы
func (h *Handler) Logout(gCtx *gin.Context) {
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GetUserByID получение пользователя по ID
// GetUserByID получение пользователя по ID
func (h *Handler) GetUserByID(gCtx *gin.Context) {
	userID := gCtx.Param("id")

	// TODO: Реализовать через UserService с правильным преобразованием ID
	// Временная заглушка
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User data retrieved",
		Data: UserProfileResponse{
			ID:       1, // Временное значение, пока не реализована логика
			Login:    "user_" + userID,
			Email:    "user" + userID + "@example.com",
			UserUUID: "uuid-for-user-" + userID,
		},
	})
}
