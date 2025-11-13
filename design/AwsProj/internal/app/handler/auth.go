package handler

import (
	"AwsProj/internal/app/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login - аутентификация
func (h *Handler) Login(gCtx *gin.Context) {
	var req LoginRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	serviceReq := &service.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	}

	loginResp, err := h.UserService.Login(serviceReq)
	if err != nil {
		gCtx.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Authentication failed",
			Message: err.Error(),
		})
		return
	}

	gCtx.JSON(http.StatusOK, LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: 24 * 60 * 60 * 1000,
		TokenType: "Bearer",
		User: UserInfo{
			ID:          loginResp.ID,
			Login:       loginResp.Login,
			Role:        loginResp.Role,
			IsModerator: loginResp.IsModerator,
		},
	})
}

// Register - регистрация
func (h *Handler) Register(gCtx *gin.Context) {
	var req RegisterRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	serviceReq := &service.RegisterRequest{
		Login:    req.Login,
		Password: req.Password,
		// Email УДАЛЁН
	}

	registerResp, err := h.UserService.Register(serviceReq)
	if err != nil {
		gCtx.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	gCtx.JSON(http.StatusOK, RegisterResponse{
		//Success: true,
		Message: "Пользователь успешно зарегистрирован",
		User: UserInfo{
			ID:    registerResp.ID,
			Login: registerResp.Login,
			// Email УДАЛЁН
		},
	})
}

// Logout - выход
func (h *Handler) Logout(gCtx *gin.Context) {
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}
