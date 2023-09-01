package ginplus

import (
	"errors"
	"github.com/gin-gonic/gin"
	"reflect"
	"strings"
)

type Response struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func newDefaultHandler(controller any, t reflect.Method, req reflect.Type) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqTmp := req
		for reqTmp.Kind() == reflect.Ptr {
			reqTmp = reqTmp.Elem()
		}
		// new一个req的实例
		reqVal := reflect.New(reqTmp)

		// 绑定请求参数
		if err := Bind(ctx, reqVal.Interface()); err != nil {
			ErrorResponse(ctx, err, "request error1")
			return
		}

		// 调用方法
		respVal := t.Func.Call([]reflect.Value{reflect.ValueOf(controller), reflect.ValueOf(ctx), reqVal})
		if !respVal[1].IsNil() {
			err, ok := respVal[1].Interface().(error)
			if ok {
				ErrorResponse(ctx, err, "request error2")
				return
			}
			ErrorResponse(ctx, errors.New("unknown error"), "request error3")
			return
		}

		// 返回结果
		DefaultResponse(ctx, respVal[0].Interface(), nil)
	}
}

func DefaultResponse(ctx *gin.Context, resp any, err error) {
	if err != nil {
		ctx.JSON(500, Response{
			Code:  1,
			Msg:   "request error",
			Data:  nil,
			Error: err.Error(),
		})
		return
	}
	ctx.JSON(200, Response{
		Code:  0,
		Msg:   "success",
		Data:  resp,
		Error: "",
	})
}

func SuccessResponse(ctx *gin.Context, resp any, msg ...string) {
	ctx.JSON(200, Response{
		Code:  0,
		Msg:   strings.Join(msg, ","),
		Data:  resp,
		Error: "nil",
	})
}

func ErrorResponse(ctx *gin.Context, err error, msg ...string) {
	if err == nil {
		SuccessResponse(ctx, nil, msg...)
		return
	}
	ctx.JSON(200, Response{
		Code:  1,
		Msg:   strings.Join(msg, ","),
		Data:  nil,
		Error: err.Error(),
	})
}
