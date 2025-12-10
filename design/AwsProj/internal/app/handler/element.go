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

// GetAllElements godoc
// @Summary Get all elements
// @Description Get list of all elements with optional search by name
// @Tags elements
// @Accept json
// @Produce json
// @Param query query string false "Search query by element name"
// @Success 200 {array} ElementResponse
// @Failure 500 {object} ErrorResponse
// @Router /elements [get]
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
			"error": "Failed to get elements",
		})
		logrus.Error(err)
		return
	}

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

	ctx.JSON(http.StatusOK, response)
}

// GetElementById godoc
// @Summary Get element by ID
// @Description Retrieve a specific element by its ID
// @Tags elements
// @Accept json
// @Produce json
// @Param id path int true "Element ID"
// @Success 200 {object} ElementResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /elements/{id} [get]
func (h *Handler) GetElementById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		logrus.Error("Invalid ID format:", err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Element not found",
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

	ctx.JSON(http.StatusOK, response)
}

// CreateElement godoc
// @Summary Create new element
// @Description Create a new chemical element (Admin only)
// @Tags elements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateElementRequest true "Element data"
// @Success 201 {object} service.CreateElementResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /elements [post]
func (h *Handler) CreateElement(ctx *gin.Context) {
	var req CreateElementRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

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

		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

// UpdateElement godoc
// @Summary Update element
// @Description Update an existing chemical element (Admin only)
// @Tags elements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Element ID"
// @Param request body UpdateElementRequest true "Updated element data"
// @Success 200 {object} service.UpdateElementResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /elements/{id} [put]
func (h *Handler) UpdateElement(ctx *gin.Context) {
	strID := ctx.Param("id")
	id, err := strconv.ParseUint(strID, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid element ID",
		})
		return
	}

	var req UpdateElementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

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

		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// DeleteElement godoc
// @Summary Delete element
// @Description Delete a chemical element (Admin only)
// @Tags elements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Element ID"
// @Success 200 {object} service.DeleteElementResponse
// @Failure 404 {object} ErrorResponse
// @Router /elements/{id} [delete]
func (h *Handler) DeleteElement(ctx *gin.Context) {
	strID := ctx.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid element ID",
		})
		return
	}

	result, err := h.ElementService.DeleteElement(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// UploadImage godoc
// @Summary Upload element image
// @Description Upload an image for a chemical element (Admin only)
// @Tags elements
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Element ID"
// @Param image formData file true "Image file"
// @Success 200 {object} service.UploadImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /elements/{id}/image [post]
func (h *Handler) UploadImage(ctx *gin.Context) {
	strID := ctx.Param("id")
	elementID, err := strconv.Atoi(strID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid element ID",
		})
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Image file is required",
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

		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
