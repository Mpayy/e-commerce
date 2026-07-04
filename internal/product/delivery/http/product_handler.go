package producthttp

import "github.com/gin-gonic/gin"

type ProductHandler interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	AdjustStock(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	Search(ctx *gin.Context)
}