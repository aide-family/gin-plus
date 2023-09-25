package ginplus

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

type Student struct {
	A *AApi
	B *BApi
}

type AApi struct {
}

func (a *AApi) Get(ctx context.Context, req *StudentReq) (*StudentResp, error) {
	return &StudentResp{}, nil
}

type BApi struct {
}

func (b *BApi) Get(ctx context.Context, req *StudentReq) (*StudentResp, error) {
	return &StudentResp{}, nil
}

type StudentReq struct {
}

type StudentResp struct {
}

func (s *Student) Get(ctx context.Context, req *StudentReq) (*StudentResp, error) {
	return &StudentResp{}, nil
}

func TestGenRoute(t *testing.T) {
	r := gin.Default()
	// var p *Student
	p := &Student{
		A: &AApi{},
		B: &BApi{},
	}
	pr := New(r, WithControllers(p))
	NewCtrlC(pr)
}
