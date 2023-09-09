package ginplus

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/gin-gonic/gin"
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

func (p *People) PostCreateInfo(ctx context.Context, req struct{ Name string }) (string, error) {
	return "PostCreateInfo", errors.New("custom error")
}

func (p *People) PutUpdateInfo(ctx context.Context, req struct{ Name string }) (struct{ Name string }, error) {
	return struct{ Name string }{Name: req.Name}, nil
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
	opts := []OptionFun{
		WithMiddlewares(func(ctx *gin.Context) {
			log.Println("main middleware")
		}),
		WithBasePath("aide-cloud"),
		WithHttpMethodPrefixes(HttpMethod{
			Prefix: "Get",
			Method: Get,
		}, HttpMethod{
			Prefix: "Post",
			Method: Post,
		}, HttpMethod{
			Prefix: "Put",
			Method: Put,
		}, HttpMethod{
			Prefix: "Delete",
			Method: Delete,
		}),
		WithControllers(&People{
			Img: &Img{
				File: &File{},
			},
		}, &Slice{}),
		WithRouteNamingRuleFunc(func(methodName string) string {
			return routeToCamel(methodName)
		}),
		WithApiConfig(ApiConfig{
			Title:   "aide-cloud-api",
			Version: "v1",
		}),
	}
	ginInstance := New(r, opts...)
	ginInstance.Run(":8080")
}

type (
	MyController struct {
	}

	MyControllerReq struct {
		Name    string `form:"name"`
		Id      uint   `uri:"id"`
		Keyword string `form:"keyword"`
	}

	MyControllerResp struct {
		Name string `json:"name"`
		Id   uint   `json:"id"`
		Age  int    `json:"-"`
		Data any    `json:"data"`
	}
)

func (l *MyController) GetInfo(ctx context.Context, req MyControllerReq) (*MyControllerResp, error) {
	log.Println(req)
	return nil, nil
}

func TestGenApi(t *testing.T) {
	r := gin.Default()
	opts := []OptionFun{
		WithBasePath("aide-cloud"),
		WithControllers(&MyController{}),
	}
	ginInstance := New(r, opts...)
	ginInstance.genOpenApiYaml()
}

func TestGenApiRun(t *testing.T) {
	r := gin.Default()
	opts := []OptionFun{
		WithBasePath("aide-cloud"),
		WithControllers(&MyController{}),
	}
	ginInstance := New(r, opts...)
	ginInstance.Run()
}

type (
	MiddController struct {
		ChildMiddController *ChildMiddController
	}

	ChildMiddController struct {
	}
)

func (l *ChildMiddController) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Println("ChildMiddController")
		},
	}
}

func (l *ChildMiddController) Info() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Info action")
	}
}

type DetailxReq struct {
	Id uint `uri:"id"`
}

type DetailxResp struct {
	Name string `json:"name"`
	Id   uint   `json:"id"`
}

func (l *ChildMiddController) Detile(ctx context.Context, req *DetailxReq) (*DetailxResp, error) {
	log.Println("Detile")
	return &DetailxResp{Name: "aide-cloud", Id: req.Id}, nil
}

func (l *MiddController) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Println("MiddController")
		},
	}
}

func (l *MiddController) GetParent() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Parent action")
	}
}

var _ Middlewarer = (*MiddController)(nil)
var _ Middlewarer = (*ChildMiddController)(nil)

func TestGenApiRunMidd(t *testing.T) {
	r := gin.Default()
	opts := []OptionFun{
		WithBasePath("aide-cloud"),
		WithControllers(&MiddController{
			ChildMiddController: &ChildMiddController{},
		}),
	}
	ginInstance := New(r, opts...)
	ginInstance.Run()
}

type (
	Path1 struct {
		Path2 *Path2
	}
	Path2 struct {
	}
)

func (p *Path2) MethoderMiddlewares() map[string][]gin.HandlerFunc {
	return map[string][]gin.HandlerFunc{
		"GetInfoByID": {
			func(ctx *gin.Context) {
				log.Println("Path2 GetInfoByID middleware1")
			},
			func(ctx *gin.Context) {
				log.Println("Path2 GetInfoByID middleware2")
			},
		},
	}
}

func (p *Path2) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Println("Path2 Middlewares 1")
		},
		func(ctx *gin.Context) {
			log.Println("Path2 Middlewares 2")
		},
	}
}

func (p *Path1) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Println("Path1 Middlewares 1")
		},
		func(ctx *gin.Context) {
			log.Println("Path1 Middlewares 2")
		},
	}
}

var _ Middlewarer = (*Path1)(nil)
var _ Middlewarer = (*Path2)(nil)
var _ MethoderMiddlewarer = (*Path2)(nil)

func (p *Path1) GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Path1 GetInfo")
	}
}

func (p *Path1) GetInfoByID(ctx context.Context, req *struct {
	Id uint `uri:"id"`
}) (*struct {
	Id uint `json:"id"`
}, error) {
	log.Println("Path1 GetInfoByID")
	return (*struct {
		Id uint `json:"id"`
	})(&struct{ Id uint }{Id: req.Id}), nil
}

func (p *Path2) GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Path2 GetInfo")
	}
}

func (p *Path2) GetInfoByID(ctx context.Context, req *struct {
	Id uint `uri:"id"`
}) (*struct {
	Id uint `json:"id"`
}, error) {
	log.Println("Path2 GetInfoByID")
	return (*struct {
		Id uint `json:"id"`
	})(&struct{ Id uint }{Id: req.Id}), nil
}

func TestRouteGroup(t *testing.T) {
	New(gin.Default(), WithControllers(&Path1{Path2: &Path2{}}))
}
