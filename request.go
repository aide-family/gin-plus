package ginplus

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Bind(c *gin.Context, params interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	if err := c.ShouldBindWith(params, b); err != nil {
		return err
	}

	if err := binding.Form.Bind(c.Request, params); err != nil {
		return err
	}

	if err := c.ShouldBindUri(params); err != nil {
		return err
	}

	if err := c.ShouldBindHeader(params); err != nil {
		return err
	}

	return nil
}
