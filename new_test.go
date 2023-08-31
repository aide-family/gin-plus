package ginplus

import (
	"github.com/gin-gonic/gin"
	"log"
	"testing"
)

type People struct {
	*Img
}

type Img struct {
	File *File
}

type File struct {
}

var _ Middlewarer = (*People)(nil)
var _ Controller = (*People)(nil)
var _ MethoderMiddlewarer = (*People)(nil)

func (l *Img) GetImgInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "GetInfo")
	}
}

func (l *File) GetFiles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "GetFiles")
	}
}

func (p *People) GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "GetInfo")
	}
}

func (p *People) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "List")
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

func (p *People) BasePath() string {
	return "/people/v1"
}

func (p *People) MethoderMiddlewares() map[string][]gin.HandlerFunc {
	return map[string][]gin.HandlerFunc{
		"GetInfo": {
			func(ctx *gin.Context) {
				log.Println("GetInfo middleware1")
			},
		},
	}
}

type Slice []string

var _ Middlewarer = (*Slice)(nil)

func (l *Slice) Middlewares() []gin.HandlerFunc {
	return nil
}

func (l *Slice) GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "GetInfo")
	}
}

func TestNew(t *testing.T) {
	r := gin.Default()
	opts := []Option{
		WithMiddlewares(func(ctx *gin.Context) {
			log.Println("main middleware")
		}),
		WithBasePath("aide-cloud"),
		WithHttpMethodPrefixes(Get, Post),
		WithControllers(&People{
			Img: &Img{
				File: &File{},
			},
		}, &Slice{}),
		WithDefaultHttpMethod(Post),
		WithRouteNamingRuleFunc(func(methodName string) string {
			return routeToCamel(methodName)
		}),
	}
	ginInstance := New(r, opts...)
	ginInstance.Run(":8080")
}
