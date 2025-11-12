package handler

import (
	"AwsProj/internal/app/config"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository     *repository.Repository
	MixingService  *service.MixingService
	ElementService *service.ElementService
	UserService    *service.UserService
	config         *config.Config // Добавляем конфиг
}

func NewHandler(
	r *repository.Repository,
	s *service.MixingService,
	e *service.ElementService,
	u *service.UserService,
	cfg *config.Config, // Добавляем конфиг в конструктор
) *Handler {
	return &Handler{
		Repository:     r,
		MixingService:  s,
		ElementService: e,
		UserService:    u,
		config:         cfg, // Инициализируем конфиг
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Публичные маршруты (без авторизации)
	public := router.Group("/api/v1")
	{
		auth := public.Group("/auth")
		{
			auth.POST("/login", h.Login)       // Логин - публичный
			auth.POST("/register", h.Register) // Регистрация - публичная
		}

	}

	// Защищенные маршруты (требуют JWT)
	protected := router.Group("/api/v1")
	protected.Use(h.WithAuthCheck()) // Применяем middleware ко всей группе
	{
		// Элементы (Elements)
		elements := protected.Group("/elements")
		{
			elements.GET("", h.GetAllElements)
			elements.GET("/:id", h.GetElementById)
			elements.POST("", h.CreateElement)
			elements.PUT("/:id", h.UpdateElement)
			elements.DELETE("/:id", h.DeleteElement)
			elements.POST("/:id/image", h.UploadImage)
		}

		// Корзина/Смешивание (Mixing)
		mixing := protected.Group("/mixing")
		{
			mixing.GET("", h.GetMixingPage)
			mixing.POST("", h.CreateMixing)
			mixing.POST("/items", h.AddToMixing)
			mixing.POST("/remove", h.RemoveFromMixing)
			mixing.GET("/cart-icon", h.GetCartIcon)
		}

		mixed := protected.Group("/mixed")
		{
			mixed.GET("", h.GetMixedList)
			mixed.GET("/:id", h.GetMixedByID)
			mixed.PUT("/:id", h.UpdateMixed)
			mixed.PUT("/:id/complete", h.CompleteMixed)
			mixed.DELETE("/:id", h.DeleteMixed)
			mixed.DELETE("/:id/items", h.DeleteFromMixed)
		}

		// Профиль пользователя (только для авторизованных)
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", h.GetUserProfile)
			auth.PUT("/profile", h.UpdateUser)
			auth.POST("/logout", h.Logout)
		}

		users := protected.Group("/users")
		{
			users.GET("/:id", h.GetUserByID)
		}

		// Ping endpoint (теперь защищенный)
		protected.GET("/ping", h.Ping)

		// Старые роуты для обратной совместимости
		protected.GET("/chemistry", h.GetAllElements)
		protected.GET("/element/:id", h.GetElementById)
	}

	// HTML роуты (можно оставить публичными или тоже защитить)
	router.GET("/mixingpage", h.GetMixingPage)
	router.POST("/create-mixing", h.CreateMixing)
}

// RegisterStatic регистрирует статические файлы
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("../../templates/*")
	router.Static("/static", "../../resources")
	router.Static("/img", "resources/img")
}

// errorHandler для обработки ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

// Ping godoc
// @Summary Ping endpoint
// @Description Check if service is working and user is authenticated
// @Tags utils
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PingResponse
// @Failure 403 {object} ErrorResponse
// @Router /ping [get]
func (h *Handler) Ping(gCtx *gin.Context) {
	gCtx.JSON(http.StatusOK, PingResponse{
		Auth:    true,
		Status:  true,
		Message: "Service is working and user is authenticated",
	})
}
