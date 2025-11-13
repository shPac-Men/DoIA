package handler

import (
	"AwsProj/internal/app/service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMixingPage - получение корзины текущего пользователя
func (h *Handler) GetMixingPage(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	result, err := h.MixingService.GetUserMixing(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get mixing data",
			Message: err.Error(),
		})
		return
	}

	handlerItems := make([]MixingItemResponse, len(result.Items))
	for i, item := range result.Items {
		handlerItems[i] = MixingItemResponse{
			ID:            item.ID,
			Title:         item.Title,
			Image:         item.Image,
			Ph:            item.PH,
			Concentration: item.Concentration,
			Volume:        item.Volume,
		}
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: MixingResponse{
			Items:      handlerItems,
			TotalItems: result.TotalItems,
			CartID:     result.CartID,
			UserID:     result.UserID,
		},
	})
}

// AddToMixing - добавление элемента в корзину
func (h *Handler) AddToMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req AddToMixingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	serviceReq := &service.AddToMixingRequest{
		ElementID: req.ElementID,
		Volume:    req.Volume,
	}

	result, err := h.MixingService.AddElementToMixing(userID, serviceReq)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") ||
			strings.Contains(err.Error(), "положительным") ||
			strings.Contains(err.Error(), "отрицательным") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Failed to add to mixing",
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

// RemoveFromMixing - удаление элемента из корзины
func (h *Handler) RemoveFromMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req RemoveFromMixingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	err := h.Repository.RemoveFromCart(userID, uint(req.ElementID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to remove from mixing",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Element removed successfully",
	})
}

// GetCartIcon - информация для иконки корзины
func (h *Handler) GetCartIcon(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusOK, CartIconResponse{
			DraftOrderID: 0,
			ItemsCount:   0,
		})
		return
	}

	response, err := h.MixingService.GetCartIcon(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get cart info",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateMixing - создание заказа из корзины (админ)
func (h *Handler) CreateMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req struct {
		AddedWater float64 `json:"added_water" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.AddedWater = 100.0
	}

	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, req.AddedWater)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to create mixing",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Mixing created successfully",
		Data: gin.H{
			"calculated_ph": calculatedPH,
			"added_water":   req.AddedWater,
			"user_id":       userID,
		},
	})
}

// GetMyMixedList - получение своих заказов
func (h *Handler) GetMyMixedList(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var filters service.MixedListRequest
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	mixedList, err := h.MixingService.GetMixedListByUser(userID, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get mixed list",
			Message: err.Error(),
		})
		return
	}

	handlerItems := make([]MixedListItem, len(mixedList))
	for i, item := range mixedList {
		handlerItems[i] = MixedListItem{
			ID:             item.ID,
			Status:         item.Status,
			DateCreate:     formatTime(item.DateCreate),
			DateUpdate:     formatTime(item.DateUpdate),
			DateFinish:     formatTimeOrEmpty(item.DateFinish),
			CreatorLogin:   item.CreatorLogin,
			ModeratorLogin: item.ModeratorLogin,
			Ph:             item.Ph,
			Concentration:  item.Concentration,
			TotalVolume:    item.TotalVolume,
			AddedWater:     item.AddedWater,
		}
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"items": handlerItems,
			"total": len(handlerItems),
		},
	})
}

// GetMyMixedByID - получение своего заказа по ID
func (h *Handler) GetMyMixedByID(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	serviceMixed, err := h.MixingService.GetMixedByID(uint(mixedID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Mixed not found",
		})
		return
	}

	currentLogin := h.auth.GetUserLogin(ctx) // ИЗМЕНЕНО: h.app → h.auth
	if serviceMixed.CreatorLogin != currentLogin {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
			Success: false,
			Error:   "Access denied",
			Message: "You can only view your own orders",
		})
		return
	}

	handlerMixed := convertMixedDetailToHandler(serviceMixed)
	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    handlerMixed,
	})
}

// GetMixedList - получение всех заказов (админ)
func (h *Handler) GetMixedList(ctx *gin.Context) {
	var filters service.MixedListRequest
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid query parameters",
		})
		return
	}

	mixedList, err := h.MixingService.GetMixedList(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get mixed list",
		})
		return
	}

	handlerItems := make([]MixedListItem, len(mixedList))
	for i, item := range mixedList {
		handlerItems[i] = MixedListItem{
			ID:             item.ID,
			Status:         item.Status,
			DateCreate:     formatTime(item.DateCreate),
			DateUpdate:     formatTime(item.DateUpdate),
			DateFinish:     formatTimeOrEmpty(item.DateFinish),
			CreatorLogin:   item.CreatorLogin,
			ModeratorLogin: item.ModeratorLogin,
			Ph:             item.Ph,
			Concentration:  item.Concentration,
			TotalVolume:    item.TotalVolume,
			AddedWater:     item.AddedWater,
		}
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"items": handlerItems,
			"total": len(handlerItems),
		},
	})
}

// GetMixedByID - получение заказа по ID (админ)
func (h *Handler) GetMixedByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	serviceMixed, err := h.MixingService.GetMixedByID(uint(mixedID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Mixed not found",
		})
		return
	}

	handlerMixed := convertMixedDetailToHandler(serviceMixed)
	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    handlerMixed,
	})
}

// UpdateMixed - обновление заказа (админ)
func (h *Handler) UpdateMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	var req UpdateMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	serviceReq := &service.UpdateMixedRequest{}
	if req.Status != nil {
		serviceReq.Status = *req.Status
	}
	if req.Concentration != nil {
		serviceReq.Concentration = *req.Concentration
	}
	if req.Ph != nil {
		serviceReq.Ph = *req.Ph
	}
	if req.TotalVolume != nil {
		serviceReq.TotalVolume = *req.TotalVolume
	}
	if req.AddedWater != nil {
		serviceReq.AddedWater = *req.AddedWater
	}

	err = h.MixingService.UpdateMixed(uint(mixedID), serviceReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to update mixed",
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Mixed updated successfully",
	})
}

// CompleteMixed - завершение заказа (админ)
func (h *Handler) CompleteMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	var req service.CompleteMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	result, err := h.MixingService.CompleteMixed(uint(mixedID), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to complete mixed",
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// DeleteMixed - удаление заказа (админ)
func (h *Handler) DeleteMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	var req service.DeleteMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		req.HardDelete = false
	}

	result, err := h.MixingService.DeleteMixed(uint(mixedID), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to delete mixed",
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
	})
}

// DeleteFromMixed - удаление элемента из заказа (админ)
func (h *Handler) DeleteFromMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid mixed ID",
		})
		return
	}

	var req service.DeleteFromMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	result, err := h.MixingService.DeleteFromMixed(uint(mixedID), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to delete from mixed",
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: result.Message,
	})
}

// Вспомогательные функции
func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z")
}

func formatTimeOrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func convertMixedDetailToHandler(serviceMixed *service.MixedDetailResponse) MixedDetailResponse {
	handlerItems := make([]MixedDetailItem, len(serviceMixed.Items))
	for i, item := range serviceMixed.Items {
		handlerItems[i] = MixedDetailItem{
			ElementID:     item.ElementID,
			Title:         item.Title,
			Image:         item.Image,
			Ph:            item.PH,
			Concentration: item.Concentration,
			Volume:        item.Volume,
			Comment:       item.Comment,
		}
	}

	dateFinish := ""
	if serviceMixed.DateFinish != nil {
		dateFinish = serviceMixed.DateFinish.Format("2006-01-02T15:04:05Z")
	}

	return MixedDetailResponse{
		ID:             serviceMixed.ID,
		Status:         serviceMixed.Status,
		DateCreate:     formatTime(serviceMixed.DateCreate),
		DateUpdate:     formatTime(serviceMixed.DateUpdate),
		DateFinish:     dateFinish,
		CreatorLogin:   serviceMixed.CreatorLogin,
		ModeratorLogin: serviceMixed.ModeratorLogin,
		Ph:             serviceMixed.Ph,
		Concentration:  serviceMixed.Concentration,
		TotalVolume:    serviceMixed.TotalVolume,
		AddedWater:     serviceMixed.AddedWater,
		Items:          handlerItems,
	}
}
