package ginplus

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"testing"
)

type (
	Api struct {
	}

	ApiDetailReq struct {
		Id uint `uri:"id"`
	}

	ApiDetailResp struct {
		Id     uint   `json:"id"`
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}

	ApiListReq struct {
		Current   int    `form:"current"`
		Size      int    `form:"size"`
		Keryworld string `form:"keyworld"`
	}
	ApiListResp struct {
		Total int64          `json:"total"`
		List  []*ApiInfoItem `json:"list"`
	}

	ApiInfoItem struct {
		Name   string `json:"name"`
		Id     uint   `json:"id"`
		Remark string `json:"remark"`
	}

	ApiUpdateReq struct {
		Id     uint   `uri:"id"`
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}
	ApiUpdateResp struct {
		Id uint `json:"id"`
	}

	DelApiReq struct {
		Id uint `uri:"id"`
	}

	DelApiResp struct {
		Id uint `json:"id"`
	}
)

func (l *Api) GetDetail(ctx context.Context, req *ApiDetailReq) (*ApiDetailResp, error) {
	log.Println("Api.GetDetail")
	return &ApiDetailResp{
		Id:     req.Id,
		Name:   "demo",
		Remark: "hello world",
	}, nil
}

func (l *Api) GetList(ctx context.Context, req *ApiListReq) (*ApiListResp, error) {
	log.Println("Api.GetList")
	return &ApiListResp{
		Total: 100,
		List: []*ApiInfoItem{
			{
				Id:     10,
				Name:   "demo",
				Remark: "hello world",
			},
		},
	}, nil
}

func (l *Api) UpdateInfo(ctx context.Context, req *ApiUpdateReq) (*ApiUpdateResp, error) {
	log.Println("Api.UpdateInfo")
	return &ApiUpdateResp{Id: req.Id}, nil
}

func (l *Api) DeleteInfo(ctx context.Context, req *DelApiReq) (*DelApiResp, error) {
	log.Println("Api.DeleteInfo")
	return &DelApiResp{Id: req.Id}, nil
}

func TestGinEngine_execute(t *testing.T) {
	New(gin.Default(), WithControllers(&Api{}))
}
