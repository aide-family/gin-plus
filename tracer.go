package ginplus

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type TracingConfig struct {
	// Name 服务名称
	Name string
	// URL 上报地址
	URL string
	// Environment 环境
	Environment string
	// ID 服务ID
	ID       string
	KeyValue func(c *gin.Context) []attribute.KeyValue
}

func defaultKeyValueFunc(c *gin.Context) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.host", c.Request.Host),
		attribute.String("http.user_agent", c.Request.UserAgent()),
		attribute.String("http.client_ip", c.ClientIP()),
	}
}

func tracerProvider(url, serviceName, environment, id string) *tracesdk.TracerProvider {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		panic(err)
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", environment),
			attribute.String("ID", id),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}

const gormSpanKey = "__gorm_span"
const gormTime = "__gorm_time"

func before(db *gorm.DB) {
	// 先从父级spans生成子span ---> 这里命名为gorm，但实际上可以自定义
	// 自己喜欢的operationName
	_, span := otel.Tracer("gorm").Start(db.Statement.Context, "gorm")
	// 利用db实例去传递span
	db.InstanceSet(gormSpanKey, span)
	db.InstanceSet(gormTime, time.Now())

	return
}

func after(db *gorm.DB) {
	// 从GORM的DB实例中取出span
	_span, isExist := db.InstanceGet(gormSpanKey)
	if !isExist {
		// 不存在就直接抛弃掉
		return
	}

	// 断言进行类型转换
	span, ok := _span.(oteltrace.Span)
	if !ok {
		return
	}
	defer span.End()

	// Error
	if db.Error != nil {
		span.SetAttributes(attribute.String("error", db.Error.Error()))
	}
	// sql --> 写法来源GORM V2的日志
	span.SetAttributes(attribute.String("sql", db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
	// rows
	span.SetAttributes(attribute.Int64("rows", db.RowsAffected))
	// elapsed
	_time, isExist := db.InstanceGet(gormTime)
	if !isExist {
		return
	}
	startTime, ok := _time.(time.Time)
	if ok {
		span.SetAttributes(attribute.String("elapsed", time.Since(startTime).String()))
	}
}

const (
	callBackBeforeName = "opentracing:before"
	callBackAfterName  = "opentracing:after"
)

type OpentracingPlugin struct{}

func (op *OpentracingPlugin) Name() string {
	return "opentracingPlugin"
}

func (op *OpentracingPlugin) Initialize(db *gorm.DB) (err error) {
	// 开始前 - 并不是都用相同的方法，可以自己自定义
	db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)
	db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)
	db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before)
	db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, before)
	db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)
	db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)

	// 结束后 - 并不是都用相同的方法，可以自己自定义
	db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after)
	db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)
	db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after)
	db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after)
	db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)
	return
}

// 告诉编译器这个结构体实现了gorm.Plugin接口
var _ gorm.Plugin = &OpentracingPlugin{}
