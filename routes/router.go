package routes

import (
	carthttp "github.com/Mpayy/e-commerce/internal/cart/delivery/http"
	"github.com/Mpayy/e-commerce/internal/middleware"
	orderhttp "github.com/Mpayy/e-commerce/internal/order/delivery/http"
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
	OrderHandler    orderhttp.OrderHandler
	CartHandler     carthttp.CartHandler
	Log             *logrus.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, userHandler userhttp.UserHandler, categoryHandler producthttp.CategoryHandler, productHandler producthttp.ProductHandler, orderHandler orderhttp.OrderHandler, cartHandler carthttp.CartHandler, log *logrus.Logger) *Router {
	return &Router{
		App:             app,
		AuthMiddleware:  authMiddleware,
		AdminMiddleware: adminMiddleware,
		UserHandler:     userHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		CartHandler:     cartHandler,
		Log:             log,
	}
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

	cart := protected.Group("/cart")
	cart.POST("", r.CartHandler.AddItem)
	cart.GET("", r.CartHandler.GetCart)
	cart.PATCH("/:product_id", r.CartHandler.UpdateItem)
	cart.DELETE("/:product_id", r.CartHandler.RemoveItem)
	cart.DELETE("", r.CartHandler.ClearCart)

	order := protected.Group("/orders")
	order.POST("", r.OrderHandler.Checkout)
	order.GET("", r.OrderHandler.GetHistory)
	order.GET("/:order_id", r.OrderHandler.GetDetail)

	adminOnly := protected.Group("/admin", r.AdminMiddleware.AdminMiddleware())
	adminOnly.POST("/categories", r.CategoryHandler.Create)
	adminOnly.POST("/products", r.ProductHandler.Create)
	adminOnly.PUT("/products/:product_id", r.ProductHandler.Update)
	adminOnly.DELETE("/products/:product_id", r.ProductHandler.Delete)
	adminOnly.PATCH("/products/:product_id/adjust-stock", r.ProductHandler.AdjustStock)
}
