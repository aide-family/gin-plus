package ginplus

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	resp IResponser
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
