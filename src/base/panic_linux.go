// +build linux darwin

package base

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

var stdErrFileHandler *os.File

func init() {
	RewriteStderrFile()
}

// RewriteStderrFile  重写panic日志输出
func RewriteStderrFile() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	stdErrFile := "/data/syspanic.log"

	fmt.Println("change stdout to : " + stdErrFile)

	file, err := os.OpenFile(stdErrFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return err
	}
	stdErrFileHandler = file //把文件句柄保存到全局变量，避免被GC回收

	if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		fmt.Println(err)
		return err
	}
	// 内存回收前关闭文件描述符
	runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
		fd.Close()
	})

	return nil
}
