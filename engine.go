package ginplus

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type (
	// Starter 开始方法的接口
	Starter interface {
		Start() error
	}
	// Stopper 结束方法的接口
	Stopper interface {
		Stop()
	}

	Server interface {
		// Starter 启动服务
		Starter
		// Stopper 停止服务
		Stopper
	}

	// CtrlC 捕获ctrl-c的控制器
	CtrlC struct {
		servers []Server
		// 信号通道
		signalChan chan os.Signal
	}
)

// NewCtrlC 初始化生成CtrlC
func NewCtrlC(services ...Server) *CtrlC {
	return &CtrlC{
		servers:    services,
		signalChan: make(chan os.Signal, 1),
	}
}

// 等待键盘信号
func (c *CtrlC) waitSignals(signals ...os.Signal) {
	signal.Notify(c.signalChan, signals...)
	<-c.signalChan
}

// 接收到kill信号
func (c *CtrlC) waitKill() {
	c.waitSignals(os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
}

func (c *CtrlC) recover() {
	if err := recover(); err != nil {
		fmt.Println(err)
		c.signalChan <- os.Kill
	}
}

// Start 开始运行程序，遇到os.Interrupt停止
func (c *CtrlC) Start() {
	go func() {
		defer c.recover()
		// 启动程序内部的服务列表
		c.startMulServices()
	}()

	c.waitKill()
	c.stopMulServices()
}

// 启动应用子服务
func (c *CtrlC) startMulServices() {
	servicesSlice := c.servers
	for _, service := range servicesSlice {
		go func(s Starter) {
			defer c.recover()
			if err := s.Start(); err != nil {
				panic(err)
			}
		}(service)
	}
}

// 停止应用子服务
func (c *CtrlC) stopMulServices() {
	servicesSlice := c.servers
	for _, service := range servicesSlice {
		service.Stop()
	}
}
