package ginplus

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestNewCtrlC(t *testing.T) {
	ctrlC := NewCtrlC(New(gin.Default()))
	ctrlC.Start()
}
