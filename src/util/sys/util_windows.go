package sys

import (
	"os"
	"os/signal"
	"syscall"
)

// ProcessExists  ，判断进程是否在运行
func ProcessExists(pid int) bool {
	if pid == 0 {
		return false
	}
	_, err := os.FindProcess(pid)
	return err == nil
}

// KillProcess  ，调用系统命令，直接KILL
func KillProcess(pid int) {
	//cmdStr := fmt.Sprintf("taskkill.exe -f -im %d", pid)
	//cmdStr := fmt.Sprintf("tskill.exe %d", pid)
	//c := exec.Command(cmdStr)
	/*
	   	c := exec.Command("taskkill", "/F", "/T", "/PID", string(pid))
	   	err := c.Start()
	       fmt.Println(pid, err)
	*/

	pro, err2 := os.FindProcess(pid)
	if err2 == nil {
		pro.Kill()
	}

	return
}

// SignProcess 程序退出控制
func SignProcess(chSign chan uint8) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill:
			//logs.Info("begin to stop ...")
			chSign <- 2
			break
		default:
			break
		}
	}
}
