package main

import (
	"fmt"

	"AwsProj/internal/app/config"
	"AwsProj/internal/app/dsn"
	"AwsProj/internal/app/handler"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"AwsProj/internal/pkg"

	_ "AwsProj/docs" // docs генерируется Swag

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform

// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru

// @license.name AS IS (NO WARRANTY)

// @host http://localhost:8082
// @schemes https http
// @BasePath /api/v1

func main() {
	_ = godotenv.Load("../../.env")
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}
	//маршрут для сваги
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// СОЗДАЕМ MIXING SERVICE
	mixingService := service.NewMixingService(rep)
	elementService, err := service.NewElementService(
		rep,
		"localhost:9000", // MinIO endpoint
		"admin",          // MinIO access key
		"admin123456",    // MinIO secret key
		"staticimages",   // bucket name
	)
	userService := service.NewUserService(rep)
	if err != nil {
		logrus.Fatalf("error initializing element service: %v", err)
	}
	// ПЕРЕДАЕМ И REP И MIXING SERVICE
	hand := handler.NewHandler(rep, mixingService, elementService, userService)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
