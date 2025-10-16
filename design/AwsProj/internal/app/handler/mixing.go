package handler

import (
	"AwsProj/internal/app/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AddToMixing(ctx *gin.Context) {
	var req service.AddToMixingRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверные данные запроса",
			"message": err.Error(),
		})
		return
	}

	// Получаем ID пользователя (пока хардкод)
	userID := uint(1)

	// Вызов сервиса
	result, err := h.MixingService.AddElementToMixing(userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") ||
			strings.Contains(err.Error(), "положительным") ||
			strings.Contains(err.Error(), "отрицательным") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"success": false,
			"error":   "Ошибка добавления в корзину",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": result.Message,
		"data":    result,
	})
}
