package main

import (
	"fmt"

	"AwsProj/internal/app/config"
	"AwsProj/internal/app/dsn"
	"AwsProj/internal/app/handler"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"AwsProj/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load("../../.env")
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

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
