package ginplus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"path"
	"reflect"
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

func genRoute(controller any) []*Route {
	t := reflect.TypeOf(controller)
	var routes []*Route

	tmp := t
	for tmp.Kind() == reflect.Ptr {
		tmp = t.Elem()
	}

	if tmp.Kind() != reflect.Struct {
		panic(fmt.Errorf("controller is %s, not struct or pointer to struct", tmp.Kind().String()))
	}

	var middlewares []gin.HandlerFunc
	// Controller中的Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
	for i := 0; i < t.NumMethod(); i++ {
		if t.Method(i).Name == "Middlewares" {
			middlewares = t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().([]gin.HandlerFunc)
		}
	}

	for i := 0; i < t.NumMethod(); i++ {
		// 解析方法名称, 生成路由, 例如: GetInfoAction -> get /info  PostPeopleAction -> post /people
		// 通过反射获取方法的返回值类型
		if isHandlerFunc(t.Method(i).Type) {
			route := parseRoute(t.Method(i).Name)
			if route == nil {
				continue
			}
			route.Path = path.Join("/", routeToCamel(tmp.Name()), route.Path)
			route.Handles = append(route.Handles, t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().(gin.HandlerFunc))
			route.Handles = append(append([]gin.HandlerFunc{}, middlewares...), route.Handles...)
			routes = append(routes, route)
		}

		// Controller中的Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
		if t.Method(i).Name == "Middlewares" {

		}
	}

	return routes
}

// parseRoute 从方法名称中解析出路由和请求方式
func parseRoute(methodName string) *Route {
	httpMethod := ""
	p := ""
	// 从方法名称中解析出路由和请求方式, 是否包含http请求方式前缀
	if methodName[:3] == "Get" || methodName[:3] == "GET" {
		httpMethod = "GET"
		p = fmt.Sprintf("%s", methodName[3:])
	} else if methodName[:4] == "Post" || methodName[:4] == "POST" {
		httpMethod = "POST"
		p = fmt.Sprintf("%s", methodName[4:])
	} else if methodName[:6] == "Delete" || methodName[:6] == "DELETE" {
		httpMethod = "DELETE"
		p = fmt.Sprintf("%s", methodName[6:])
	} else if methodName[:5] == "Patch" || methodName[:5] == "PATCH" {
		httpMethod = "PATCH"
		p = fmt.Sprintf("%s", methodName[5:])
	} else if methodName[:4] == "Head" || methodName[:4] == "HEAD" {
		httpMethod = "HEAD"
		p = fmt.Sprintf("%s", methodName[4:])
	} else if methodName[:3] == "Put" || methodName[:3] == "PUT" {
		httpMethod = "PUT"
		p = fmt.Sprintf("%s", methodName[3:])
	} else if methodName[:6] == "Option" {
		httpMethod = "OPTION"
		p = fmt.Sprintf("%s", methodName[6:])
	}

	if p == "" && httpMethod == "" {
		return nil
	}

	return &Route{
		Path:       "/" + routeToCamel(p),
		HttpMethod: httpMethod,
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
