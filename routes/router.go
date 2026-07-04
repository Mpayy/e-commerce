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
	ProductHandler  producthttp.ProductHandler
	Log             *logrus.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, userHandler userhttp.UserHandler, categoryHandler producthttp.CategoryHandler, productHandler producthttp.ProductHandler, log *logrus.Logger) *Router {
	return &Router{App: app, AuthMiddleware: authMiddleware, AdminMiddleware: adminMiddleware, UserHandler: userHandler, CategoryHandler: categoryHandler, ProductHandler: productHandler, Log: log}
}

func (r *Router) SetupRouter() {
	r.App.Use(middleware.Logger(r.Log))

	public := r.App.Group("/api/v1")
	public.POST("/register", r.UserHandler.Register)
	public.POST("/login", r.UserHandler.Login)
	public.GET("/categories", r.CategoryHandler.GetAll)
	public.GET("/products/:product_id", r.ProductHandler.GetByID)
	public.GET("/products", r.ProductHandler.Search)

	protected := r.App.Group("/api/v1")
	protected.Use(r.AuthMiddleware.AuthMiddleware())

	protected.GET("/profile", r.UserHandler.GetProfile)
	protected.DELETE("/logout", r.UserHandler.Logout)

	adminOnly := protected.Group("/admin", r.AdminMiddleware.AdminMiddleware())
	adminOnly.POST("/categories", r.CategoryHandler.Create)
	adminOnly.POST("/products", r.ProductHandler.Create)
	adminOnly.PUT("/products/:product_id", r.ProductHandler.Update)
	adminOnly.DELETE("/products/:product_id", r.ProductHandler.Delete)
	adminOnly.PATCH("/products/:product_id/adjust-stock", r.ProductHandler.AdjustStock)
}
