package handler

import (
	"AwsProj/internal/app/service"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var registerReq service.RegisterRequest

	// Парсим тело запроса
	if err := ctx.ShouldBindJSON(&registerReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Регистрируем пользователя
	response, err := h.userService.Register(&registerReq)
	if err != nil {
		// Конфликт - пользователь уже существует
		if err.Error() == "пользователь с таким логином уже существует" {
			ctx.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "User already exists",
				"message": err.Error(),
			})
			return
		}
		// Ошибки валидации
		if err.Error() == "логин должен быть от 3 до 25 символов" ||
			err.Error() == "пароль должен быть не менее 6 символов" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Validation error",
				"message": err.Error(),
			})
			return
		}
		// Остальные ошибки
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Registration failed",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *UserHandler) GetUserProfile(ctx *gin.Context) {
	// Получаем ID пользователя из контекста (после аутентификации)
	// Пока используем хардкод для тестирования
	userID := uint(1) // В реальном приложении получаем из JWT или сессии

	profile, err := h.userService.GetUserProfile(userID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user profile",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

func (h *UserHandler) GetUserByID(ctx *gin.Context) {
	// Получаем ID из параметров пути
	idStr := ctx.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	profile, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	// Получаем ID текущего пользователя (хардкод для тестирования)
	userID := uint(1) // В реальном приложении из JWT

	// Парсим тело запроса
	var updateReq service.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем обновление
	err := h.userService.UpdateUser(userID, &updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "уже существует") {
			ctx.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Login already exists",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "должен быть") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Validation error",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update user",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Данные пользователя успешно обновлены",
		"data": gin.H{
			"user_id": userID,
		},
	})
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var loginReq service.LoginRequest

	// Парсим тело запроса
	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем аутентификацию
	response, err := h.userService.Login(&loginReq)
	if err != nil {
		if err.Error() == "неверный логин или пароль" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication failed",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Login failed",
			"message": err.Error(),
		})
		return
	}

	// В реальном приложении здесь генерируется JWT токен
	// Пока просто возвращаем данные пользователя

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
func (h *UserHandler) Logout(ctx *gin.Context) {
	// Получаем ID текущего пользователя (хардкод для тестирования)
	userID := uint(1) // В реальном приложении из JWT

	// Парсим тело запроса (опционально)
	var logoutReq service.LogoutRequest
	if err := ctx.ShouldBindJSON(&logoutReq); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем деавторизацию
	response, err := h.userService.Logout(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Logout failed",
			"message": err.Error(),
		})
		return
	}

	// В реальном приложении здесь:
	// - Удаляем токен из cookies
	// - Очищаем заголовки авторизации
	// - Инвалидируем сессию

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// В user_handler.go
func (h *UserHandler) TestAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Для тестирования - получаем userID из query параметра
		userIDStr := ctx.Query("user_id")
		if userIDStr != "" {
			if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				ctx.Set("userID", uint(userID))
			}
		} else {
			// По умолчанию userID = 1
			ctx.Set("userID", uint(1))
		}
		ctx.Next()
	}
}
