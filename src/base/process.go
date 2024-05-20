package base

import (
	"runtime"
)

// IsWIN 是否WIN操作系统
func IsWIN() bool {
	return runtime.GOOS == "windows"
}
