# gin plus

> 用户对gin框架增强, 实现通过根据方法名称实现路由注册

## 安装

```shell
go get -u github.com/aide-cloud/gin-plus
```

## 使用

```go
package main

import (
	"log"

	"github.com/aide-cloud/gin-plus"

	"github.com/gin-gonic/gin"
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

func main() {
	r := gin.Default()
	ginInstance := New(r, WithControllers(&People{}))
	ginInstance.Run(":8080")
}

```

