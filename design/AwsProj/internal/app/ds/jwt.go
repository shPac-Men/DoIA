package ds

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserID      uint      `json:"user_id"`      // ID из БД
	UserUUID    uuid.UUID `json:"user_uuid"`    // UUID для дополнительной безопасности
	Login       string    `json:"login"`        // Логин пользователя
	Role        string    `json:"role"`         // "guest", "visitor", "admin"
	IsModerator bool      `json:"is_moderator"` // Флаг модератора из БД
	Scopes      []string  `json:"scopes"`       // Дополнительные права
}
