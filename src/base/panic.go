package base

import (
	"fmt"
	"reflect"
)

// HandlePanic 处理panic
func HandlePanic() {
	if err := recover(); err != nil {
		LogPanicErr(err, "panic : "+fmt.Sprintf("%s", err))
	}
}

// Go 自定义的GO函数，捕获panic
func Go(cb interface{}, args ...interface{}) {
	f := reflect.ValueOf(cb)
	go func() {
		defer HandlePanic()
		n := len(args)
		if n > 0 {
			refargs := make([]reflect.Value, n)
			for i := 0; i < n; i++ {
				refargs[i] = reflect.ValueOf(args[i])
			}
			f.Call(refargs)
		} else {
			f.Call(nil)
		}
	}()
}
