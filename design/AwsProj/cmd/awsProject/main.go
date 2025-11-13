package main

import (
	"fmt"

	"AwsProj/internal/app/config"
	"AwsProj/internal/app/dsn"
	"AwsProj/internal/app/handler"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"AwsProj/internal/pkg"
	"AwsProj/internal/pkg/middleware"

	_ "AwsProj/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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

	// Config
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("❌ Config error: %v", err)
	}
	logrus.Info("✅ Config loaded")

	// Database
	postgresString := dsn.FromEnv()
	fmt.Println("📦 Database:", postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("❌ Repository error: %v", errRep)
	}
	logrus.Info("✅ Repository initialized")

	// Redis Config
	redisConfig := config.NewRedisConfig()
	logrus.Infof("✅ Redis config: %s:%s", redisConfig.Host, redisConfig.Port)

	logrus.Infof("DEBUG Redis Password: '%s' (length: %d)",
		redisConfig.Password,
		len(redisConfig.Password))
	logrus.Infof("DEBUG Redis SecretKey: '%s'", redisConfig.SecretKey)

	// Redis Session Store
	store, err := pkg.NewRedisSessionStore(redisConfig)
	if err != nil {
		logrus.Fatalf("❌ Redis error: %v", err)
	}
	logrus.Info("✅ Redis session store initialized")

	// Services
	userService := service.NewUserService(rep, conf)
	logrus.Info("✅ UserService initialized")

	mixingService := service.NewMixingService(rep)
	logrus.Info("✅ MixingService initialized")

	elementService, err := service.NewElementService(
		rep,
		"localhost:9000",
		"admin",
		"admin123456",
		"staticimages",
	)
	if err != nil {
		logrus.Fatalf("❌ ElementService error: %v", err)
	}
	logrus.Info("✅ ElementService initialized")

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(conf)
	logrus.Info("✅ AuthMiddleware initialized")

	// Handler
	hand := handler.NewHandler(rep, mixingService, elementService, userService, conf, authMiddleware)
	logrus.Info("✅ Handler initialized")

	// Redis session middleware
	router.Use(middleware.RedisSessionMiddleware(store))
	logrus.Info("✅ Redis session middleware registered")

	// Application
	application := pkg.NewApp(conf, router, hand, store)
	logrus.Info("✅ Application created")

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logrus.Info("✅ Swagger at /swagger/index.html")

	// Run
	logrus.Info("🚀 Starting server...")
	application.RunApp()
}
