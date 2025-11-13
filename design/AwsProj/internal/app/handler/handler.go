package handler

import (
	"AwsProj/internal/app/config"
	"AwsProj/internal/app/repository"
	"AwsProj/internal/app/service"
	"AwsProj/internal/pkg/middleware"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository     *repository.Repository
	MixingService  *service.MixingService
	ElementService *service.ElementService
	UserService    *service.UserService
	config         *config.Config
	auth           *middleware.AuthMiddleware // ИЗМЕНИЛИ ТИП
}

func NewHandler(
	r *repository.Repository,
	ms *service.MixingService,
	es *service.ElementService,
	us *service.UserService,
	cfg *config.Config,
	auth *middleware.AuthMiddleware, // ИЗМЕНИЛИ ТИП
) *Handler {
	return &Handler{
		Repository:     r,
		MixingService:  ms,
		ElementService: es,
		UserService:    us,
		config:         cfg,
		auth:           auth,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.Use(h.auth.GuestAccess()) // h.auth вместо h.app

	api := router.Group("/api/v1")

	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		elements := api.Group("/elements")
		{
			elements.GET("", h.GetAllElements)
			elements.GET("/:id", h.GetElementById)
		}

		api.GET("/ping", h.PingPublic)
	}

	protected := api.Group("")
	protected.Use(h.auth.WithAuthCheck()) // h.auth вместо h.app
	{
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", h.GetUserProfile)
			auth.PUT("/profile", h.UpdateUser)
			auth.POST("/logout", h.Logout)
		}

		mixing := protected.Group("/mixing")
		{
			mixing.GET("", h.GetMixingPage)
			mixing.POST("/items", h.AddToMixing)
			mixing.POST("/remove", h.RemoveFromMixing)
			mixing.GET("/cart-icon", h.GetCartIcon)
		}

		mixed := protected.Group("/mixed")
		{
			mixed.GET("/my", h.GetMyMixedList)
			mixed.GET("/my/:id", h.GetMyMixedByID)
		}
	}

	admin := api.Group("")
	admin.Use(h.auth.WithAuthCheck()) // h.auth вместо h.app
	admin.Use(h.auth.AdminAccess())   // h.auth вместо h.app
	{
		elements := admin.Group("/elements")
		{
			elements.POST("", h.CreateElement)
			elements.PUT("/:id", h.UpdateElement)
			elements.DELETE("/:id", h.DeleteElement)
			elements.POST("/:id/image", h.UploadImage)
		}

		mixed := admin.Group("/mixed")
		{
			mixed.GET("", h.GetMixedList)
			mixed.GET("/:id", h.GetMixedByID)
			mixed.POST("", h.CreateMixing)
			mixed.PUT("/:id", h.UpdateMixed)
			mixed.PUT("/:id/complete", h.CompleteMixed)
			mixed.DELETE("/:id/items", h.DeleteFromMixed)
			mixed.DELETE("/:id", h.DeleteMixed)
		}

		users := admin.Group("/users")
		{
			users.GET("", h.GetAllUsers)
			users.GET("/:id", h.GetUserByID)
			users.PUT("/:id/role", h.UpdateUserRole)
		}
	}

	h.RegisterStatic(router)
}

func (h *Handler) PingPublic(gCtx *gin.Context) {
	role := h.auth.GetUserRole(gCtx) // h.auth вместо h.app
	userID := h.auth.GetUserID(gCtx) // h.auth вместо h.app

	gCtx.JSON(200, PingResponse{
		Status:  true,
		Auth:    userID != 0,
		UserID:  userID,
		Role:    role,
		Message: "Service is working",
	})
}

func (h *Handler) Ping(gCtx *gin.Context) {
	userID := h.auth.GetUserID(gCtx)   // h.auth вместо h.app
	role := h.auth.GetUserRole(gCtx)   // h.auth вместо h.app
	login := h.auth.GetUserLogin(gCtx) // h.auth вместо h.app

	gCtx.JSON(200, PingResponse{
		Status:  true,
		Auth:    true,
		UserID:  userID,
		Role:    role,
		Message: fmt.Sprintf("Authenticated as %s (%s)", login, role),
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	//router.LoadHTMLGlob("templates/*")
	router.Static("/static", "resources/styles")
	router.Static("/img", "resources/img")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, ErrorResponse{
		Success: false,
		Error:   "Error",
		Message: err.Error(),
	})
}
