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

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/chemistry", h.GetAllElements)
	router.GET("/element/:id", h.GetElementById)
	//router.POST("/element/delete", h.DeleteElement) //удалить потом ибо не нужно, но проверить
	router.POST("/element/add-to-mixing", h.AddToMixing)
	router.GET("/mixingpage", h.GetMixingPage)
	router.POST("/create-mixing", h.CreateMixing) //обединить с 4
	// В вашем router setup
	router.POST("/remove-from-mixing", h.RemoveFromMixing)

	router.POST("/elements", h.CreateElement) // add element
	router.PUT("/elements/:id", h.UpdateElement)
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
