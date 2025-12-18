package handler

import (
	"AwsProj/internal/app/service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetMixingPage godoc
// @Summary Get user's mixing cart
// @Description Get current user's mixing cart with all added elements
// @Tags mixing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MixingResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mixing [get]
func (h *Handler) GetMixingPage(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ GetMixingPage: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	logrus.Infof("📦 GetMixingPage for user %d", userID)

	result, err := h.MixingService.GetUserMixing(userID)
	if err != nil {
		logrus.Errorf("❌ Failed to get mixing: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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

	logrus.Infof("✅ Cart retrieved: %d items", len(handlerItems))

	ctx.JSON(http.StatusOK, MixingResponse{
		Items:      handlerItems,
		TotalItems: result.TotalItems,
		CartID:     result.CartID,
		UserID:     result.UserID,
	})
}

// AddToMixing godoc
// @Summary Add element to mixing cart
// @Description Add a chemical element to user's mixing cart with specified volume
// @Tags mixing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToMixingRequest true "Element to add with volume"
// @Success 200 {object} service.AddToMixingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /mixing/items [post]
func (h *Handler) AddToMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ AddToMixing: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req AddToMixingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ AddToMixing: Invalid request - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("➕ AddToMixing: user=%d, element=%d, volume=%.2f", userID, req.ElementID, req.Volume)

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
			logrus.Warnf("⚠️ Validation error: %v", err)
		}

		ctx.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Element added to cart for user %d", userID)
	ctx.JSON(http.StatusOK, result)
}

// RemoveFromMixing godoc
// @Summary Remove element from mixing cart
// @Description Remove a chemical element from user's mixing cart
// @Tags mixing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RemoveFromMixingRequest true "Element ID to remove"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /mixing/remove [post]
func (h *Handler) RemoveFromMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ RemoveFromMixing: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req RemoveFromMixingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ RemoveFromMixing: Invalid request - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("🗑️ RemoveFromMixing: user=%d, element=%d", userID, req.ElementID)

	err := h.Repository.RemoveFromCart(userID, uint(req.ElementID))
	if err != nil {
		logrus.Errorf("❌ Failed to remove: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Element removed from cart for user %d", userID)
	ctx.Status(http.StatusNoContent)
}

// GetCartIcon godoc
// @Summary Get cart icon info
// @Description Get cart item count for display in UI (public endpoint, works for guests too)
// @Tags mixing
// @Accept json
// @Produce json
// @Security BearerAuth  <--- ДОБАВИТЬ ЭТУ СТРОКУ
// @Success 200 {object} CartIconResponse
// @Router /mixing/cart-icon [get]
func (h *Handler) GetCartIcon(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)

	if userID == 0 {
		logrus.Debug("📍 GetCartIcon: Guest user")
		ctx.JSON(http.StatusOK, CartIconResponse{
			DraftOrderID: 0,
			ItemsCount:   0,
		})
		return
	}

	logrus.Debugf("📍 GetCartIcon for user %d", userID)

	response, err := h.MixingService.GetCartIcon(userID)
	if err != nil {
		logrus.Errorf("❌ Failed to get cart icon: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateMixing godoc
// @Summary Create mixing order from cart
// @Description Create a new mixing order from user's cart (Admin only)
// @Tags mixing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateMixingRequest true "Mixing order data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /admin/mixed [post]
func (h *Handler) CreateMixing(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ CreateMixing: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req struct {
		AddedWater float64 `json:"added_water" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.AddedWater = 100.0
	}

	logrus.Infof("➕ CreateMixing: user=%d, added_water=%.2f", userID, req.AddedWater)

	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, req.AddedWater)
	if err != nil {
		logrus.Errorf("❌ Failed to create mixing: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Mixing created: user=%d, pH=%.2f", userID, calculatedPH)

	ctx.JSON(http.StatusOK, gin.H{
		"calculated_ph": calculatedPH,
		"added_water":   req.AddedWater,
		"user_id":       userID,
	})
}

// GetMyMixedList godoc
// @Summary Get user's mixing orders
// @Description Get list of all mixing orders created by current user
// @Tags mixed
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (draft, pending, completed)"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset results (default: 0)"
// @Success 200 {array} MixedListItem
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mixed/my [get]
func (h *Handler) GetMyMixedList(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ GetMyMixedList: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var filters service.MixedListRequest
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		logrus.Warnf("❌ GetMyMixedList: Invalid filters - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("📋 GetMyMixedList: user=%d", userID)

	mixedList, err := h.MixingService.GetMixedListByUser(userID, filters)
	if err != nil {
		logrus.Errorf("❌ Failed to get mixed list: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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
			ItemsCount:     item.ItemsCount, // <--- ВОТ ЗДЕСЬ
		}
	}

	logrus.Infof("✅ Retrieved %d orders for user %d", len(handlerItems), userID)
	ctx.JSON(http.StatusOK, handlerItems)
}

// GetMyMixedByID godoc
// @Summary Get user's mixing order by ID
// @Description Get detailed info about user's specific mixing order
// @Tags mixed
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} MixedDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /mixed/my/{id} [get]
func (h *Handler) GetMyMixedByID(ctx *gin.Context) {
	userID := h.auth.GetUserID(ctx)
	if userID == 0 {
		logrus.Warn("❌ GetMyMixedByID: Unauthorized access")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ GetMyMixedByID: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	logrus.Infof("📖 GetMyMixedByID: user=%d, mixed=%d", userID, mixedID)

	serviceMixed, err := h.MixingService.GetMixedByID(uint(mixedID))
	if err != nil {
		logrus.Errorf("❌ Order not found: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Mixed not found",
		})
		return
	}

	currentLogin := h.auth.GetUserLogin(ctx)
	if serviceMixed.CreatorLogin != currentLogin {
		logrus.Warnf("⚠️ Access denied: user=%d tried to access order=%d (creator=%s)", userID, mixedID, serviceMixed.CreatorLogin)
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "You can only view your own orders",
		})
		return
	}

	handlerMixed := convertMixedDetailToHandler(serviceMixed)
	logrus.Infof("✅ Retrieved order %d for user %d", mixedID, userID)

	ctx.JSON(http.StatusOK, handlerMixed)
}

// GetMixedList godoc
// @Summary Get all mixing orders
// @Description Get list of all mixing orders in system (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param creator query string false "Filter by creator login"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset results (default: 0)"
// @Success 200 {array} MixedListItem
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /admin/mixed [get]
func (h *Handler) GetMixedList(ctx *gin.Context) {
	var filters service.MixedListRequest
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		logrus.Warnf("❌ GetMixedList: Invalid filters - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query parameters",
		})
		return
	}

	logrus.Infof("📋 GetMixedList: admin request")

	mixedList, err := h.MixingService.GetMixedList(filters)
	if err != nil {
		logrus.Errorf("❌ Failed to get mixed list: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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

	logrus.Infof("✅ Retrieved %d orders (admin)", len(handlerItems))
	ctx.JSON(http.StatusOK, handlerItems)
}

// GetMixedByID godoc
// @Summary Get mixing order by ID
// @Description Get detailed info about any mixing order (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} MixedDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/mixed/{id} [get]
func (h *Handler) GetMixedByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ GetMixedByID: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	logrus.Infof("📖 GetMixedByID: admin request, mixed=%d", mixedID)

	serviceMixed, err := h.MixingService.GetMixedByID(uint(mixedID))
	if err != nil {
		logrus.Errorf("❌ Order not found: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Mixed not found",
		})
		return
	}

	handlerMixed := convertMixedDetailToHandler(serviceMixed)
	logrus.Infof("✅ Retrieved order %d (admin)", mixedID)

	ctx.JSON(http.StatusOK, handlerMixed)
}

// UpdateMixed godoc
// @Summary Update mixing order
// @Description Update mixing order parameters (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body UpdateMixedRequest true "Updated order data"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/mixed/{id} [put]
func (h *Handler) UpdateMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ UpdateMixed: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	var req UpdateMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ UpdateMixed: Invalid request - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	logrus.Infof("✏️ UpdateMixed: admin, mixed=%d", mixedID)

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
		logrus.Errorf("❌ Failed to update: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Order updated: mixed=%d", mixedID)
	ctx.Status(http.StatusNoContent)
}

// CompleteMixed godoc
// @Summary Complete mixing order
// @Description Mark mixing order as completed (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body service.CompleteMixedRequest true "Completion data"
// @Success 200 {object} service.CompleteMixedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/mixed/{id}/complete [put]
func (h *Handler) CompleteMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ CompleteMixed: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	var req service.CompleteMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		logrus.Warnf("❌ CompleteMixed: Invalid request - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	logrus.Infof("✅ CompleteMixed: admin, mixed=%d", mixedID)

	result, err := h.MixingService.CompleteMixed(uint(mixedID), &req)
	if err != nil {
		logrus.Errorf("❌ Failed to complete: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Order completed: mixed=%d", mixedID)
	ctx.JSON(http.StatusOK, result)
}

// DeleteMixed godoc
// @Summary Delete mixing order
// @Description Soft or hard delete mixing order (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body service.DeleteMixedRequest true "Delete options"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/mixed/{id} [delete]
func (h *Handler) DeleteMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ DeleteMixed: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	var req service.DeleteMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		req.HardDelete = false
	}

	logrus.Infof("🗑️ DeleteMixed: admin, mixed=%d, hard=%v", mixedID, req.HardDelete)

	_, err = h.MixingService.DeleteMixed(uint(mixedID), &req)
	if err != nil {
		logrus.Errorf("❌ Failed to delete: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Order deleted: mixed=%d", mixedID)
	ctx.Status(http.StatusNoContent)
}

// DeleteFromMixed godoc
// @Summary Delete element from order
// @Description Remove a chemical element from mixing order (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body service.DeleteFromMixedRequest true "Element to remove"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/mixed/{id}/items [delete]
func (h *Handler) DeleteFromMixed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Warnf("❌ DeleteFromMixed: Invalid ID - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mixed ID",
		})
		return
	}

	var req service.DeleteFromMixedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ DeleteFromMixed: Invalid request - %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	logrus.Infof("🗑️ DeleteFromMixed: admin, mixed=%d", mixedID)

	_, err = h.MixingService.DeleteFromMixed(uint(mixedID), &req)
	if err != nil {
		logrus.Errorf("❌ Failed to delete from mixed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ Element removed from order: mixed=%d", mixedID)
	ctx.Status(http.StatusNoContent)
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
		ItemsCount:     serviceMixed.ItemsCount, // <--- ДОБАВИТЬ ВОТ ЭТО
	}
}
