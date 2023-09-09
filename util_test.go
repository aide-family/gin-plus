package ginplus

import (
	"context"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestGinEngine_parseRoute(t *testing.T) {
	instance := New(gin.Default())
	t.Log(instance.parseRoute("Delete"))
}

func TestGinEngine_isPublic(t *testing.T) {
	if !isPublic("Abc") {
		t.Log("Abc is public")
	}

	if isPublic("abc") {
		t.Log("abc is privite")
	}
}

type Pub struct {
	v1  *V1
	V1x *V1
	*V1
}

type V1 struct {
}

type PubReq struct {
	EID  string `uri:"eid"`
	Id   int    `uri:"id"`
	Name string `json:"name"`
}

type PubResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (p *Pub) GetPingA(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}

func (p *Pub) getPingB(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}

func (p *V1) GetList(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}

func TestPriviteMethod(t *testing.T) {
	New(gin.Default(), WithControllers(&Pub{
		v1: &V1{},
	}), AppendHttpMethodPrefixes(
		HttpMethod{
			Prefix: "get",
			Method: httpMethod{get},
		},
	))
}

func TestGinEngine_genStructRoute(t *testing.T) {
	New(gin.Default(), WithControllers(&Pub{
		v1: &V1{},
	}), AppendHttpMethodPrefixes(
		HttpMethod{
			Prefix: "Get",
			Method: httpMethod{get},
		},
	))
}

type GenRouteV1 struct{}
type GenRouteV2 struct{}
type GenRouteApi struct {
	ApiV1 *GenRouteV1
	ApiV2 *GenRouteV2
}

func (g *GenRouteV1) GetPing(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}
func (g *GenRouteV2) PostPing(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}
func (g *GenRouteApi) PutPing(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}
func (g *GenRouteApi) DeletePing(ctx context.Context, req *PubReq) (*PubResp, error) {
	return nil, nil
}

func TestGinEngine_genRoute(t *testing.T) {
	New(gin.Default(), WithControllers(&GenRouteApi{
		ApiV1: &GenRouteV1{},
		ApiV2: &GenRouteV2{},
	}))
}

type LogicApi struct{}

type LogicApiReq struct {
	EID string `uri:"eid" skip:"true"`
	PID uint   `uri:"pid" skip:"true"`
	Id  int    `uri:"id"`
}
type LogicApiResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (l *LogicApi) GetPing(ctx context.Context, req *LogicApiReq) (*LogicApiResp, error) {
	return &LogicApiResp{
		Code: 0,
		Msg:  "ok",
		Data: req,
	}, nil
}

func TestGinEngine_GenRoute(t *testing.T) {
	i := New(gin.Default())
	group := i.Group("/enterprise/:eid/project/:pid")
	i.GenRoute(group, &LogicApi{}).RegisterSwaggerUI()
	NewCtrlC(i).Start()
}
