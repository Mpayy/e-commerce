package http

import "github.com/gin-gonic/gin"

type UserHandler interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	GetProfile(ctx *gin.Context)
	Logout(ctx *gin.Context)
}