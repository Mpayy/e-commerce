package carthttp

import "github.com/gin-gonic/gin"

type CartHandler interface {
	AddItem(ctx *gin.Context)
}
