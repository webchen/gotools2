package sys

import (
	"os"
	"runtime"
	"time"
)

// Program 程序本身的一些信息
type Program struct {
	os         string
	start      time.Time //	开始运行时间
	pid        int       // 进程id
	memoryUsed uint64
	routineNum int
}

// NewProgram 初始化
func NewProgram() *Program {
	return &Program{
		os:         runtime.GOOS,
		start:      time.Now(),
		pid:        os.Getpid(),
		memoryUsed: 0,
		routineNum: 0,
	}
}

// GetRoNum 获取总的协程数
func (p *Program) GetRoNum() int {
	p.routineNum = runtime.NumGoroutine()
	return p.routineNum
}

// GetMemoryUsed 获取总的内存使用量
func (p *Program) GetMemoryUsed() uint64 {
	status := &runtime.MemStats{}
	runtime.ReadMemStats(status)
	p.memoryUsed = status.Sys
	return p.memoryUsed
}

// GetOSName 获取操作系统的名字
func (p *Program) GetOSName() string {
	return p.os
}

// GetStartTime 获取开始运行的时间
func (p *Program) GetStartTime() time.Time {
	return p.start
}
