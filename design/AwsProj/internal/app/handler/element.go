package handler

import (
	"net/http"
	"strconv"
	"strings"

	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetAllElements получение всех элементов
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
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get elements",
			Message: err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Преобразуем в JSON-ответ
	var response []ElementResponse
	for _, elem := range elements {
		response = append(response, ElementResponse{
			ID:            elem.ID,
			Name:          elem.Name,
			Description:   elem.Description,
			Ph:            elem.Ph,
			Concentration: elem.Concentration,
			Image:         elem.Img,
		})
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"items": response,
			"total": len(response),
			"query": search,
		},
	})
}

// GetElementById получение элемента по ID
func (h *Handler) GetElementById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid ID format",
			Message: "ID must be a number",
		})
		logrus.Error("Invalid ID format:", err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Element not found",
			Message: err.Error(),
		})
		logrus.Error("Element not found:", err)
		return
	}

	response := ElementResponse{
		ID:            element.ID,
		Name:          element.Name,
		Description:   element.Description,
		Ph:            element.Ph,
		Concentration: element.Concentration,
		Image:         element.Img,
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// CreateElement создание элемента (только для админов)
func (h *Handler) CreateElement(ctx *gin.Context) {
	var req CreateElementRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Message: err.Error(),
		})
		return
	}

	// Конвертируем в service request
	serviceReq := &service.CreateElementRequest{
		Name:          req.Name,
		Description:   req.Description,
		Ph:            req.Ph,
		Concentration: req.Concentration,
	}

	result, err := h.ElementService.CreateElement(serviceReq)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "уже существует") ||
			strings.Contains(err.Error(), "обязательно") ||
			strings.Contains(err.Error(), "диапазон") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Failed to create element",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// UpdateElement обновление элемента (только для админов)
func (h *Handler) UpdateElement(ctx *gin.Context) {
	strID := ctx.Param("id")
	id, err := strconv.ParseUint(strID, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid element ID",
			Message: "ID must be a number",
		})
		return
	}

	var req UpdateElementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Message: err.Error(),
		})
		return
	}

	// Конвертируем в service request
	serviceReq := &service.UpdateElementRequest{
		Name:          req.Name,
		Description:   req.Description,
		Ph:            req.Ph,
		Concentration: req.Concentration,
	}

	result, err := h.ElementService.UpdateElement(int(id), serviceReq)
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

		ctx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Failed to update element",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// DeleteElement удаление элемента (только для админов)
func (h *Handler) DeleteElement(ctx *gin.Context) {
	strID := ctx.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid element ID",
			Message: "ID must be a number",
		})
		return
	}

	result, err := h.ElementService.DeleteElement(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Failed to delete element",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// UploadImage загрузка изображения для элемента (только для админов)
func (h *Handler) UploadImage(ctx *gin.Context) {
	strID := ctx.Param("id")
	elementID, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid element ID",
			Message: "ID must be a number",
		})
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "File not found",
			Message: "Image file is required",
		})
		return
	}

	result, err := h.ElementService.UploadImage(elementID, fileHeader)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "размер") ||
			strings.Contains(err.Error(), "поддерживаются") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Failed to upload image",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}
