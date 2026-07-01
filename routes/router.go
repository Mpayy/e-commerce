package routes

import (
	"github.com/Mpayy/e-commerce/internal/middleware"
	producthttp "github.com/Mpayy/e-commerce/internal/product/delivery/http"
	userhttp "github.com/Mpayy/e-commerce/internal/user/delivery/http"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Router struct {
	App             *gin.Engine
	AuthMiddleware  *middleware.AuthMiddleware
	AdminMiddleware *middleware.AdminMiddleware
	UserHandler     userhttp.UserHandler
	CategoryHandler producthttp.CategoryHandler
	Log             *logrus.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, userHandler userhttp.UserHandler, categoryHandler producthttp.CategoryHandler, log *logrus.Logger) *Router {
	return &Router{App: app, AuthMiddleware: authMiddleware, AdminMiddleware: adminMiddleware, UserHandler: userHandler, CategoryHandler: categoryHandler, Log: log}
}

func (r *Router) SetupRouter() {
	r.App.Use(middleware.Logger(r.Log))

	public := r.App.Group("/api/v1")
	public.POST("/register", r.UserHandler.Register)
	public.POST("/login", r.UserHandler.Login)

	protected := r.App.Group("/api/v1")
	protected.Use(r.AuthMiddleware.AuthMiddleware())

	protected.GET("/profile", r.UserHandler.GetProfile)
	protected.DELETE("/logout", r.UserHandler.Logout)

	categoryAdmin := protected.Group("/categories", r.AdminMiddleware.AdminMiddleware())
	categoryAdmin.POST("", r.CategoryHandler.Create)
}
