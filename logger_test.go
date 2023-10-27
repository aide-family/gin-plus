package ginplus

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	Logger().Info("hello world")
}

func TestSetLogger(t *testing.T) {
	SetLogger(zap.NewExample())
	Logger().Info("hello world")
}
