package pkg

import (
	"fmt"

	"AwsProj/internal/app/config"

	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler interface {
		RegisterHandler(*gin.Engine)
		RegisterStatic(*gin.Engine)
	}
	SessionStore redis.Store
}

func NewApp(
	c *config.Config,
	r *gin.Engine,
	h interface {
		RegisterHandler(*gin.Engine)
		RegisterStatic(*gin.Engine)
	},
	store redis.Store,
) *Application {
	return &Application{
		Config:       c,
		Router:       r,
		Handler:      h,
		SessionStore: store,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")
	a.Handler.RegisterHandler(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
