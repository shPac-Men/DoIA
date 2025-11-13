package middleware

import (
	"github.com/gin-contrib/sessions"

	"github.com/gin-gonic/gin"
)

func RedisSessionMiddleware(store sessions.Store) gin.HandlerFunc {
	return sessions.Sessions("session_id", store)
}

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		session := sessions.Default(gCtx)
		userID := session.Get("user_id")

		if userID == nil {
			gCtx.Next()
			return
		}

		gCtx.Set("user_id", userID)
		gCtx.Set("login", session.Get("login"))
		gCtx.Set("role", session.Get("role"))
		gCtx.Set("is_moderator", session.Get("is_moderator"))

		gCtx.Next()
	}
}
