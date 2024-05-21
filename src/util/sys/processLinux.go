//go:build !windows
// +build !windows

package sys

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/webchen/gotools/help/logs"
)

// ProcessExists  ，判断进程是否在运行
func ProcessExists(pid int) bool {
	if pid == 0 {
		return false
	}
	killErr := syscall.Kill(pid, syscall.Signal(0))
	return killErr == nil
}

// KillProcess ，KILL进程
func KillProcess(pid int) {
	syscall.Kill(-pid, syscall.SIGKILL)
}

// SignProcess 程序退出控制
func SignProcess(chSign chan uint8) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill:
			logs.Info("begin to stop ...", nil)
			chSign <- 2
			break
		case syscall.SIGUSR2:
			chSign <- 1
			logs.Info("begin to stop ...", nil)
			break
		default:
			break
		}
	}
}
