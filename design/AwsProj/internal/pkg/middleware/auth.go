package middleware

// internal/app/middleware/auth.go

import (
	"AwsProj/internal/app/config"
	"AwsProj/internal/app/ds"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const jwtPrefix = "Bearer "

type AuthMiddleware struct {
	Config *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{Config: cfg}
}

// GuestAccess - позволяет доступ всем, устанавливает роль "guest" если нет токена
func (m *AuthMiddleware) GuestAccess() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")

		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			gCtx.Set("user_id", uint(0))
			gCtx.Set("role", "guest")
			gCtx.Set("is_moderator", false)
			gCtx.Next()
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]
		claims := &ds.JWTClaims{}

		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.Config.JWT.Token), nil
		})

		if err != nil || !token.Valid {
			logrus.Warnf("Invalid token in GuestAccess: %v", err)
			gCtx.Set("user_id", uint(0))
			gCtx.Set("role", "guest")
			gCtx.Set("is_moderator", false)
			gCtx.Next()
			return
		}

		gCtx.Set("user_id", claims.UserID)
		gCtx.Set("user_uuid", claims.UserUUID)
		gCtx.Set("login", claims.Login)
		gCtx.Set("role", claims.Role)
		gCtx.Set("is_moderator", claims.IsModerator)
		gCtx.Set("scopes", claims.Scopes)

		logrus.Infof("User authenticated in GuestAccess: ID=%d, Login=%s, Role=%s",
			claims.UserID, claims.Login, claims.Role)

		gCtx.Next()
	}
}

// WithAuthCheck - требует валидный JWT токен
func (m *AuthMiddleware) WithAuthCheck() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")

		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			logrus.Warn("No Bearer token in Authorization header")
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization required",
				"message": "Missing or invalid Authorization header",
			})
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]
		claims := &ds.JWTClaims{}

		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.Config.JWT.Token), nil
		})

		if err != nil || !token.Valid {
			logrus.Warnf("JWT validation failed: %v", err)
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token validation failed",
			})
			return
		}

		gCtx.Set("user_id", claims.UserID)
		gCtx.Set("user_uuid", claims.UserUUID)
		gCtx.Set("login", claims.Login)
		gCtx.Set("role", claims.Role)
		gCtx.Set("is_moderator", claims.IsModerator)
		gCtx.Set("scopes", claims.Scopes)

		logrus.Infof("User authenticated: ID=%d, Login=%s, Role=%s",
			claims.UserID, claims.Login, claims.Role)

		gCtx.Next()
	}
}

// AdminAccess - требует роль admin или is_moderator = true
func (m *AuthMiddleware) AdminAccess() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		isModerator, exists := gCtx.Get("is_moderator")

		if !exists {
			gCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "User role not found",
			})
			return
		}

		if !isModerator.(bool) {
			role, _ := gCtx.Get("role")
			logrus.Warnf("Access denied for user with role: %v", role)
			gCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "Administrator access required",
			})
			return
		}

		gCtx.Next()
	}
}

// Вспомогательные функции

func (m *AuthMiddleware) GetUserID(gCtx *gin.Context) uint {
	userID, exists := gCtx.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

func (m *AuthMiddleware) GetUserRole(gCtx *gin.Context) string {
	role, exists := gCtx.Get("role")
	if !exists {
		return "guest"
	}
	return role.(string)
}

func (m *AuthMiddleware) GetUserLogin(gCtx *gin.Context) string {
	login, exists := gCtx.Get("login")
	if !exists {
		return ""
	}
	return login.(string)
}

func (m *AuthMiddleware) IsModerator(gCtx *gin.Context) bool {
	isMod, exists := gCtx.Get("is_moderator")
	if !exists {
		return false
	}
	return isMod.(bool)
}
