package handler

import (
	"AwsProj/internal/app/service"
	"net/http"

	"github.com/gin-contrib/sessions"
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

	// Сохраняем сессию в Redis (для браузера через куку)
	session := sessions.Default(gCtx)
	session.Set("user_id", loginResp.ID)
	session.Set("login", loginResp.Login)
	session.Set("role", loginResp.Role)
	session.Set("is_moderator", loginResp.IsModerator)
	if err := session.Save(); err != nil {
		gCtx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Session save failed",
			Message: err.Error(),
		})
		return
	}

	// Возвращаем JWT для REST клиентов
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
		},
	})
}

// Logout - выход
func (h *Handler) Logout(gCtx *gin.Context) {
	session := sessions.Default(gCtx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()

	gCtx.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}
