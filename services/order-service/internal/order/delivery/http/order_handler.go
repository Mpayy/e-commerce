package http

import (
	"github.com/gin-gonic/gin"
)

type OrderHandler interface {
	Checkout(ctx *gin.Context)
	GetHistory(ctx *gin.Context)
	GetDetail(ctx *gin.Context)
	GetSalesAnalytics(ctx *gin.Context)
	GetAdminOrderList(ctx *gin.Context)
	GetAdminOrderDetail(ctx *gin.Context)
	CancelOrder(ctx *gin.Context)
}
