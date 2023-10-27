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
	if isNil(controller) {
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
	mid, isMid := isMiddleware(controller)
	if isMid {
		//Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
		middlewares = mid.Middlewares()
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

	methodMiddlewaresMap := make(map[string][]gin.HandlerFunc)
	methodMid, privateMidOk := isMethodMiddleware(controller)
	if privateMidOk {
		methodMiddlewaresMap = methodMid.MethodeMiddlewares()
	}

	if !skipAnonymous {
		for i := 0; i < t.NumMethod(); i++ {
			methodName := t.Method(i).Name
			if !isPublic(methodName) {
				continue
			}

			route := l.parseRoute(methodName)
			if route == nil {
				continue
			}

			privateMid := methodMiddlewaresMap[methodName]

			// 接口私有中间件
			route.Handles = append(route.Handles, privateMid...)

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
				l.genOpenAPI(routeGroup, req, resp, route, methodName)
				// 注册路由回调函数
				handleFunc := l.defaultHandler(controller, t.Method(i), req)
				l.registerCallHandler(route, routeGroup, handleFunc)
				continue
			}
		}
	}

	l.genStructRoute(routeGroup, controller)
}

// 生成openAPI数据
func (l *GinEngine) genOpenAPI(group *gin.RouterGroup, req, resp reflect.Type, route *Route, methodName string) {
	reqName := req.Name()
	respName := resp.Name()
	reqTagInfo := getTag(req)
	apiRoute := ApiRoute{
		Path:       route.Path,
		HttpMethod: strings.ToLower(route.HttpMethod),
		MethodName: methodName,
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

// registerCallHandler 注册回调函数
func (l *GinEngine) registerCallHandler(route *Route, routeGroup *gin.RouterGroup, handleFunc gin.HandlerFunc) {
	// 具体的action
	route.Handles = append(route.Handles, handleFunc)
	routeGroup.Handle(strings.ToUpper(route.HttpMethod), route.Path, route.Handles...)
}

// genStructRoute 递归注册结构体路由
func (l *GinEngine) genStructRoute(parentGroup *gin.RouterGroup, controller any) {
	if isNil(controller) {
		return
	}
	tmp := reflect.TypeOf(controller)
	for tmp.Kind() == reflect.Ptr {
		if tmp == nil {
			return
		}
		tmp = tmp.Elem()
	}
	if isStruct(tmp) {
		// 递归获取内部的controller
		for i := 0; i < tmp.NumField(); i++ {
			// 判断field值是否为nil
			if isNil(reflect.ValueOf(controller).Elem().Field(i).Interface()) {
				continue
			}
			field := tmp.Field(i)

			for field.Type.Kind() == reflect.Ptr {
				field.Type = field.Type.Elem()
			}
			if !isStruct(field.Type) {
				continue
			}

			if !isPublic(field.Name) {
				continue
			}

			newController := reflect.ValueOf(controller).Elem().Field(i).Interface()
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

// isMiddleware 判断是否为Controller类型
func isMiddleware(c any) (IMiddleware, bool) {
	mid, ok := c.(IMiddleware)
	return mid, ok
}

// isController 判断是否为Controller类型
func isController(c any) (Controller, bool) {
	ctrl, ok := c.(Controller)
	return ctrl, ok
}

// isMethodMiddleware 判断是否为MethodMiddleware类型
func isMethodMiddleware(c any) (MethodeMiddleware, bool) {
	mid, ok := c.(MethodeMiddleware)
	return mid, ok
}

// isStruct 判断是否为struct类型
func isStruct(t reflect.Type) bool {
	tmp := t
	for tmp.Kind() == reflect.Ptr {
		if tmp == nil {
			return false
		}
		tmp = tmp.Elem()
	}
	return tmp.Kind() == reflect.Struct
}
