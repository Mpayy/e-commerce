package routes

import (
	carthttp "github.com/Mpayy/e-commerce/internal/cart/delivery/http"
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
	CartHandler     carthttp.CartHandler
	Log             *logrus.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, userHandler userhttp.UserHandler, categoryHandler producthttp.CategoryHandler, productHandler producthttp.ProductHandler, cartHandler carthttp.CartHandler, log *logrus.Logger) *Router {
	return &Router{App: app, AuthMiddleware: authMiddleware, AdminMiddleware: adminMiddleware, UserHandler: userHandler, CategoryHandler: categoryHandler, ProductHandler: productHandler, CartHandler: cartHandler, Log: log}
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

	protected.POST("/cart", r.CartHandler.AddItem)
	protected.GET("/cart", r.CartHandler.GetCart)
	protected.PATCH("/cart/:product_id", r.CartHandler.UpdateItem)
	protected.DELETE("/cart/:product_id", r.CartHandler.RemoveItem)
	protected.DELETE("/cart", r.CartHandler.ClearCart)

	adminOnly := protected.Group("/admin", r.AdminMiddleware.AdminMiddleware())
	adminOnly.POST("/categories", r.CategoryHandler.Create)
	adminOnly.POST("/products", r.ProductHandler.Create)
	adminOnly.PUT("/products/:product_id", r.ProductHandler.Update)
	adminOnly.DELETE("/products/:product_id", r.ProductHandler.Delete)
	adminOnly.PATCH("/products/:product_id/adjust-stock", r.ProductHandler.AdjustStock)
}
