package routes

import (
	_ "github.com/Mpayy/e-commerce/monolith/docs"
	carthttp "github.com/Mpayy/e-commerce/monolith/internal/cart/delivery/http"
	orderhttp "github.com/Mpayy/e-commerce/monolith/internal/order/delivery/http"
	userhttp "github.com/Mpayy/e-commerce/monolith/internal/user/delivery/http"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	App            *gin.Engine
	AuthMiddleware *middleware.AuthMiddleware
	UserHandler    userhttp.UserHandler
	OrderHandler   orderhttp.OrderHandler
	CartHandler    carthttp.CartHandler
	Log            *logger.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, userHandler userhttp.UserHandler, orderHandler orderhttp.OrderHandler, cartHandler carthttp.CartHandler, log *logger.Logger) *Router {
	return &Router{
		App:            app,
		AuthMiddleware: authMiddleware,
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		CartHandler:    cartHandler,
		Log:            log,
	}
}

func (r *Router) SetupRouter() *gin.Engine {
	r.App.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := r.App.Group("/api/v1")
	public.POST("/register", r.UserHandler.Register)
	public.POST("/login", r.UserHandler.Login)

	protected := r.App.Group("/api/v1")
	protected.Use(r.AuthMiddleware.RequireAuth())

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

	return r.App
}
