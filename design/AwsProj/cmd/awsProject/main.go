package main

import (
	"fmt"

	"AwsProj/internal/app/config"
	"AwsProj/internal/app/dsn"
	"AwsProj/internal/app/handler"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"AwsProj/internal/pkg"
	"AwsProj/internal/pkg/middleware" // ДОБАВИЛИ

	_ "AwsProj/docs"

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
	// Загрузка .env файла
	_ = godotenv.Load("../../.env")

	// Создание роутера
	router := gin.Default()

	// CORS middleware
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

	// 1. Загружаем конфиг ПЕРВЫМ делом
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("❌ Error loading config: %v", err)
	}
	logrus.Info("✅ Config loaded successfully")

	// 2. Инициализируем подключение к БД
	postgresString := dsn.FromEnv()
	fmt.Println("📦 Database connection string:", postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("❌ Error initializing repository: %v", errRep)
	}
	logrus.Info("✅ Repository initialized")

	// 3. Создаем сервисы (dependency injection снизу-вверх)

	// UserService - требует repo и config для JWT генерации
	userService := service.NewUserService(rep, conf)
	logrus.Info("✅ UserService initialized")

	// MixingService - требует только repo
	mixingService := service.NewMixingService(rep)
	logrus.Info("✅ MixingService initialized")

	// ElementService - требует repo и MinIO параметры
	elementService, err := service.NewElementService(
		rep,
		"localhost:9000", // MinIO endpoint
		"admin",          // MinIO access key
		"admin123456",    // MinIO secret key
		"staticimages",   // bucket name
	)
	if err != nil {
		logrus.Fatalf("❌ Error initializing element service: %v", err)
	}
	logrus.Info("✅ ElementService initialized")

	// 4. Создаем AuthMiddleware (вместо application для middleware)
	authMiddleware := middleware.NewAuthMiddleware(conf)
	logrus.Info("✅ AuthMiddleware initialized")

	// 5. Создаем Handler с всеми зависимостями
	hand := handler.NewHandler(rep, mixingService, elementService, userService, conf, authMiddleware)
	logrus.Info("✅ Handler initialized")

	// 6. Создаем Application с handler
	application := pkg.NewApp(conf, router, hand)
	logrus.Info("✅ Application created")

	// 7. Регистрируем маршрут для Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logrus.Info("✅ Swagger UI available at /swagger/index.html")

	// 8. Запускаем приложение
	logrus.Info("🚀 Starting application...")
	application.RunApp()
}
