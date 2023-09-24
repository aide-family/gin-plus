package ginplus

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	GinHandleFunc = "gin.HandlerFunc"
)

// 判断该方法返回值是否为 gin.HandlerFunc类型
func isHandlerFunc(t reflect.Type) bool {
	if t.Kind() != reflect.Func {
		return false
	}
	if t.NumOut() != 1 {
		return false
	}
	return t.Out(0).String() == GinHandleFunc
}

// isCallBack 判断是否为CallBack类型
func isCallBack(t reflect.Type) (reflect.Type, reflect.Type, bool) {
	// 通过反射获取方法的返回值类型
	if t.Kind() != reflect.Func {
		return nil, nil, false
	}

	if t.NumIn() != 3 || t.NumOut() != 2 {
		return nil, nil, false
	}

	if t.Out(1).String() != "error" {
		return nil, nil, false
	}

	if t.In(1).String() != "context.Context" {
		return nil, nil, false
	}

	// new一个out 0的实例和in 2的实例
	req := t.In(2)
	resp := t.Out(0)

	return req, resp, true
}

func isNil(value interface{}) bool {
	// 使用反射获取值的类型和值
	val := reflect.ValueOf(value)

	// 检查值的类型
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		// 对于指针、接口、切片、映射、通道和函数类型，使用IsNil()方法来检查是否为nil
		return val.IsNil()
	default:
		// 对于其他类型，无法直接检查是否为nil
		return false
	}
}

func (l *GinEngine) GenRoute(parentGroup *gin.RouterGroup, controller any) *GinEngine {
	if isNil(controller) {
		logger.Warn("controller is nil")
		return l
	}
	l.genRoute(parentGroup, controller, false)
	return l
}

func (l *GinEngine) genRoute(parentGroup *gin.RouterGroup, controller any, skipAnonymous bool) {
	if controller == nil {
		return
	}
	t := reflect.TypeOf(controller)

	tmp := t
	for tmp.Kind() == reflect.Ptr {
		if tmp == nil {
			return
		}
		tmp = tmp.Elem()
	}

	if !isPublic(tmp.Name()) {
		return
	}

	var middlewares []gin.HandlerFunc
	midd, isMidd := isMiddlewarer(controller)
	if isMidd {
		//Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
		middlewares = midd.Middlewares()
	}

	basePath := l.routeNamingRuleFunc(tmp.Name())
	ctrl, isCtrl := isController(controller)
	if isCtrl {
		basePath = ctrl.BasePath()
	}
	parentRouteGroup := parentGroup
	if parentRouteGroup == nil {
		parentRouteGroup = l.Group("/")
	}
	routeGroup := parentRouteGroup.Group(path.Join(basePath), middlewares...)

	methoderMiddlewaresMap := make(map[string][]gin.HandlerFunc)
	methoderMidd, privateMiddOk := isMethoderMiddlewarer(controller)
	if privateMiddOk {
		methoderMiddlewaresMap = methoderMidd.MethoderMiddlewares()
	}

	if !skipAnonymous {
		for i := 0; i < t.NumMethod(); i++ {
			metheodName := t.Method(i).Name
			if !isPublic(metheodName) {
				continue
			}

			route := l.parseRoute(metheodName)
			if route == nil {
				continue
			}

			privateMidd := methoderMiddlewaresMap[metheodName]

			// 接口私有中间件
			route.Handles = append(route.Handles, privateMidd...)

			if isHandlerFunc(t.Method(i).Type) {
				// 具体的action
				handleFunc := t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().(gin.HandlerFunc)
				route.Handles = append(route.Handles, handleFunc)
				routeGroup.Handle(strings.ToUpper(route.HttpMethod), route.Path, route.Handles...)
				continue
			}

			// 判断是否为CallBack类型
			req, resp, isCb := isCallBack(t.Method(i).Type)
			if isCb {
				// 生成路由openAPI数据
				l.genOpenAPI(routeGroup, req, resp, route, metheodName)
				// 注册路由回调函数
				handleFunc := l.defaultHandler(controller, t.Method(i), req)
				l.registerCallhandler(route, routeGroup, handleFunc)
				continue
			}
		}
	}

	l.genStructRoute(routeGroup, tmp)
}

// 生成openAPI数据
func (l *GinEngine) genOpenAPI(group *gin.RouterGroup, req, resp reflect.Type, route *Route, metheodName string) {
	reqName := req.Name()
	respName := resp.Name()
	reqTagInfo := getTag(req)
	apiRoute := ApiRoute{
		Path:       route.Path,
		HttpMethod: strings.ToLower(route.HttpMethod),
		MethodName: metheodName,
		ReqParams: Field{
			Name: reqName,
			Info: reqTagInfo,
		},
		RespParams: Field{
			Name: respName,
			Info: getTag(resp),
		},
	}

	// 处理Uri参数
	for _, tagInfo := range reqTagInfo {
		uriKey := tagInfo.Tags.UriKey
		skip := tagInfo.Tags.Skip
		if uriKey != "" && uriKey != "-" && skip != "true" {
			route.Path = path.Join(route.Path, fmt.Sprintf(":%s", uriKey))
		}
	}

	apiPath := path.Join(group.BasePath(), route.Path)
	apiRoute.Path = apiPath

	if _, ok := l.apiRoutes[apiPath]; !ok {
		l.apiRoutes[apiPath] = make([]ApiRoute, 0, 1)
	}
	l.apiRoutes[apiPath] = append(l.apiRoutes[apiPath], apiRoute)
}

// registerCallhandler 注册回调函数
func (l *GinEngine) registerCallhandler(route *Route, routeGroup *gin.RouterGroup, handleFunc gin.HandlerFunc) {
	// 具体的action
	route.Handles = append(route.Handles, handleFunc)
	routeGroup.Handle(strings.ToUpper(route.HttpMethod), route.Path, route.Handles...)
}

// genStructRoute 递归注册结构体路由
func (l *GinEngine) genStructRoute(parentGroup *gin.RouterGroup, controller reflect.Type) {
	if isNil(controller) {
		return
	}
	tmp := controller
	if isStruct(tmp) {
		// 递归获取内部的controller
		for i := 0; i < tmp.NumField(); i++ {
			field := tmp.Field(i)
			for field.Type.Kind() == reflect.Ptr {
				if field.Type == nil {
					break
				}
				field.Type = field.Type.Elem()
			}
			if !isStruct(field.Type) {
				continue
			}

			if !isPublic(field.Name) {
				continue
			}

			// new一个新的controller
			newController := reflect.New(field.Type).Interface()
			l.genRoute(parentGroup, newController, field.Anonymous)
		}
	}
}

// isPublic 判断是否为公共方法
func isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}

	first := name[0]
	if first < 'A' || first > 'Z' {
		return false
	}

	return true
}

// parseRoute 从方法名称中解析出路由和请求方式
func (l *GinEngine) parseRoute(methodName string) *Route {
	method := ""
	routePath := ""

	for prefix, httpMethodKey := range l.httpMethodPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			method = strings.ToLower(httpMethodKey.key)
			routePath = strings.TrimPrefix(methodName, prefix)
			if routePath == "" {
				routePath = strings.ToLower(methodName)
			}
			break
		}
	}

	if method == "" || routePath == "" {
		return nil
	}

	return &Route{
		Path:       path.Join("/", l.routeNamingRuleFunc(routePath)),
		HttpMethod: method,
	}
}

// routeToCamel 将路由转换为驼峰命名
func routeToCamel(route string) string {
	if route == "" {
		return ""
	}

	// 首字母小写
	if route[0] >= 'A' && route[0] <= 'Z' {
		route = string(route[0]+32) + route[1:]
	}

	return route
}

// isMiddlewarer 判断是否为Controller类型
func isMiddlewarer(c any) (Middlewarer, bool) {
	midd, ok := c.(Middlewarer)
	return midd, ok
}

// isController 判断是否为Controller类型
func isController(c any) (Controller, bool) {
	ctrl, ok := c.(Controller)
	return ctrl, ok
}

// isMethoderMiddlewarer 判断是否为MethoderMiddlewarer类型
func isMethoderMiddlewarer(c any) (MethoderMiddlewarer, bool) {
	midd, ok := c.(MethoderMiddlewarer)
	return midd, ok
}

// isStruct 判断是否为struct类型
func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}
