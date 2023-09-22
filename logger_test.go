package ginplus

import (
	"testing"
)

func TestLogger(t *testing.T) {
	Logger().Info("hello world")
}
