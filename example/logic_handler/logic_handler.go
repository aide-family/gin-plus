package main

import (
	"context"

	ginplus "github.com/aide-cloud/gin-plus"
	"github.com/gin-gonic/gin"
)

type (
	Api struct {
		V1 *V1
		V2 *V2
	}

	V1 struct{}

	V2 struct{}

	Req struct {
		Id   uint   `uri:"id"`
		Name string `form:"name"`
		Data []any  `json:"data"`
	}
	Resp struct {
		Id   uint   `json:"id"`
		Name string `json:"name"`
	}
)

func (a *Api) GetInfo(ctx context.Context, req *Req) (*Resp, error) {
	return &Resp{
		Id:   req.Id,
		Name: "gin-plus" + req.Name,
	}, nil
}

func (l *V1) PostInfo(ctx context.Context, req *Req) (*Resp, error) {
	return &Resp{
		Id:   req.Id,
		Name: "gin-plus" + req.Name,
	}, nil
}

func (l *V2) PutInfo(ctx context.Context, req *Req) (*Resp, error) {
	return &Resp{
		Id:   req.Id,
		Name: "gin-plus" + req.Name,
	}, nil
}

func main() {
	instance := ginplus.New(gin.Default(), ginplus.WithControllers(&Api{
		V1: &V1{},
		V2: &V2{},
	}))

	// http://localhost:8080/api/info/12?name=xx
	ginplus.NewCtrlC(instance).Start()
}
