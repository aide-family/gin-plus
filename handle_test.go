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
