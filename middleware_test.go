package ginplus

import (
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	NewMiddleware(WithResponse(NewResponse()), WithServerName("gin-plus"), WithID("id"), WithEnv("default"))
}
