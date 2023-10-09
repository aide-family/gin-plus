package ginplus

import (
	"strings"

	"github.com/spf13/viper"
)

const defaultOpenApiYaml = "openapi.yaml"

type (
	Info struct {
		Title   string `yaml:"title,omitempty"`
		Version string `yaml:"version,omitempty"`
	}

	Properties struct {
		Properties map[string]SchemaInfo `yaml:"properties,omitempty"`
		Type       string                `yaml:"type,omitempty"`
	}

	SchemaInfo struct {
		Type        string                `yaml:"type,omitempty"`
		Title       string                `yaml:"title,omitempty"`
		Format      string                `yaml:"format,omitempty"`
		Description string                `yaml:"description,omitempty"`
		Properties  map[string]SchemaInfo `yaml:"properties,omitempty"`
		Items       Properties            `yaml:"items,omitempty"`
	}

	Schema struct {
		Schema SchemaInfo `yaml:"schema,omitempty"`
	}

	ApiContent map[string]Schema

	ApiResponse struct {
		Description string     `yaml:"description"`
		Content     ApiContent `yaml:"content,omitempty"`
	}

	ApiRequest struct {
		Content ApiContent `yaml:"content,omitempty"`
	}

	Parameter struct {
		Name     string     `yaml:"name,omitempty"`
		In       string     `yaml:"in,omitempty"`
		Required bool       `yaml:"required,omitempty"`
		Schema   SchemaInfo `yaml:"schema,omitempty"`
	}

	ApiHttpMethod struct {
		OperationId string              `yaml:"operationId,omitempty"`
		Tags        []string            `yaml:"tags,omitempty"`
		Responses   map[int]ApiResponse `yaml:"responses,omitempty"`
		Parameters  []Parameter         `yaml:"parameters,omitempty"`
		RequestBody ApiRequest          `yaml:"requestBody,omitempty"`
	}

	Path map[string]map[string]ApiHttpMethod

	ApiTemplate struct {
		Openapi string          `yaml:"openapi,omitempty"`
		Info    Info            `yaml:"info,omitempty"`
		Paths   map[string]Path `yaml:"paths,omitempty"`
	}
)

func (l *GinEngine) genOpenApiYaml() {
	apiYaml := defaultOpenApiYaml
	if l.defaultOpenApiYaml != "" && strings.HasSuffix(l.defaultOpenApiYaml, ".yaml") {
		apiYaml = l.defaultOpenApiYaml
	}

	// 写入之前先清空viper
	viper.Reset()

	viper.SetConfigFile(apiYaml)
	viper.SetConfigPermissions(0644)
	viper.SetConfigType("yaml")
	viper.Set("openapi", "3.1.3")
	viper.Set("info", Info{
		Title:   l.apiConfig.Title,
		Version: l.apiConfig.Version,
	})
	viper.Set("paths", l.apiToYamlModel())

	if err := viper.WriteConfig(); err != nil {
		panic(err)
	}
}

func isUri(uriStr string) bool {
	return uriStr != "" && uriStr != "-"
}

func (l *GinEngine) apiToYamlModel() Path {
	apiReoutes := l.apiRoutes

	apiPath := make(Path)
	for url, info := range apiReoutes {
		if _, ok := apiPath[url]; !ok {
			apiPath[url] = make(map[string]ApiHttpMethod)
		}

		methodRoute := make(map[string]ApiHttpMethod)
		for _, route := range info {
			methodRoute[route.HttpMethod] = ApiHttpMethod{
				OperationId: route.MethodName,
				Tags:        nil,
				Responses: map[int]ApiResponse{
					200: {
						Content: map[string]Schema{
							"application/json": {
								Schema: SchemaInfo{
									Type:       "object",
									Title:      route.RespParams.Name,
									Properties: genProperties(route.RespParams.Info),
								},
							},
						},
					},
				},
				Parameters: func() []Parameter {
					infos := route.ReqParams.Info
					res := make([]Parameter, 0, len(infos))
					for _, fieldInfo := range infos {
						if fieldInfo.Tags.FormKey == "" && fieldInfo.Tags.UriKey == "" {
							continue
						}

						name := fieldInfo.Tags.FormKey
						in := "query"
						isUriParam := isUri(fieldInfo.Tags.UriKey)
						if isUriParam {
							name = fieldInfo.Tags.UriKey
							in = "path"
						}

						res = append(res, Parameter{
							Name:     name,
							In:       in,
							Required: isUriParam,
							Schema: SchemaInfo{
								Type:        getTypeMap(fieldInfo.Type),
								Title:       fieldInfo.Tags.Title,
								Format:      fieldInfo.Tags.Format,
								Description: fieldInfo.Tags.Desc,
							},
						})
					}

					return res
				}(),
				RequestBody: ApiRequest{
					Content: map[string]Schema{
						"application/json": {
							Schema: SchemaInfo{
								Type:       "object",
								Title:      route.ReqParams.Name,
								Properties: genProperties(route.ReqParams.Info),
							},
						},
					},
				},
			}
		}
		apiPath[url] = methodRoute
	}
	return apiPath
}

func genProperties(fieldList []FieldInfo) map[string]SchemaInfo {
	if len(fieldList) == 0 {
		return nil
	}
	resp := make(map[string]SchemaInfo)
	for _, info := range fieldList {
		jsonKey := info.Tags.JsonKey
		if jsonKey == "-" || jsonKey == "" {
			continue
		}
		fieldType := getTypeMap(info.Type)

		schema := SchemaInfo{
			Type:        fieldType,
			Title:       info.Tags.Title,
			Format:      info.Tags.Format,
			Description: info.Tags.Desc,
		}

		switch fieldType {
		case "object":
			schema.Properties = genProperties(info.Info)
		case "array":
			schema.Items.Properties = genProperties(info.Info)
			schema.Items.Type = getTypeMap(info.ChildType)
		}

		resp[info.Tags.JsonKey] = schema
	}

	return resp
}

// "array", "boolean", "integer", "null", "number", "object", "string"
func getTypeMap(typeStr string) string {
	switch typeStr {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return "integer"
	case "float32", "float64", "ufloat32", "ufloat64", "complex64", "complex128":
		return "number"
	case "boolean", "string", "array", "bool":
		return typeStr
	default:
		if strings.HasPrefix(typeStr, "[]") {
			return "array"
		}
		return "object"
	}
}
