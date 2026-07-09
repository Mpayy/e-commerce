package orderhttp

import (
	"github.com/gin-gonic/gin"
)

type OrderHandler interface {
	Checkout(ctx *gin.Context)
}
