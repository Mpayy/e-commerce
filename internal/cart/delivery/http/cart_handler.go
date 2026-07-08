package carthttp

import "github.com/gin-gonic/gin"

type CartHandler interface {
	AddItem(ctx *gin.Context)
	UpdateItem(ctx *gin.Context)
	RemoveItem(ctx *gin.Context)
	GetCart(ctx *gin.Context)
}
