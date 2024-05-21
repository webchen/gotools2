package fun

import (
	"regexp"
	"strings"
)

// 检查s是否空字符串，否则返回d
func CheckAndGetEmptyString(s string, d string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return d
	}
	return s
}

// 检查手机号是否合法
func CheckPhone(phone string) bool {
	p := `^1[3456789]\d{9}$`
	reg := regexp.MustCompile(p)
	return reg.MatchString(phone)
}

// 检查email是否合法
func CheckEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+.[a-zA-Z]{2,}$`, email)
	return matched
}

// 检查整数
func CheckIntger(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}
