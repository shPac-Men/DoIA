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

// @host localhost:8082
// @schemes http
// @BasePath /api/v1
func main() {
	_ = godotenv.Load("../../.env")
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Загружаем конфиг ПЕРВЫМ делом
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Маршрут для сваггера
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// СОЗДАЕМ СЕРВИСЫ
	mixingService := service.NewMixingService(rep)
	elementService, err := service.NewElementService(
		rep,
		"localhost:9000", // MinIO endpoint
		"admin",          // MinIO access key
		"admin123456",    // MinIO secret key
		"staticimages",   // bucket name
	)
	if err != nil {
		logrus.Fatalf("error initializing element service: %v", err)
	}

	userService := service.NewUserService(rep)

	// ПЕРЕДАЕМ КОНФИГ В HANDLER (добавляем conf в параметры)
	hand := handler.NewHandler(rep, mixingService, elementService, userService, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
