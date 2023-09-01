package ginplus

import (
	"testing"
)

func TestGinEngine_execute(t *testing.T) {
	instance := &GinEngine{
		Title:   "aide-cloud-api",
		Version: "1.0.0",
	}
	instance.genOpenApiYaml(instance.apiToYamlModel())
}
