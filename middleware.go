package ginplus

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Middleware struct {
	resp        IResponser
	tracing     oteltrace.TracerProvider
	propagators propagation.TextMapPropagator
}

// NewMiddleware 创建中间件
func NewMiddleware(resp IResponser) *Middleware {
	return &Middleware{
		resp: resp,
	}
}

// Cors 直接放行所有跨域请求并放行所有 OPTIONS 方法
func (l *Middleware) Cors(headers ...map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin == "null" || origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token,X-Token,X-User-Id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS,DELETE,PUT")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, New-Token, New-Expires-At")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		for _, header := range headers {
			for k, v := range header {
				c.Writer.Header().Set(k, v)
			}
		}

		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Max-Age", "3600")
			l.resp.Response(c, nil, nil)
			return
		}
		// 处理请求
		c.Next()
	}
}

type InterceptorConfig struct {
	// IP 如果IP不为空, 则不允许该IP访问, 否则所有IP都不允许访问
	IPList []string
	Method string
	Path   string
	Msg    any
}

// Interceptor 拦截器, 拦截指定API, 用于控制API当下不允许访问
func (l *Middleware) Interceptor(configs ...InterceptorConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, config := range configs {
			if config.Method == c.Request.Method && config.Path == c.Request.URL.Path {
				if len(config.IPList) != 0 {
					clientIP := c.ClientIP()
					for _, ip := range config.IPList {
						if ip == clientIP {
							l.resp.Response(c, config.Msg, nil)
							return
						}
					}
				} else {
					l.resp.Response(c, config.Msg, nil)
					return
				}
			}
		}
		c.Next()
	}
}

type TokenBucket struct {
	capacity  int64      // 桶的容量
	rate      float64    // 令牌放入速率
	tokens    float64    // 当前令牌数量
	lastToken time.Time  // 上一次放令牌的时间
	mtx       sync.Mutex // 互斥锁
}

// Allow 判断是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mtx.Lock()
	defer tb.mtx.Unlock()
	now := time.Now()
	// 计算需要放的令牌数量
	tb.tokens = tb.tokens + tb.rate*now.Sub(tb.lastToken).Seconds()
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	// 判断是否允许请求
	if tb.tokens >= 1 {
		tb.tokens--
		tb.lastToken = now
		return true
	} else {
		return false
	}
}

// IpLimit IP限制, 用于控制API的访问频率
func (l *Middleware) IpLimit(capacity int64, rate float64, msg ...string) gin.HandlerFunc {
	syncTokenMap := sync.Map{}

	return func(c *gin.Context) {
		cliectIP := c.ClientIP()
		if _, ok := syncTokenMap.Load(cliectIP); !ok {
			clentTb := TokenBucket{
				capacity:  capacity,
				rate:      rate,
				tokens:    0,
				lastToken: time.Now(),
			}
			syncTokenMap.Store(cliectIP, &clentTb)
		}

		tb, ok := syncTokenMap.Load(cliectIP)
		if !ok {
			logger.Error("ip limit error, not found token bucket")
			l.resp.Response(c, nil, nil, msg...)
			return
		}

		if !tb.(*TokenBucket).Allow() {
			l.resp.Response(c, nil, nil, msg...)
			return
		}
		c.Next()
	}
}

// Tracing 链路追踪
func (l *Middleware) Tracing(name string) gin.HandlerFunc {
	l.tracing = trace.NewTracerProvider()
	l.propagators = otel.GetTextMapPropagator()
	otel.SetTracerProvider(l.tracing)

	tracer := l.tracing.Tracer(name)
	return func(c *gin.Context) {
		ctx := l.propagators.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(name, c.FullPath(), c.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}
		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		sc := span.SpanContext()
		if sc.HasTraceID() {
			c.Header("trace_id", sc.TraceID().String())
		} else {
			c.Header("trace_id", "not-trace")
		}

		c.Request.Header.Set("trace_id", span.SpanContext().TraceID().String())
		c.Request.Header.Set("span_id", span.SpanContext().SpanID().String())
		c.Next()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
			attribute.String("http.trace_id", span.SpanContext().TraceID().String()),
			attribute.String("http.span_id", span.SpanContext().SpanID().String()),
		)

		// 上报错误
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}

		// 上报状态码
		status := c.Writer.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
	}
}

// Logger 日志
func (l *Middleware) Logger(serverName string, timeLayout ...string) gin.HandlerFunc {
	layout := time.RFC3339
	if len(timeLayout) > 0 {
		layout = timeLayout[0]
	}
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		kv := []zap.Field{
			zap.String("timestamp", time.Now().Format(layout)),
			zap.String("start_time", startTime.Format(layout)),
			zap.String("end_time", endTime.Format(layout)),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.String("req_method", reqMethod),
			zap.String("req_uri", reqUri),
			zap.String("latency_time", latencyTime.String()),
		}

		if l.tracing != nil {
			ctx := l.propagators.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
			span := oteltrace.SpanContextFromContext(ctx)
			traceID := ""
			if span.HasTraceID() {
				traceID = span.TraceID().String()
			}

			spanID := ""
			if span.HasSpanID() {
				spanID = span.SpanID().String()
			}
			kv = append(kv, zap.String("trace_id", traceID), zap.String("span_id", spanID))
		}

		logger.Info(serverName, kv...)
	}
}
