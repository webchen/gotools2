package fun

import (
	"net"
	"net/http"
)

// RemoteIP 获取客户端IP
func RemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("XRealIP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("XForwardedFor"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}
