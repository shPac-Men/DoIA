package handler

import (
	"AwsProj/internal/app/service"
	"net/http"

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
