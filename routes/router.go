package routes

import (
	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/internal/user/delivery/http"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Router struct {
	App            *gin.Engine
	AuthMiddleware *middleware.AuthMiddleware
	UserHandler    http.UserHandler
	Log            *logrus.Logger
}

func NewRouter(app *gin.Engine, authMiddleware *middleware.AuthMiddleware, userHandler http.UserHandler, log *logrus.Logger) *Router {
	return &Router{App: app, AuthMiddleware: authMiddleware, UserHandler: userHandler, Log: log}
}

func (r *Router) SetupRouter() {
	public := r.App.Group("/api/v1", middleware.Logger(r.Log))
	public.POST("/register", r.UserHandler.Register)
	public.POST("/login", r.UserHandler.Login)

	protected := r.App.Group("/api/v1", r.AuthMiddleware.AuthMiddleware())
	protected.GET("/profile", r.UserHandler.GetProfile)
	protected.DELETE("/logout", r.UserHandler.Logout)
}
