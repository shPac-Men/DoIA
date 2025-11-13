package pkg

import (
	"AwsProj/internal/app/config"
	"fmt"

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
}

func NewApp(c *config.Config, r *gin.Engine, h interface {
	RegisterHandler(*gin.Engine)
	RegisterStatic(*gin.Engine)
}) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	// УДАЛЕНО: a.Handler.RegisterStatic(a.Router)
	// RegisterStatic вызывается внутри RegisterHandler

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
