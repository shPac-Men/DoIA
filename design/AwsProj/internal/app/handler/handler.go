// package handler

// import (
// 	"AwsProj/internal/app/repository"
// 	"AwsProj/internal/app/service"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// )

// type Handler struct {
// 	Repository     *repository.Repository
// 	MixingService  *service.MixingService
// 	ElementService *service.ElementService
// 	UserService    *service.UserService
// }

// func NewHandler(r *repository.Repository, s *service.MixingService, e *service.ElementService, u *service.UserService) *Handler {
// 	return &Handler{
// 		Repository:     r,
// 		MixingService:  s,
// 		ElementService: e,
// 		UserService:    u,
// 	}
// }

// func (h *Handler) RegisterHandler(router *gin.Engine) {
// 	// API v1 группа
// 	userHandler := NewUserHandler(h.UserService)
// 	api := router.Group("/api/v1")
// 	{
// 		// Элементы (Elements)
// 		elements := api.Group("/elements")
// 		{
// 			elements.GET("", h.GetAllElements)         // GET /api/v1/elements - список элементов
// 			elements.GET("/:id", h.GetElementById)     // GET /api/v1/elements/:id - получить элемент
// 			elements.POST("", h.CreateElement)         // POST /api/v1/elements - создать элемент
// 			elements.PUT("/:id", h.UpdateElement)      // PUT /api/v1/elements/:id - обновить элемент
// 			elements.DELETE("/:id", h.DeleteElement)   // DELETE /api/v1/elements/:id - удалить элемент сделать жестким
// 			elements.POST("/:id/image", h.UploadImage) //POST добавление изображения.
// 		}

// 		// Корзина/Смешивание (Mixing)
// 		mixing := api.Group("/mixing")
// 		{
// 			mixing.GET("", h.GetMixingPage)            // GET /api/v1/mixing - получить корзину
// 			mixing.POST("", h.CreateMixing)            // POST /api/v1/mixing - создать смешивание
// 			mixing.POST("/items", h.AddToMixing)       // POST /api/v1/mixing/items - добавить в корзину
// 			mixing.POST("/remove", h.RemoveFromMixing) // POST /api/v1/mixing/remove - удалить из корзины
// 			mixing.GET("/cart-icon", h.GetCartIcon)    //GET иконки корзины
// 			//mixing.GET("/all", h.GetMixedList)         // GET список с фильтрацией по диапазону даты формирования и статусу
// 		}
// 		mixed := api.Group("/mixed")
// 		{
// 			mixed.GET("", h.GetMixedList) // GET список с фильтрацией по диапазону даты формирования и статусу
// 			mixed.GET("/:id", h.GetMixedByID)
// 			mixed.PUT("/:id", h.UpdateMixed)              // PUT /api/v1/mixed/:id - обновление заявки
// 			mixed.PUT("/:id/complete", h.CompleteMixed)   // PUT /api/v1/mixed/:id/complete - формирование заявки
// 			mixed.DELETE("/:id", h.DeleteMixed)           // DELETE /api/v1/mixed/:id - удаление заявки
// 			mixed.DELETE("/:id/items", h.DeleteFromMixed) // DELETE /api/v1/mixed/:id/items - удаление элемента из заявки
// 		}
// 		auth := api.Group("/auth")
// 		{
// 			auth.POST("/register", userHandler.Register)     // POST /api/v1/auth/register - регистрация
// 			auth.GET("/profile", userHandler.GetUserProfile) // GET /api/v1/auth/profile - данные пользователя
// 			auth.PUT("/profile", userHandler.UpdateUser)     // PUT /api/v1/auth/profile - обновление пользователя
// 			auth.POST("/login", userHandler.Login)           // POST /api/v1/auth/login - аутентификация
// 			auth.POST("/logout", userHandler.Logout)
// 		}
// 		users := api.Group("/users")
// 		{
// 			users.GET("/:id", userHandler.GetUserByID) // GET /api/v1/users/:id - данные любого пользователя по ID
// 		}
// 		// Старые веб-роуты (можно оставить временно для обратной совместимости)
// 		api.GET("/chemistry", h.GetAllElements)   // старый роут для обратной совместимости
// 		api.GET("/element/:id", h.GetElementById) // старый роут для обратной совместимости

// 	}

// 	// HTML роуты (если еще нужны для фронтенда)
// 	router.GET("/mixingpage", h.GetMixingPage)    // веб-версия страницы смешивания
// 	router.POST("/create-mixing", h.CreateMixing) // веб-версия создания смешивания
// }

// // RegisterStatic То же самое, что и с маршрутами, регистрируем статику
// func (h *Handler) RegisterStatic(router *gin.Engine) {
// 	router.LoadHTMLGlob("../../templates/*")
// 	router.Static("/static", "../../resources") ///styles", "resources/styles
// 	router.Static("/img", "resources/img")

// }

// // errorHandler для более удобного вывода ошибок
// func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
// 	logrus.Error(err.Error())
// 	ctx.JSON(errorStatusCode, gin.H{
// 		"status":      "error",
// 		"description": err.Error(),
// 	})
// }

// func (h *Handler) Ping(gCtx *gin.Context) {
// 	// Теперь этот эндпоинт защищен middleware
// 	gCtx.JSON(http.StatusOK, gin.H{
// 		"auth":   true,
// 		"status": true,
// 	})
// }

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
