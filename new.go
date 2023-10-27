package ginplus

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/aide-cloud/gin-plus/swagger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ Server = (*GinEngine)(nil)

type HandlerFunc func(controller any, t reflect.Method, req reflect.Type) gin.HandlerFunc

type (
	GinEngine struct {
		*gin.Engine

		// 中间件
		middlewares []gin.HandlerFunc
		// 控制器
		controllers []any
		// 绑定前缀和http请求方法的映射
		httpMethodPrefixes map[string]httpMethod
		// 路由组基础路径
		basePath string
		// 自定义路由命名规则函数
		routeNamingRuleFunc func(methodName string) string
		// 自定义handler函数
		defaultHandler HandlerFunc
		// 自定义Response接口实现
		defaultResponse IResponse
		// 默认Bind函数
		defaultBind func(c *gin.Context, params any) error

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

		// prom metrics
		metrics *Metrics
		// ping
		ping *Ping
	}

	// Metrics prometheus metrics配置
	Metrics struct {
		Path string
	}

	// Ping ping配置
	Ping struct {
		HandlerFunc gin.HandlerFunc
	}

	// ApiConfig 文档配置,
	ApiConfig Info

	// RouteNamingRuleFunc 自定义路由命名函数
	RouteNamingRuleFunc func(methodName string) string

	// IMiddleware 中间件接口, 实现该接口的结构体将会被把中间件添加到该路由组的公共中间件中
	IMiddleware interface {
		Middlewares() []gin.HandlerFunc
	}

	// MethodeMiddleware 中间件接口, 为每个方法添加中间件
	MethodeMiddleware interface {
		MethodeMiddlewares() map[string][]gin.HandlerFunc
	}

	// Controller 控制器接口, 实现该接口的对象可以自定义模块的路由
	Controller interface {
		BasePath() string
	}

	// Route 路由参数结构
	Route struct {
		Path       string
		HttpMethod string
		Handles    []gin.HandlerFunc
	}

	// ApiRoute api路由参数结构, 用于生成文档
	ApiRoute struct {
		Path       string
		HttpMethod string
		MethodName string
		ReqParams  Field
		RespParams Field
	}

	// OptionFun GinEngine配置函数
	OptionFun func(*GinEngine)

	// httpMethod http请求方法
	httpMethod struct {
		key string
	}
	// HttpMethod http请求方法, 绑定前缀和http请求方法的映射
	HttpMethod struct {
		Prefix string
		Method httpMethod
	}

	// GraphqlConfig graphql配置
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
	defaultTitle       = "github.com/aide-cloud/gin-plus"
	defaultVersion     = "v0.5.0"
	defaultMetricsPath = "/metrics"
	defaultPingPath    = "/ping"
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

// New 返回一个GinEngine实例
func New(r *gin.Engine, opts ...OptionFun) *GinEngine {
	instance := &GinEngine{
		Engine:              r,
		httpMethodPrefixes:  defaultPrefixes,
		defaultOpenApiYaml:  defaultOpenApiYaml,
		defaultResponse:     NewResponse(),
		defaultBind:         Bind,
		routeNamingRuleFunc: routeToCamel,
		apiRoutes:           make(map[string][]ApiRoute),
		genApiEnable:        true,
		apiConfig: ApiConfig{
			Title:   defaultTitle,
			Version: defaultVersion,
		},
		ping: &Ping{HandlerFunc: func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		}},
		metrics: &Metrics{Path: defaultMetricsPath},
		addr:    ":8080",
	}
	for _, opt := range opts {
		opt(instance)
	}

	if instance.defaultHandler == nil {
		instance.defaultHandler = instance.newDefaultHandler
	}

	instance.Use(instance.middlewares...)

	for _, c := range instance.controllers {
		instance.genRoute(nil, c, false)
	}

	return instance
}

// Start 启动HTTP服务器
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
			Logger().Sugar().Infof("[GIN-PLUS] [INFO] Server listen: %s\n", err)
		}
	}()

	Logger().Sugar().Infof("[GIN-PLUS] [INFO] Server is running at %s", l.addr)

	return nil
}

// Stop 停止HTTP服务器
func (l *GinEngine) Stop() {
	//创建超时上下文，Shutdown可以让未处理的连接在这个时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//停止HTTP服务器
	if err := l.server.Shutdown(ctx); err != nil {
		Logger().Sugar().Errorf("[GIN-PLUS] [INFO] Server Shutdown: %v", err)
	}

	Logger().Sugar().Infof("[GIN-PLUS] [INFO] Server stopped")
}

func registerPing(instance *GinEngine, ping *Ping) {
	if ping != nil {
		instance.GET(defaultPingPath, ping.HandlerFunc)
		return
	}
}

func registerMetrics(instance *GinEngine, metrics *Metrics) {
	if metrics == nil {
		return
	}
	if metrics.Path == "" {
		metrics.Path = defaultMetricsPath
	}
	instance.GET(metrics.Path, gin.WrapH(promhttp.Handler()))
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

func (l *GinEngine) RegisterPing(ping ...*Ping) *GinEngine {
	if len(ping) > 0 {
		l.ping = ping[0]
	}
	registerPing(l, l.ping)
	return l
}

// RegisterMetrics 注册prometheus metrics
func (l *GinEngine) RegisterMetrics(metrics ...*Metrics) *GinEngine {
	if len(metrics) > 0 {
		l.metrics = metrics[0]
	}
	registerMetrics(l, l.metrics)
	return l
}

// RegisterSwaggerUI 注册swagger ui
func (l *GinEngine) RegisterSwaggerUI() *GinEngine {
	registerSwaggerUI(l, l.genApiEnable)
	return l
}

// RegisterGraphql 注册graphql
func (l *GinEngine) RegisterGraphql(config ...*GraphqlConfig) *GinEngine {
	if len(config) > 0 {
		l.graphqlConfig = *config[0]
	}
	registerGraphql(l, l.graphqlConfig)
	return l
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
		prefixHttpMethodMap := make(map[string]httpMethod)
		for _, prefix := range prefixes {
			if prefix.Prefix == "" || prefix.Method.key == "" {
				continue
			}
			prefixHttpMethodMap[prefix.Prefix] = prefix.Method
		}
		g.httpMethodPrefixes = prefixHttpMethodMap
	}
}

// AppendHttpMethodPrefixes append the prefixes.
func AppendHttpMethodPrefixes(prefixes ...HttpMethod) OptionFun {
	return func(g *GinEngine) {
		prefixHttpMethodMap := g.httpMethodPrefixes
		if prefixHttpMethodMap == nil {
			prefixHttpMethodMap = make(map[string]httpMethod)
		}
		for _, prefix := range prefixes {
			if prefix.Prefix == "" || prefix.Method.key == "" {
				continue
			}
			prefixHttpMethodMap[prefix.Prefix] = prefix.Method
		}
		g.httpMethodPrefixes = prefixHttpMethodMap
	}
}

// WithBasePath sets the base path.
func WithBasePath(basePath string) OptionFun {
	return func(g *GinEngine) {
		g.basePath = path.Join("/", basePath)
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
			Logger().Sugar().Infof("[GIN-PLUS] [WARNING] filename has no (.yaml) suffix,  so the default (%s) is used as the filename.\n", defaultOpenApiYaml)
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

// WithDefaultHandler 自定义handler函数
func WithDefaultHandler(handler HandlerFunc) OptionFun {
	return func(g *GinEngine) {
		g.defaultHandler = handler
	}
}

// WithDefaultResponse 自定义Response接口实现
func WithDefaultResponse(response IResponse) OptionFun {
	return func(g *GinEngine) {
		g.defaultResponse = response
	}
}

// WithMetrics 自定义Metrics
func WithMetrics(metrics *Metrics) OptionFun {
	return func(g *GinEngine) {
		g.metrics = metrics
	}
}

// WithPing 自定义Ping
func WithPing(ping *Ping) OptionFun {
	return func(g *GinEngine) {
		g.ping = ping
	}
}

// WithBind 自定义Bind
func WithBind(bind func(c *gin.Context, params any) error) OptionFun {
	return func(g *GinEngine) {
		g.defaultBind = bind
	}
}
