package main

import (
	"log"

	ginplus "github.com/aide-cloud/gin-plus"
	"github.com/gin-gonic/gin"
)

var _ ginplus.Controller = (*Api)(nil)
var _ ginplus.Middlewarer = (*Api)(nil)

type Api struct {
}

func (a *Api) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Println("middleware 1")
		},
	}
}

func (a *Api) BasePath() string {
	return "/api/v1"
}

func (a *Api) GetA() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "[Get] hello world")
	}
}

func NewApi() *Api {
	return &Api{}
}

func main() {
	instance := ginplus.New(gin.Default(), ginplus.WithControllers(NewApi()))

	ginplus.NewCtrlC(instance).Start()
}
