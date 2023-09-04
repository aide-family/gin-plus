package main

import (
	"github.com/gin-gonic/gin"

	ginplus "github.com/aide-cloud/gin-plus"
)

func main() {
	instance := ginplus.New(gin.Default(),
		ginplus.WithAddr(":8080"),
		ginplus.WithGenApiEnable(false),
		ginplus.WithGraphqlConfig(ginplus.GraphqlConfig{
			Enable:     true,
			HandlePath: "/graphql",
			ViewPath:   "/graphql",
			Root:       &Root{},
			Content:    content,
		}))

	ginplus.NewCtrlC(instance).Start()
}
