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

// GetAllElements godoc
// @Summary Get all elements
// @Description Get list of all elements with optional search by name
// @Tags elements
// @Accept json
// @Produce json
// @Param query query string false "Search query by element name"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/elements [get]
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
			"error":   true,
			"message": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Преобразуем в JSON-ответ
	//сериализация данных из бд в джсон
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
		"success": true,
		"data":    response,
		"meta": gin.H{
			"total":    len(response),
			"query":    search,
			"has_more": false, // можно добавить пагинацию позже
		},
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
			"success": false,
			"error":   "Invalid ID format",
			"message": "ID must be a number",
		})
		logrus.Error("Invalid ID format:", err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
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

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *Handler) CreateElement(ctx *gin.Context) {
	var req service.CreateElementRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверные данные запроса",
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
			"success": false,
			"error":   "Ошибка создания элемента",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": result.Message,
		"data":    result,
	})
}

// deepseek
// func (h *Handler) AddToMixing(ctx *gin.Context) {
// 	// 1. Получаем ID элемента из JSON тела запроса
// 	var request struct {
// 		ElementID int     `json:"element_id" binding:"required"`
// 		Volume    float32 `json:"volume,omitempty"` // опционально
// 	}

// 	if err := ctx.ShouldBindJSON(&request); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"success": false,
// 			"error":   "Invalid request data",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	// 2. Получаем ID пользователя (пока хардкод, потом через аутентификацию)
// 	userID := uint(1)

// 	// 3. Объём по умолчанию
// 	volume := request.Volume
// 	if volume == 0 {
// 		volume = 100.0 // значение по умолчанию
// 	}

// 	// 4. Добавляем элемент в корзину
// 	err := h.Repository.AddElementToCart(userID, uint(request.ElementID), volume)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   "Failed to add element to mixing",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	// 5. Возвращаем JSON ответ вместо редиректа
// 	ctx.JSON(http.StatusOK, gin.H{
// 		"success": true,
// 		"message": "Element added to mixing successfully",
// 		"data": gin.H{
// 			"element_id": request.ElementID,
// 			"volume":     volume,
// 			"user_id":    userID,
// 		},
// 	})
// }

// internal/app/handler/element.go
func (h *Handler) GetMixingPage(ctx *gin.Context) {
	userID := uint(1) // потом из аутентификации

	result, err := h.MixingService.GetUserMixing(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get mixing data",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Items,
		"meta": gin.H{
			"total_items": result.TotalItems,
			"cart_id":     result.CartID,
			"user_id":     result.UserID,
		},
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
			"success": false,
			"error":   "Failed to create mixing",
			"message": err.Error(),
		})
		return
	}

	// 3. Возвращаем JSON результат вместо HTML
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mixing successfully created",
		"data": gin.H{
			"calculated_ph": calculatedPH,
			"added_water":   request.AddedWater,
			"user_id":       userID,
		},
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
			"success": false,
			"error":   "Неверный ID элемента",
			"message": "ID должен быть числом",
		})
		return
	}

	var req service.UpdateElementRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверные данные запроса",
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
			"success": false,
			"error":   "Ошибка обновления элемента",
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

func (h *Handler) DeleteElement(ctx *gin.Context) {
	// Получаем ID из URL
	strID := ctx.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный ID элемента",
			"message": "ID должен быть числом",
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
			"success": false,
			"error":   "Ошибка удаления элемента",
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

func (h *Handler) UploadImage(ctx *gin.Context) {
	// Получаем ID элемента
	strID := ctx.Param("id")
	elementID, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный ID элемента",
			"message": "ID должен быть числом",
		})
		return
	}

	// Получаем файл из формы
	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Файл не найден",
			"message": "Необходимо загрузить изображение",
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
			"success": false,
			"error":   "Ошибка загрузки изображения",
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
