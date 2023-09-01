package ginplus

import (
	"context"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type (
	GinEngine struct {
		*gin.Engine
		middlewares        []gin.HandlerFunc
		controllers        []any
		httpMethodPrefixes []httpMethod
		basePath           string
		defaultHttpMethod  httpMethod
		// 自定义路由命名规则函数
		routeNamingRuleFunc func(methodName string) string

		// 文档配置
		apiConfig ApiConfig
		apiRoutes map[string][]ApiRoute
	}

	ApiConfig struct {
		Openapi      string
		Info         Info
		GenApiEnable bool
	}

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

	Option func(*GinEngine)

	httpMethod string

	CallBack[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)
)

const (
	Get    httpMethod = "Get"
	Post   httpMethod = "Post"
	Put    httpMethod = "Put"
	Delete httpMethod = "Delete"
	Patch  httpMethod = "Patch"
	Head   httpMethod = "Head"
	Ootion httpMethod = "Option"
)

// defaultPrefixes is the default prefixes.
var defaultPrefixes = []httpMethod{Get, Post, Put, Delete, Patch, Head, Ootion}

// New returns a GinEngine instance.
func New(r *gin.Engine, opts ...Option) *GinEngine {
	instance := &GinEngine{
		Engine:              r,
		httpMethodPrefixes:  defaultPrefixes,
		defaultHttpMethod:   Get,
		routeNamingRuleFunc: routeToCamel,
		apiRoutes:           make(map[string][]ApiRoute),
		apiConfig: ApiConfig{
			Openapi: "3.0.3",
			Info: Info{
				Title:   "aide-cloud-api",
				Version: "v1",
			},
			GenApiEnable: true,
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

	if instance.apiConfig.GenApiEnable {
		instance.genOpenApiYaml()
	}

	return instance
}

// WithControllers sets the controllers.
func WithControllers(controllers ...any) Option {
	return func(g *GinEngine) {
		g.controllers = controllers
	}
}

// WithMiddlewares sets the middlewares.
func WithMiddlewares(middlewares ...gin.HandlerFunc) Option {
	return func(g *GinEngine) {
		g.middlewares = middlewares
	}
}

// WithHttpMethodPrefixes sets the prefixes.
func WithHttpMethodPrefixes(prefixes ...httpMethod) Option {
	return func(g *GinEngine) {
		g.httpMethodPrefixes = prefixes
	}
}

// WithBasePath sets the base path.
func WithBasePath(basePath string) Option {
	return func(g *GinEngine) {
		g.basePath = path.Join("/", basePath)
	}
}

// WithDefaultHttpMethod sets the default http method.
func WithDefaultHttpMethod(method httpMethod) Option {
	return func(g *GinEngine) {
		g.defaultHttpMethod = method
	}
}

// WithRouteNamingRuleFunc 自定义路由命名函数
func WithRouteNamingRuleFunc(ruleFunc RouteNamingRuleFunc) Option {
	return func(g *GinEngine) {
		g.routeNamingRuleFunc = ruleFunc
	}
}

// WithApiConfig sets the title.
func WithApiConfig(c ApiConfig) Option {
	return func(g *GinEngine) {
		g.apiConfig = c
	}
}
