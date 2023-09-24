package ginplus

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

type Student struct {
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
	var p *Student
	p = &Student{}
	pr := New(r).GenRoute(r.Group("/api"), p)
	NewCtrlC(pr)
}
