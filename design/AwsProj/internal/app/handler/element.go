package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllElements(ctx *gin.Context) {
	var elements []ds.Elements
	var err error

	search := ctx.Query("query")
	if search == "" {
		elements, err = h.Repository.GetAllElements()
	} else {
		elements, err = h.Repository.SearchElementByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get elements",
			"message": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Преобразуем в JSON-ответ
	var response []ElementResponse
	for _, elem := range elements {
		response = append(response, ElementResponse{
			ID:            elem.ID,
			Image:         elem.Img,
			Title:         elem.Name,
			Concentration: formatConcentration(elem.Concentration),
			PH:            formatPH(elem.Ph),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": response,
		"total": len(response),
		"query": search,
	})
}

// Вспомогательные функции для форматирования
func formatConcentration(conc float32) string {
	return fmt.Sprintf("%.2fM", conc)
}

func formatPH(ph float32) string {
	return fmt.Sprintf("%.1f", ph)
}

func (h *Handler) GetElementById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"message": "ID must be a number",
		})
		logrus.Error("Invalid ID format:", err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "Element not found",
			"message": err.Error(),
		})
		logrus.Error("Element not found:", err)
		return
	}

	// Сериализация в JSON-ответ
	response := ElementResponse{
		ID:            element.ID,
		Image:         element.Img,
		Title:         element.Name,
		Concentration: formatConcentration(element.Concentration),
		PH:            formatPH(element.Ph),
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) CreateElement(ctx *gin.Context) {
	var req service.CreateElementRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Вызов сервиса
	result, err := h.ElementService.CreateElement(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "уже существует") ||
			strings.Contains(err.Error(), "обязательно") ||
			strings.Contains(err.Error(), "диапазон") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to create element",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusCreated, result)
}

func (h *Handler) GetMixingPage(ctx *gin.Context) {
	userID := uint(1) // потом из аутентификации

	result, err := h.MixingService.GetUserMixing(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get mixing data",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":       result.Items,
		"total_items": result.TotalItems,
		"cart_id":     result.CartID,
		"user_id":     result.UserID,
	})
}

func (h *Handler) CreateMixing(ctx *gin.Context) {
	// 1. Получаем данные из JSON тела запроса
	var request struct {
		AddedWater float64 `json:"added_water" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Если JSON невалидный, используем значение по умолчанию
		request.AddedWater = 100.0
	}

	userID := uint(1)

	// 2. Вызываем метод репозитория
	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, request.AddedWater)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create mixing",
			"message": err.Error(),
		})
		return
	}

	// 3. Возвращаем JSON результат вместо HTML
	ctx.JSON(http.StatusOK, gin.H{
		"calculated_ph": calculatedPH,
		"added_water":   request.AddedWater,
		"user_id":       userID,
	})
}

func (h *Handler) RemoveFromMixing(ctx *gin.Context) {
	// Получаем ID элемента из формы
	elementIDStr := ctx.PostForm("element_id")
	elementID, err := strconv.Atoi(elementIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
		return
	}

	// Захардкоженный пользователь
	userID := uint(1)

	// Удаляем элемент из корзины
	err = h.Repository.RemoveFromCart(userID, uint(elementID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Перенаправляем обратно на страницу корзины
	ctx.Redirect(http.StatusFound, "/mixingpage")
}

func (h *Handler) UpdateElement(ctx *gin.Context) {
	// Получаем ID из URL
	strID := ctx.Param("id")
	id, err := strconv.ParseUint(strID, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid element ID",
			"message": "ID must be a number",
		})
		return
	}

	var req service.UpdateElementRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Вызов сервиса
	result, err := h.ElementService.UpdateElement(int(id), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "уже существует") ||
			strings.Contains(err.Error(), "диапазон") ||
			strings.Contains(err.Error(), "отрицательной") ||
			strings.Contains(err.Error(), "нет данных") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to update element",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteElement(ctx *gin.Context) {
	// Получаем ID из URL
	strID := ctx.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid element ID",
			"message": "ID must be a number",
		})
		return
	}

	// Вызов сервиса
	result, err := h.ElementService.DeleteElement(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to delete element",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) UploadImage(ctx *gin.Context) {
	// Получаем ID элемента
	strID := ctx.Param("id")
	elementID, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid element ID",
			"message": "ID must be a number",
		})
		return
	}

	// Получаем файл из формы
	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "File not found",
			"message": "Image file is required",
		})
		return
	}

	// Вызов сервиса
	result, err := h.ElementService.UploadImage(elementID, fileHeader)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "размер") ||
			strings.Contains(err.Error(), "поддерживаются") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to upload image",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusOK, result)
}
