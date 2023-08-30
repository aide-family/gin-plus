package ginplus

import (
	"github.com/gin-gonic/gin"
	"log"
	"testing"
)

type People struct {
}

func (p *People) GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "GetInfo")
	}
}

func (p *People) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(context *gin.Context) {
			log.Println("middleware1")
		},
		func(context *gin.Context) {
			log.Println("middleware2")
		},
	}
}

func TestNew(t *testing.T) {
	r := gin.Default()
	ginInstance := New(r, WithControllers(&People{}))
	ginInstance.Run(":8080")
}
