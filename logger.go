package ginplus

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger     = zap.NewExample()
	loggerOnce = sync.Once{}
)

// SetLogger 设置日志记录器
func SetLogger(l *zap.Logger) {
	loggerOnce.Do(func() {
		logger = l
	})
}

// Logger 返回日志记录器
func Logger() *zap.Logger {
	return logger
}
