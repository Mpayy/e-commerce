package producthttp

import "github.com/gin-gonic/gin"

type ProductHandler interface {
	Create(ctx *gin.Context)
}