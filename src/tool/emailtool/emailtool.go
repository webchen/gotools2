package emailtool

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/webchen/gotools2/src/base/conf"
)

var username = conf.GetConfig("email.sender", "altidc@qq.com").(string)
var password = conf.GetConfig("email.password", "111111").(string)
var host = conf.GetConfig("email.host", "smtp.qq.com:25").(string)
var to = conf.GetConfig("email.to", "altidc@qq.com").(string)

/*
SendMail 发送邮件

username 发送者邮件
password 授权码
host 主机地址 smtp.qq.com:587 或 smtp.qq.com:25
to 接收邮箱 多个接收邮箱使用 ; 隔开
name 发送人名称
subject 发送主题
body 发送内容
mailType 发送邮件内容类型
*/
func SendMail(to, name, subject, body, mailType string) error {
	//return nil

	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", username, password, hp[0])
	var contentType string
	if mailType == "html" {
		contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + name + "<" + username + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, username, sendTo, msg)
	fmt.Printf("send email result:  %+v", err)
	return err

}

// SendAlertEmail 发送报警邮件
func SendAlertEmail(body string) {
	SendMail(to, "网关报警", "网关报警", body, "")
}

// SendNormalEmail 普通邮件
func SendNormalEmail(title, body string) {
	SendMail(to, title, "网关信息", body, "")
}
