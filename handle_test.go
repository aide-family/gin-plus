package ginplus

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type HandleApi struct {
}

func (l *HandleApi) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(200, "[Get] hello world")
	}
}

func NewHandleApiApi() *HandleApi {
	return &HandleApi{}
}

func TestApiHandleFunc(t *testing.T) {
	New(gin.Default(), WithControllers(NewHandleApiApi()))
}
