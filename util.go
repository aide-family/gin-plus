package ginplus

import (
	"github.com/gin-gonic/gin"
	"path"
	"reflect"
	"strings"
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

func (l *GinEngine) genRoute(p string, controller any) []*Route {
	t := reflect.TypeOf(controller)
	var routes []*Route

	tmp := t
	for tmp.Kind() == reflect.Ptr {
		tmp = t.Elem()
	}

	var middlewares []gin.HandlerFunc
	midd, isMidd := isMiddlewarer(controller)
	if isMidd {
		//Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
		middlewares = midd.Middlewares()
	}

	basePath := path.Join(p, l.routeNamingRuleFunc(tmp.Name()))
	ctrl, isCtrl := isController(controller)
	if isCtrl {
		basePath = ctrl.BasePath()
	}

	methoderMiddlewaresMap := make(map[string][]gin.HandlerFunc)
	methoderMidd, privateMiddOk := isMethoderMiddlewarer(controller)
	if privateMiddOk {
		methoderMiddlewaresMap = methoderMidd.MethoderMiddlewares()
	}

	for i := 0; i < t.NumMethod(); i++ {
		// 解析方法名称, 生成路由, 例如: GetInfoAction -> get /info  PostPeopleAction -> post /people
		// 通过反射获取方法的返回值类型
		if isHandlerFunc(t.Method(i).Type) {
			metheodName := t.Method(i).Name
			privateMidd := methoderMiddlewaresMap[metheodName]
			route := l.parseRoute(metheodName)
			if route == nil {
				continue
			}
			route.Path = path.Join(basePath, route.Path)
			// 组下公共中间件
			route.Handles = append(route.Handles, middlewares...)
			// 接口私有中间件
			route.Handles = append(route.Handles, privateMidd...)
			// 具体的action
			route.Handles = append(route.Handles, t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().(gin.HandlerFunc))
			routes = append(routes, route)
		}
	}

	return routes
}

// parseRoute 从方法名称中解析出路由和请求方式
func (l *GinEngine) parseRoute(methodName string) *Route {
	method := strings.ToUpper(string(l.defaultHttpMethod))
	p := methodName

	for _, prefix := range l.httpMethodPrefixes {
		pre := string(prefix)
		if strings.HasPrefix(methodName, pre) {
			method = strings.ToUpper(pre)
			p = strings.TrimPrefix(methodName, pre)
			break
		}
	}

	if p == "" || method == "" {
		return nil
	}

	return &Route{
		Path:       "/" + l.routeNamingRuleFunc(p),
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
