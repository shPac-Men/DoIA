package pkg

import (
	"AwsProj/internal/app/ds"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

func (a *Application) WithAuthCheck(gCtx *gin.Context) {
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		gCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.Config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithStatus(http.StatusForbidden)
		log.Println(err)
		return
	}

	gCtx.Next()
}
