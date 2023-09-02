package ginplus

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aide-cloud/gin-plus/swagger"
	"github.com/gin-gonic/gin"
)

type (
	GinEngine struct {
		*gin.Engine
		middlewares        []gin.HandlerFunc
		controllers        []any
		httpMethodPrefixes map[string]httpMethod
		basePath           string
		defaultHttpMethod  httpMethod
		// 自定义路由命名规则函数
		routeNamingRuleFunc func(methodName string) string

		// 文档配置
		apiConfig          ApiConfig
		defaultOpenApiYaml string
		apiRoutes          map[string][]ApiRoute

		// 生成API路由开关, 默认为true
		genApiEnable bool
	}

	ApiConfig Info

	RouteNamingRuleFunc func(methodName string) string

	Middlewarer interface {
		Middlewares() []gin.HandlerFunc
	}

	MethoderMiddlewarer interface {
		MethoderMiddlewares() map[string][]gin.HandlerFunc
	}

	Controller interface {
		BasePath() string
	}

	Route struct {
		Path       string
		HttpMethod string
		Handles    []gin.HandlerFunc
	}

	ApiRoute struct {
		Path       string
		HttpMethod string
		MethodName string
		ReqParams  Field
		RespParams Field
	}

	OptionFun func(*GinEngine)

	httpMethod struct {
		key string
	}
	HttpMethod struct {
		Prefix string
		Method httpMethod
	}
)

const (
	get    = "Get"
	post   = "Post"
	put    = "Put"
	del    = "Delete"
	patch  = "Patch"
	head   = "Head"
	option = "Option"
)

const (
	defaultTitle  = "github.com/aide-cloud/gin-plus"
	defaultVrsion = "v0.1.2"
)

var (
	Get    = httpMethod{key: get}
	Post   = httpMethod{key: post}
	Put    = httpMethod{key: put}
	Delete = httpMethod{key: del}
	Patch  = httpMethod{key: patch}
	Head   = httpMethod{key: head}
	Option = httpMethod{key: option}
)

// defaultPrefixes is the default prefixes.
var defaultPrefixes = map[string]httpMethod{
	get:    Get,
	post:   Post,
	put:    Put,
	del:    Delete,
	patch:  Patch,
	head:   Head,
	option: Option,
}

// New returns a GinEngine instance.
func New(r *gin.Engine, opts ...OptionFun) *GinEngine {
	instance := &GinEngine{
		Engine:              r,
		httpMethodPrefixes:  defaultPrefixes,
		defaultOpenApiYaml:  defaultOpenApiYaml,
		defaultHttpMethod:   Get,
		routeNamingRuleFunc: routeToCamel,
		apiRoutes:           make(map[string][]ApiRoute),
		genApiEnable:        true,
		apiConfig: ApiConfig{
			Title:   defaultTitle,
			Version: defaultVrsion,
		},
	}
	for _, opt := range opts {
		opt(instance)
	}

	instance.Use(instance.middlewares...)

	routes := make([]*Route, 0)
	basePath := "/"
	for _, c := range instance.controllers {
		routes = append(routes, instance.genRoute(basePath, c, nil, false)...)
	}

	for _, route := range routes {
		instance.Handle(strings.ToUpper(route.HttpMethod), path.Join(instance.basePath, route.Path), route.Handles...)
	}

	registerSwaggerUI(instance, instance.genApiEnable)

	return instance
}

func registerSwaggerUI(instance *GinEngine, enable bool) {
	if !enable {
		return
	}
	instance.genOpenApiYaml()
	fp, _ := fs.Sub(swagger.Dist, "dist")
	instance.StaticFS("/swagger-ui", http.FS(fp))
	instance.GET("/openapi/doc/swagger", func(ctx *gin.Context) {
		file, _ := os.ReadFile(instance.defaultOpenApiYaml)
		ctx.Writer.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = ctx.Writer.Write(file)
	})
}

// WithControllers sets the controllers.
func WithControllers(controllers ...any) OptionFun {
	return func(g *GinEngine) {
		g.controllers = controllers
	}
}

// WithMiddlewares sets the middlewares.
func WithMiddlewares(middlewares ...gin.HandlerFunc) OptionFun {
	return func(g *GinEngine) {
		g.middlewares = middlewares
	}
}

// WithHttpMethodPrefixes sets the prefixes.
func WithHttpMethodPrefixes(prefixes ...HttpMethod) OptionFun {
	return func(g *GinEngine) {
		prefixeHttpMethodMap := make(map[string]httpMethod)
		for _, prefix := range prefixes {
			if prefix.Prefix == "" || prefix.Method.key == "" {
				continue
			}
			prefixeHttpMethodMap[prefix.Prefix] = prefix.Method
		}
		g.httpMethodPrefixes = prefixeHttpMethodMap
	}
}

// AppendHttpMethodPrefixes append the prefixes.
func AppendHttpMethodPrefixes(prefixes ...HttpMethod) OptionFun {
	return func(g *GinEngine) {
		prefixeHttpMethodMap := g.httpMethodPrefixes
		if prefixeHttpMethodMap == nil {
			prefixeHttpMethodMap = make(map[string]httpMethod)
		}
		for _, prefix := range prefixes {
			if prefix.Prefix == "" || prefix.Method.key == "" {
				continue
			}
			prefixeHttpMethodMap[prefix.Prefix] = prefix.Method
		}
		g.httpMethodPrefixes = prefixeHttpMethodMap
	}
}

// WithBasePath sets the base path.
func WithBasePath(basePath string) OptionFun {
	return func(g *GinEngine) {
		g.basePath = path.Join("/", basePath)
	}
}

// WithDefaultHttpMethod sets the default http method.
func WithDefaultHttpMethod(method httpMethod) OptionFun {
	return func(g *GinEngine) {
		g.defaultHttpMethod = method
	}
}

// WithRouteNamingRuleFunc 自定义路由命名函数
func WithRouteNamingRuleFunc(ruleFunc RouteNamingRuleFunc) OptionFun {
	return func(g *GinEngine) {
		g.routeNamingRuleFunc = ruleFunc
	}
}

// WithApiConfig sets the title.
func WithApiConfig(c ApiConfig) OptionFun {
	return func(g *GinEngine) {
		g.apiConfig = c
	}
}

// WithOpenApiYaml 自定义api文件存储位置和文件名称
func WithOpenApiYaml(dir, filename string) OptionFun {
	return func(g *GinEngine) {
		if !strings.HasSuffix(filename, ".yaml") {
			fmt.Printf("[GIN-PLUS] [WARNING] filename has no (.yaml) suffix,  so the default (%s) is used as the filename.\n", defaultOpenApiYaml)
		}
		g.defaultOpenApiYaml = path.Join(dir, filename)
	}
}
