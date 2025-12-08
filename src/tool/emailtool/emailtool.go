package emailtool

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cast"
	"github.com/webchen/gotools2/src/base/conf"
	"gopkg.in/gomail.v2"
)

func SendMail(to, name, subject, body, mailType string, files []string) error {
	from := conf.GetConfig("email.from_address", "").(string)
	host := conf.GetConfig("email.host", "").(string)
	port := conf.GetConfig("email.port", "").(string)
	username := conf.GetConfig("email.username", "").(string)
	password := conf.GetConfig("email.password", "").(string)

	if from == "" || host == "" || port == "" || username == "" || password == "" {
		return fmt.Errorf("email 配置缺失")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", name)
	m.SetBody("text/html", strings.ReplaceAll(body, "\n", "<br />"))
	if len(files) > 0 {
		for _, v := range files {
			has, _ := fileExists(v)
			if !has {
				continue
			}
			m.Attach(v)
		}
	}
	d := gomail.NewDialer(host, cast.ToInt(port), username, password)
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Printf("send email result:  %+v\n", err)
	}
	return err
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
