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

	// CORS Middleware - ПЕРВЫЙ
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

	// Redis Client for Tokens
	redisClient, err := pkg.NewRedisClientForTokens(redisConfig)
	if err != nil {
		logrus.Fatalf("❌ Redis client error: %v", err)
	}
	logrus.Info("✅ Redis client initialized")

	// Redis Token Service
	redisTokenService := pkg.NewRedisTokenService(redisClient)
	logrus.Info("✅ RedisTokenService initialized")

	// Redis Session Store
	store, err := pkg.NewRedisSessionStore(redisConfig)
	if err != nil {
		logrus.Fatalf("❌ Session store error: %v", err)
	}
	logrus.Info("✅ Session store initialized")

	// ⭐ Redis Session Middleware - ВТОРОЙ (ДО ВСЕХ ХЕНДЛЕРОВ)
	router.Use(middleware.RedisSessionMiddleware(store))
	logrus.Info("✅ Session middleware registered")

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
	authMiddleware := middleware.NewAuthMiddleware(conf, redisTokenService)
	logrus.Info("✅ AuthMiddleware initialized")

	// Handler
	hand := handler.NewHandler(
		rep,
		mixingService,
		elementService,
		userService,
		conf,
		authMiddleware,
		redisTokenService,
	)
	logrus.Info("✅ Handler initialized")

	// Register Routes
	hand.RegisterHandler(router)
	logrus.Info("✅ Routes registered")

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logrus.Info("✅ Swagger at /swagger/index.html")

	// Построение адреса сервера из конфига
	serverAddr := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)

	// Run
	logrus.Infof("🚀 Starting server on %s...", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		logrus.Fatalf("❌ Server error: %v", err)
	}
}
