package ginplus

import (
	"os"
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

	id, _ := os.Hostname()

	ginR.Use(
		midd.Tracing(TracingConfig{
			Name:        serverName,
			URL:         "http://localhost:14268/api/traces",
			Environment: "test",
			ID:          id,
		}),
		midd.Logger(serverName),
		midd.IpLimit(100, 0.5, "iplimit"),
		midd.Interceptor(),
	)

	r := New(ginR, WithControllers(NewHandleApiApi()))

	NewCtrlC(r).Start()
}
