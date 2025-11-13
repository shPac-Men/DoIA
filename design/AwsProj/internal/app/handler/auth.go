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

	// Конвертируем в service request
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
	}

	registerResp, err := h.UserService.Register(serviceReq)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "пользователь с таким логином уже существует" {
			status = http.StatusConflict
		}

		gCtx.JSON(status, ErrorResponse{
			Success: false,
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	gCtx.JSON(http.StatusCreated, RegisterResponse{
		User: UserInfo{
			ID:          registerResp.ID,
			Login:       registerResp.Login,
			Role:        "visitor",
			IsModerator: registerResp.IsModerator,
		},
		Message: registerResp.Message,
	})
}

// Logout - выход
func (h *Handler) Logout(gCtx *gin.Context) {
	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}
