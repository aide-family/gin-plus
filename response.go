package ginplus

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type IResponser interface {
	Response(ctx *gin.Context, resp any, err error, msg ...string)
}

type IValidater interface {
	Validate() error
}

type response struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  any    `json:"data"`
	Error string `json:"error"`
}

var _ IResponser = (*response)(nil)

func NewResponse() IResponser {
	return &response{}
}

func (l *response) Response(ctx *gin.Context, resp any, err error, msg ...string) {
	defer ctx.Abort()
	if err != nil {
		ctx.JSON(500, response{
			Code:  1,
			Msg:   strings.Join(msg, ","),
			Data:  nil,
			Error: err.Error(),
		})
		return
	}
	ctx.JSON(200, response{
		Code:  0,
		Msg:   strings.Join(msg, ","),
		Data:  resp,
		Error: "",
	})
}

func (l *GinEngine) newDefaultHandler(controller any, t reflect.Method, req reflect.Type) gin.HandlerFunc {
	// 缓存反射数据, 避免在请求中再处理导致性能问题
	reqTmp := req
	for reqTmp.Kind() == reflect.Ptr {
		reqTmp = reqTmp.Elem()
	}
	// new一个req的实例
	reqVal := reflect.New(reqTmp)

	handleFunc := t.Func
	controllerVal := reflect.ValueOf(controller)
	return func(ctx *gin.Context) {
		// 绑定请求参数
		if err := l.defaultBind(ctx, reqVal.Interface()); err != nil {
			l.defaultResponse.Response(ctx, nil, err, "request params bind error")
			return
		}

		// Validate
		if validater, ok := reqVal.Interface().(IValidater); ok {
			if err := validater.Validate(); err != nil {
				l.defaultResponse.Response(ctx, nil, err, "request params validate error")
				return
			}
		}

		// 调用方法
		respVal := handleFunc.Call([]reflect.Value{controllerVal, reflect.ValueOf(ctx), reqVal})
		if !respVal[1].IsNil() {
			err, ok := respVal[1].Interface().(error)
			if ok {
				l.defaultResponse.Response(ctx, nil, err, "request logic error")
				return
			}
			l.defaultResponse.Response(ctx, nil, errors.New("unknown error"), "request server error")
			return
		}

		// 返回结果
		l.defaultResponse.Response(ctx, respVal[0].Interface(), nil, "success")
	}
}
