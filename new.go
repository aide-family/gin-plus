package ginplus

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aide-cloud/gin-plus/swagger"
	"github.com/gin-gonic/gin"
)

var _ Server = (*GinEngine)(nil)

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

		// 内置启动地址
		// 默认为: :8080
		addr   string
		server *http.Server

		// graphql配置
		graphqlConfig GraphqlConfig
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

	GraphqlConfig struct {
		// Enable 是否启用
		Enable bool
		// HandlePath graphql请求路径
		HandlePath string
		// SchemaPath graphql schema文件路径
		ViewPath string
		// Root graphql 服务根节点
		Root any
		// Content graphql schema文件内容
		Content embed.FS
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
		addr: ":8080",
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

	// graphql
	registerGraphql(instance, instance.graphqlConfig)

	return instance
}

func (l *GinEngine) Start() error {
	if l.server == nil {
		//创建HTTP服务器
		server := &http.Server{
			Addr:    l.addr,
			Handler: l.Engine,
		}
		l.server = server
	}

	//启动HTTP服务器
	go func() {
		if err := l.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[GIN-PLUS] [INFO] Server listen: %s\n", err)
		}
	}()

	log.Println("[GIN-PLUS] [INFO] Server is running at", l.addr)

	return nil
}

func (l *GinEngine) Stop() {
	//创建超时上下文，Shutdown可以让未处理的连接在这个时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//停止HTTP服务器
	if err := l.server.Shutdown(ctx); err != nil {
		log.Fatal("[GIN-PLUS] [INFO] Server Shutdown:", err)
	}

	log.Println("[GIN-PLUS] [INFO] Server stopped")
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

func registerGraphql(instance *GinEngine, config GraphqlConfig) {
	if !config.Enable || config.Root == nil {
		return
	}
	if config.HandlePath == "" {
		config.HandlePath = DefaultHandlePath
	}
	if config.ViewPath == "" {
		config.ViewPath = DefaultViewPath
	}
	instance.POST(config.HandlePath, gin.WrapH(Handler(config.Root, config.Content)))
	instance.GET(config.ViewPath, gin.WrapF(View(config.HandlePath)))
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
			log.Printf("[GIN-PLUS] [WARNING] filename has no (.yaml) suffix,  so the default (%s) is used as the filename.\n", defaultOpenApiYaml)
		}
		g.defaultOpenApiYaml = path.Join(dir, filename)
	}
}

// WithGenApiEnable 设置是否生成API路由
func WithGenApiEnable(enable bool) OptionFun {
	return func(g *GinEngine) {
		g.genApiEnable = enable
	}
}

// WithAddr 设置启动地址
func WithAddr(addr string) OptionFun {
	return func(g *GinEngine) {
		g.addr = addr
	}
}

// WithHttpServer 设置启动地址
func WithHttpServer(server *http.Server) OptionFun {
	return func(g *GinEngine) {
		server.Handler = g.Engine
		if server.Addr == "" {
			server.Addr = g.addr
		}
		g.server = server
	}
}

// WithGraphqlConfig 设置graphql配置
func WithGraphqlConfig(config GraphqlConfig) OptionFun {
	return func(g *GinEngine) {
		g.graphqlConfig = config
	}
}
