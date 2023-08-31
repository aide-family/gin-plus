package ginplus

import (
	"context"
	"github.com/gin-gonic/gin"
	"reflect"
	"testing"
)

type A struct {
}

type B struct {
}

func (a *A) Middlewares() []gin.HandlerFunc {
	return nil
}

func Test_isController(t *testing.T) {
	type args struct {
		c any
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				c: &A{},
			},
			want: true,
		},
		{
			name: "test",
			args: args{
				c: &B{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := isMiddlewarer(tt.args.c)
			if got != tt.want {
				t.Errorf("isController() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	Call struct {
	}
	Req struct {
		Name string `json:"name"`
	}
	Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
)

func (c *Call) CallBack(ctx context.Context, req Req) (Resp, error) {
	return Resp{}, nil
}

func Test_isCallBack(t *testing.T) {
	ty := reflect.TypeOf(&Call{})
	m, ok := ty.MethodByName("CallBack")
	if !ok {
		t.Error("not found method")
	}

	req, resp, ok := isCallBack(m.Type)
	if ok {
		t.Error("not found method CallBack: ", m.Type.String())
	}

	t.Log(req, resp)
}
