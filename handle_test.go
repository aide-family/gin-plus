package ginplus

import (
	"embed"
	"testing"

	"github.com/gin-gonic/gin"
)

// Content holds all the SDL file content.
//
//go:embed sdl
var content embed.FS

type Root struct{}

func (r *Root) Ping() string {
	return "pong"
}

func TestGraphql(t *testing.T) {
	instance := New(gin.Default(), WithGraphqlConfig(GraphqlConfig{
		Enable:     true,
		HandlePath: "/graphql",
		ViewPath:   "/graphql",
		Root:       &Root{},
		Content:    content,
	}))

	instance.Run(":8080")
}

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
