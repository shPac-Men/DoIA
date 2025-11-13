package handler

import (
	"AwsProj/internal/app/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserProfile - получение профиля текущего пользователя
func (h *Handler) GetUserProfile(gCtx *gin.Context) {
	userID, exists := gCtx.Get("user_id")
	if !exists {
		gCtx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	profile, err := h.UserService.GetUserProfile(userID.(uint))
	if err != nil {
		gCtx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Not found",
			Message: err.Error(),
		})
		return
	}

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: UserProfileResponse{
			ID:          profile.ID,
			Login:       profile.Login,
			Role:        profile.Role,
			IsModerator: profile.IsModerator,
		},
	})
}

// UpdateUser - обновление данных пользователя
func (h *Handler) UpdateUser(gCtx *gin.Context) {
	userID, exists := gCtx.Get("user_id")
	if !exists {
		gCtx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req UpdateUserRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Конвертируем в service request
	serviceReq := &service.UpdateUserRequest{}
	if req.Login != nil {
		serviceReq.Login = *req.Login
	}
	if req.Password != nil {
		serviceReq.Password = *req.Password
	}

	err := h.UserService.UpdateUser(userID.(uint), serviceReq)
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Update failed",
			Message: err.Error(),
		})
		return
	}

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// GetUserByID - получение пользователя по ID
func (h *Handler) GetUserByID(gCtx *gin.Context) {
	userID := gCtx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	profile, err := h.UserService.GetUserByID(id)
	if err != nil {
		gCtx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: UserProfileResponse{
			ID:          profile.ID,
			Login:       profile.Login,
			Role:        profile.Role,
			IsModerator: profile.IsModerator,
		},
	})
}

// GetAllUsers - получение всех пользователей (админ)
func (h *Handler) GetAllUsers(gCtx *gin.Context) {
	users, err := h.UserService.GetAllUsers()
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to get users",
		})
		return
	}

	result := make([]UserProfileResponse, len(users))
	for i, user := range users {
		result[i] = UserProfileResponse{
			ID:          user.ID,
			Login:       user.Login,
			Role:        user.Role,
			IsModerator: user.IsModerator,
		}
	}

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateUserRole - изменение роли пользователя (админ)
func (h *Handler) UpdateUserRole(gCtx *gin.Context) {
	userID := gCtx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req UpdateUserRoleRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	err := h.UserService.UpdateUserRole(id, req.IsModerator)
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to update role",
		})
		return
	}

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}
