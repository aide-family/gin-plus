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
	ginR := gin.New()

	midd := NewMiddleware(NewResponse())
	serverName := "gin-plus"

	ginR.Use(midd.Tracing(serverName), midd.Logger(serverName))

	r := New(ginR, WithControllers(NewHandleApiApi()))

	NewCtrlC(r).Start()
}
