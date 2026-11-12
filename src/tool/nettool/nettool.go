package nettool

import (
	"strings"

	"github.com/toolkits/net"
)

var localIP []string

func init() {
	localIP, _ = net.IntranetIP()
}

// GetLocalIP 获取本机IP
func GetLocalIP() []string {
	return localIP
}

// GetLocalIPStr 获取IP字符串
func GetLocalIPStr() string {
	var sb strings.Builder
	for _, v := range localIP {
		sb.WriteString(v)
		sb.WriteString("_")
	}
	return sb.String()
}

// GetLocalFirstIPStr 获取第1个IP（一台机器可能有多个IP），主要用于记录日志时的服务器IP
func GetLocalFirstIPStr() string {
	for _, v := range localIP {
		return v
	}
	return "127.0.0.1"
}
