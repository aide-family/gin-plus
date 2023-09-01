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

func (l *GinEngine) genRoute(p string, controller any, skipAnonymous bool) []*Route {
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

	if !skipAnonymous {
		for i := 0; i < t.NumMethod(); i++ {
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

			if isHandlerFunc(t.Method(i).Type) {
				// 具体的action
				route.Handles = append(route.Handles, t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().(gin.HandlerFunc))
				routes = append(routes, route)
				continue
			}

			// 判断是否为CallBack类型
			req, resp, isCb := isCallBack(t.Method(i).Type)
			if isCb {
				reqName := req.Name()
				respName := resp.Name()
				apiRoute := ApiRoute{
					Path:       route.Path,
					HttpMethod: strings.ToLower(route.HttpMethod),
					MethodName: metheodName,
					ReqParams: Field{
						Name: reqName,
						Info: getTag(req),
					},
					RespParams: Field{
						Name: respName,
						Info: getTag(resp),
					},
				}
				if _, ok := l.ApiRoutes[route.Path]; !ok {
					l.ApiRoutes[route.Path] = make([]ApiRoute, 0, 1)
				}
				l.ApiRoutes[route.Path] = append(l.ApiRoutes[route.Path], apiRoute)

				// 具体的action
				route.Handles = append(route.Handles, newDefaultHandler(controller, t.Method(i), req))
				routes = append(routes, route)
				continue
			}
		}
	}

	if isStruct(tmp) {
		// 递归获取内部的controller
		for i := 0; i < tmp.NumField(); i++ {
			field := tmp.Field(i)
			for field.Type.Kind() == reflect.Ptr {
				field.Type = field.Type.Elem()
			}
			if !isStruct(field.Type) {
				continue
			}

			// new一个新的controller
			newController := reflect.New(field.Type).Interface()
			routes = append(routes, l.genRoute(basePath, newController, field.Anonymous)...)
		}
	}

	return routes
}

//// getFields 获取结构体的字段
//func getFields(t reflect.Type, parentName string) map[string]Param {
//	if t.Kind() != reflect.Struct || t.Kind() != reflect.Ptr {
//		return nil
//	}
//	parentNameTmp := parentName
//	if parentName != "" {
//		parentName = parentName + "."
//	}
//	fields := make(map[string]Param)
//	for i := 0; i < t.NumField(); i++ {
//		field := t.FieldInfo(i)
//		for field.Type.Kind() == reflect.Ptr {
//			field.Type = field.Type.Elem()
//		}
//
//		fieldName := parentNameTmp + field.Name
//
//		if !isStruct(field.Type) {
//			fields[fieldName] = Param{
//				Name:  fieldName,
//				Type:  field.Type.String(),
//				FieldInfo: nil,
//			}
//			continue
//		}
//
//		fields[fieldName] = Param{
//			Name:  fieldName,
//			Type:  field.Type.String(),
//			FieldInfo: getFields(field.Type, fieldName),
//		}
//	}
//	return fields
//}

// parseRoute 从方法名称中解析出路由和请求方式
func (l *GinEngine) parseRoute(methodName string) *Route {
	method := strings.ToLower(string(l.defaultHttpMethod))
	p := methodName

	for _, prefix := range l.httpMethodPrefixes {
		pre := string(prefix)
		if strings.HasPrefix(methodName, pre) {
			method = strings.ToLower(pre)
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

// isStruct 判断是否为struct类型
func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}
