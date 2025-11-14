package handler

import (
	"AwsProj/internal/app/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUserProfile - получение профиля текущего пользователя

// GetUserProfile godoc
// @Summary Get current user profile
// @Description Get profile information of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse "User profile retrieved"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/profile [get]
func (h *Handler) GetUserProfile(gCtx *gin.Context) {
	userID, exists := gCtx.Get("user_id")
	if !exists {
		logrus.Warn("❌ GetUserProfile: Unauthorized access")
		gCtx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userIDVal := userID.(uint)
	logrus.Infof("👤 GetUserProfile: user=%d", userIDVal)

	profile, err := h.UserService.GetUserProfile(userIDVal)
	if err != nil {
		logrus.Errorf("❌ User not found: %v", err)
		gCtx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Not found",
			Message: err.Error(),
		})
		return
	}

	logrus.Infof("✅ Profile retrieved: user=%d, login=%s", profile.ID, profile.Login)

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

// UpdateUser godoc
// @Summary Update user profile
// @Description Update login and/or password of current user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserRequest true "Updated user data (all fields optional)"
// @Success 200 {object} SuccessResponse "User updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 409 {object} ErrorResponse "Login already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/profile [put]
func (h *Handler) UpdateUser(gCtx *gin.Context) {
	userID, exists := gCtx.Get("user_id")
	if !exists {
		logrus.Warn("❌ UpdateUser: Unauthorized access")
		gCtx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req UpdateUserRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ UpdateUser: Invalid request - %v", err)
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	userIDVal := userID.(uint)
	logrus.Infof("✏️ UpdateUser: user=%d", userIDVal)

	// Конвертируем в service request
	serviceReq := &service.UpdateUserRequest{}
	if req.Login != nil {
		serviceReq.Login = *req.Login
		logrus.Debugf("   Updating login to: %s", *req.Login)
	}
	if req.Password != nil {
		serviceReq.Password = *req.Password
		logrus.Debug("   Updating password")
	}

	err := h.UserService.UpdateUser(userIDVal, serviceReq)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Проверяем тип ошибки
		errMsg := err.Error()
		if errMsg == "login already exists" || errMsg == "логин уже существует" {
			statusCode = http.StatusConflict // 409
			logrus.Warnf("⚠️ UpdateUser: Login already exists")
		} else if errMsg == "invalid login format" || errMsg == "invalid password" {
			statusCode = http.StatusBadRequest // 400
			logrus.Warnf("⚠️ UpdateUser: Validation error - %v", err)
		}

		gCtx.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   "Update failed",
			Message: err.Error(),
		})
		return
	}

	logrus.Infof("✅ User updated: user=%d", userIDVal)

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// GetUserByID - получение пользователя по ID

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get user profile information by user ID (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} SuccessResponse "User retrieved"
// @Failure 400 {object} ErrorResponse "Invalid user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - Admin access required"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/users/{id} [get]
func (h *Handler) GetUserByID(gCtx *gin.Context) {
	userID := gCtx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		logrus.Warnf("❌ GetUserByID: Invalid ID - %v", err)
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	logrus.Infof("👤 GetUserByID: admin request, user=%d", id)

	profile, err := h.UserService.GetUserByID(id)
	if err != nil {
		logrus.Errorf("❌ User not found: %v", err)
		gCtx.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	logrus.Infof("✅ User retrieved: id=%d, login=%s", profile.ID, profile.Login)

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

// GetAllUsers godoc
// @Summary Get all users
// @Description Get list of all users in system (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse "Users list retrieved"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/users [get]
func (h *Handler) GetAllUsers(gCtx *gin.Context) {
	logrus.Info("📋 GetAllUsers: admin request")

	users, err := h.UserService.GetAllUsers()
	if err != nil {
		logrus.Errorf("❌ Failed to get users: %v", err)
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

	logrus.Infof("✅ Retrieved %d users", len(result))

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateUserRole - изменение роли пользователя (админ)

// UpdateUserRole godoc
// @Summary Update user role
// @Description Grant or revoke moderator role for a user (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserRoleRequest true "Role update data"
// @Success 200 {object} SuccessResponse "User role updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID or request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - Admin access required"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/users/{id}/role [put]
func (h *Handler) UpdateUserRole(gCtx *gin.Context) {
	userID := gCtx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		logrus.Warnf("❌ UpdateUserRole: Invalid ID - %v", err)
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req UpdateUserRoleRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ UpdateUserRole: Invalid request - %v", err)
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	logrus.Infof("👑 UpdateUserRole: admin request, user=%d, is_moderator=%v", id, req.IsModerator)

	err := h.UserService.UpdateUserRole(id, req.IsModerator)
	if err != nil {
		logrus.Errorf("❌ Failed to update role: %v", err)
		gCtx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to update role",
		})
		return
	}

	roleStr := "user"
	if req.IsModerator {
		roleStr = "moderator"
	}

	logrus.Infof("✅ User role updated: user=%d, role=%s", id, roleStr)

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}
