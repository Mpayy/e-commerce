package orderhttp

import (
	"github.com/gin-gonic/gin"
)

type OrderHandler interface {
	Checkout(ctx *gin.Context)
	GetHistory(ctx *gin.Context)
	GetDetail(ctx *gin.Context)
}
