package handler

import (
	"AwsProj/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/chemistry", h.GetAllElements)
	router.GET("/element/:id", h.GetElementById)
	router.POST("/element/delete", h.DeleteElement) //нахуй
	//router.GET("/mixingpage", h.CalculatePage) потом реализовать
	router.POST("/element/add-to-cart", h.AddToCart)
	router.GET("/mixingpage", h.GetCartPage)
	router.POST("/create-request", h.CreateRequest)
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
