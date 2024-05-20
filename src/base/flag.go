package base

import (
	"flag"
	"strings"
)

var checked = false
var buildOs = ""
var daemon = 1
var buildWithConfig = 1

// IsBuild 是否编译
func IsBuild() bool {
	return strings.TrimSpace(BuildOsName()) != ""
}

// BuildOsName 要编译的系统名称
func BuildOsName() string {
	if checked {
		return buildOs
	}
	checkFlags()
	return buildOs
}

func IsDaemon() bool {
	if checked {
		return daemon == 1
	}
	checkFlags()
	return daemon == 1
}

func BuildWithConfig() bool {
	return buildWithConfig == 1
}

func checkFlags() {
	if checked {
		return
	}
	flag.StringVar(&buildOs, "buildOs", "", "1) linux (default) \n 2) windows \n 3) mac \n 4) freebsd")
	flag.IntVar(&buildWithConfig, "buildWithConfig", 1, "default : 1. if no cover config, 0 is ok")
	flag.IntVar(&daemon, "daemon", 0, "1: daemon")

	checked = true
	flag.Parse()
}
