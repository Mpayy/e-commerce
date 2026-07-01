package producthttp

import "github.com/gin-gonic/gin"

type CategoryHandler interface {
	Create(ctx *gin.Context)
	GetAll(ctx *gin.Context)
}