package ginplus

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IResponser interface {
	Response(ctx *gin.Context, resp any, err error)
}

type IValidater interface {
	Validate() error
}

type response struct {
	Error error `json:"error"`
	Data  any   `json:"data"`
}

var _ IResponser = (*response)(nil)

func NewResponse() IResponser {
	return &response{}
}

func (l *response) Response(ctx *gin.Context, resp any, err error) {
	defer ctx.Abort()
	ctx.JSON(200, &response{
		Error: err,
		Data:  resp,
	})
}

func (l *GinEngine) newDefaultHandler(controller any, t reflect.Method, req reflect.Type) gin.HandlerFunc {
	// 缓存反射数据, 避免在请求中再处理导致性能问题
	reqTmp := req
	for reqTmp.Kind() == reflect.Ptr {
		reqTmp = reqTmp.Elem()
	}

	handleFunc := t.Func
	controllerVal := reflect.ValueOf(controller)
	return func(ctx *gin.Context) {
		// new一个req的实例
		reqVal := reflect.New(reqTmp)
		// 绑定请求参数
		if err := l.defaultBind(ctx, reqVal.Interface()); err != nil {
			Logger().Info("defaultBind req err", zap.Error(err))
			l.defaultResponse.Response(ctx, nil, err)
			return
		}

		// Validate
		if validater, ok := reqVal.Interface().(IValidater); ok {
			if err := validater.Validate(); err != nil {
				Logger().Info("Validate req err", zap.Error(err))
				l.defaultResponse.Response(ctx, nil, err)
				return
			}
		}

		// 调用方法
		respVal := handleFunc.Call([]reflect.Value{controllerVal, reflect.ValueOf(ctx), reqVal})
		if !respVal[1].IsNil() {
			err, ok := respVal[1].Interface().(error)
			if ok {
				Logger().Info("handleFunc Call err", zap.Error(err))
				l.defaultResponse.Response(ctx, nil, err)
				return
			}
			Logger().Info("handleFunc Call abnormal err", zap.Error(err))
			l.defaultResponse.Response(ctx, nil, errors.New("response error"))
			return
		}

		// 返回结果
		l.defaultResponse.Response(ctx, respVal[0].Interface(), nil)
	}
}
