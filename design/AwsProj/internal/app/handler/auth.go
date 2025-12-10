package handler

import (
	"AwsProj/internal/app/service"
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Login godoc
// @Summary User login
// @Description Authenticate user with credentials and receive JWT token + session cookie
// @Description JWT token is saved in Redis for revocation support
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(gCtx *gin.Context) {
	var req LoginRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ Login: Invalid request - %v", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	logrus.Infof("🔓 Login attempt: %s", req.Login)

	serviceReq := &service.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	}

	loginResp, err := h.UserService.Login(serviceReq)
	if err != nil {
		logrus.Warnf("❌ Login failed for user %s: %v", req.Login, err)
		gCtx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid login or password",
		})
		return
	}

	// Сохраняем JWT в Redis с TTL
	ctx, cancel := context.WithTimeout(gCtx.Request.Context(), 5*time.Second)
	defer cancel()

	tokenTTL := 24 * time.Hour
	if err := h.redisTokenService.SaveToken(ctx, loginResp.ID, loginResp.Token, tokenTTL); err != nil {
		logrus.Errorf("❌ Failed to save token to Redis: %v", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save authentication token",
		})
		return
	}

	logrus.Infof("✅ Token saved to Redis for user %d (%s)", loginResp.ID, loginResp.Login)

	// Сохраняем сессию в Redis (для браузера через куку)
	session := sessions.Default(gCtx)
	session.Set("user_id", loginResp.ID)
	session.Set("login", loginResp.Login)
	session.Set("role", loginResp.Role)
	session.Set("is_moderator", loginResp.IsModerator)
	if err := session.Save(); err != nil {
		logrus.Errorf("❌ Failed to save session: %v", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save session",
		})
		return
	}

	logrus.Infof("✅ Session created for user %d", loginResp.ID)

	// Возвращаем JWT для REST клиентов
	gCtx.JSON(http.StatusOK, LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: 24 * 60 * 60 * 1000, // 24 часа в миллисекундах
		TokenType: "Bearer",
		User: UserInfo{
			ID:          loginResp.ID,
			Login:       loginResp.Login,
			Role:        loginResp.Role,
			IsModerator: loginResp.IsModerator,
		},
	})
}

// Register godoc
// @Summary User registration
// @Description Create a new user account with login and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} UserInfo "User successfully registered"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(gCtx *gin.Context) {
	var req RegisterRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		logrus.Warnf("❌ Register: Invalid request - %v", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("📝 Register attempt: %s", req.Login)

	serviceReq := &service.RegisterRequest{
		Login:    req.Login,
		Password: req.Password,
	}

	registerResp, err := h.UserService.Register(serviceReq)
	if err != nil {
		logrus.Warnf("❌ Register failed for %s: %v", req.Login, err)
		gCtx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("✅ User registered: ID=%d, Login=%s", registerResp.ID, registerResp.Login)

	// Используем 201 Created и возвращаем только созданный ресурс
	gCtx.JSON(http.StatusCreated, UserInfo{
		ID:    registerResp.ID,
		Login: registerResp.Login,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user and revoke JWT token (both from Redis and session)
// @Description After logout, the token is no longer valid even if signature is correct
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(gCtx *gin.Context) {
	userID := h.auth.GetUserID(gCtx)
	userLogin := h.auth.GetUserLogin(gCtx)

	logrus.Infof("🚪 Logout request from user %d (%s)", userID, userLogin)

	// Отзываем JWT из Redis
	if userID != 0 {
		ctx, cancel := context.WithTimeout(gCtx.Request.Context(), 5*time.Second)
		defer cancel()

		if err := h.redisTokenService.RevokeToken(ctx, userID); err != nil {
			logrus.Warnf("⚠️ Failed to revoke token from Redis: %v", err)
			// Не возвращаем ошибку клиенту, продолжаем очистку сессии
		} else {
			logrus.Infof("✅ Token revoked from Redis for user %d", userID)
		}
	}

	// Очищаем сессию
	session := sessions.Default(gCtx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	if err := session.Save(); err != nil {
		logrus.Warnf("⚠️ Failed to clear session: %v", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear session",
		})
		return
	}

	logrus.Infof("✅ User %d logged out successfully", userID)

	// 204 No Content - стандартный ответ для Logout без тела ответа
	gCtx.Status(http.StatusNoContent)
}
