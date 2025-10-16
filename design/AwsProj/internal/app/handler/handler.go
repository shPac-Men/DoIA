package handler

import (
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository     *repository.Repository
	MixingService  *service.MixingService
	ElementService *service.ElementService
}

func NewHandler(r *repository.Repository, s *service.MixingService, e *service.ElementService) *Handler {
	return &Handler{
		Repository:     r,
		MixingService:  s,
		ElementService: e,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// API v1 группа
	api := router.Group("/api/v1")
	{
		// Элементы (Elements)
		elements := api.Group("/elements")
		{
			elements.GET("", h.GetAllElements)         // GET /api/v1/elements - список элементов
			elements.GET("/:id", h.GetElementById)     // GET /api/v1/elements/:id - получить элемент
			elements.POST("", h.CreateElement)         // POST /api/v1/elements - создать элемент
			elements.PUT("/:id", h.UpdateElement)      // PUT /api/v1/elements/:id - обновить элемент
			elements.DELETE("/:id", h.DeleteElement)   // DELETE /api/v1/elements/:id - удалить элемент сделать жестким
			elements.POST("/:id/image", h.UploadImage) //POST добавление изображения.
		}

		// Корзина/Смешивание (Mixing)
		mixing := api.Group("/mixing")
		{
			mixing.GET("", h.GetMixingPage)            // GET /api/v1/mixing - получить корзину
			mixing.POST("", h.CreateMixing)            // POST /api/v1/mixing - создать смешивание
			mixing.POST("/items", h.AddToMixing)       // POST /api/v1/mixing/items - добавить в корзину
			mixing.POST("/remove", h.RemoveFromMixing) // POST /api/v1/mixing/remove - удалить из корзины
		}

		// Старые веб-роуты (можно оставить временно для обратной совместимости)
		api.GET("/chemistry", h.GetAllElements)   // старый роут для обратной совместимости
		api.GET("/element/:id", h.GetElementById) // старый роут для обратной совместимости
	}

	// HTML роуты (если еще нужны для фронтенда)
	router.GET("/mixingpage", h.GetMixingPage)    // веб-версия страницы смешивания
	router.POST("/create-mixing", h.CreateMixing) // веб-версия создания смешивания
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("../../templates/*")
	router.Static("/static", "../../resources") ///styles", "resources/styles
	router.Static("/img", "resources/img")

}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
