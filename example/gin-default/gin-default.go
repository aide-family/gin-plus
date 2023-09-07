package main

import (
	ginplus "github.com/aide-cloud/gin-plus"
	"github.com/gin-gonic/gin"
)

func main() {
	instance := ginplus.New(gin.Default())

	instance.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "hello world",
		})
	})

	ginplus.NewCtrlC(instance).Start()
}
