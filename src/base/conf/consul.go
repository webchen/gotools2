package conf

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/webchen/gotools2/src/tool/nettool"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/cast"
)

var consulClient *consulapi.Client

func init() {
	loadBaseConfig()
	initConsulClient()
}

func checkConsul() bool {
	/*
		if consulClient == nil {
			log.Println("初始化consul失败，不更新配置和注册")
			return false
		}
		return true
	*/
	return consulClient != nil
}

func initConsulClient() {
	if !checkBaseConfigData() {
		return
	}
	var err error
	if baseConfigData["consul"] == nil {
		return
	}
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Token = baseConfigData["consul"]["token"].(string)
	serverList := baseConfigData["consul"]["server"].([]interface{})

	consulConfig.Address = serverList[rand.Intn(len(serverList))].(string)
	consulClient, err = consulapi.NewClient(consulConfig)
	if err != nil {
		log.Println("consul error : ", err)
	}
}

func ConsulRegister() {
	if !checkConsul() {
		return
	}
	cnf := baseConfigData["consul"]

	if cnf["register"].(map[string]interface{})["open"].(string) != "1" {
		return
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = cnf["register"].(map[string]interface{})["id"].(string) + nettool.GetLocalFirstIPStr() // 服务节点的名称
	registration.Name = cnf["register"].(map[string]interface{})["name"].(string)                            // 服务名称
	registration.Port = cast.ToInt(cnf["register"].(map[string]interface{})["port"])                         // 服务端口
	registration.Tags = cnf["register"].(map[string]interface{})["tags"].([]string)                          // tag，可以为空
	registration.Address = nettool.GetLocalFirstIPStr()                                                      // 服务 IP

	if cnf["check"].(map[string]interface{})["open"].(string) == "1" {
		checkPort := cast.ToInt(cnf["check"].(map[string]interface{})["port"])
		registration.Check = &consulapi.AgentServiceCheck{ // 健康检查
			HTTP:                           fmt.Sprintf("http://%s:%d%s", registration.Address, checkPort, "/"),
			Timeout:                        cnf["check"].(map[string]interface{})["timeout"].(string),
			Interval:                       cnf["check"].(map[string]interface{})["interval"].(string),                       // 健康检查间隔
			DeregisterCriticalServiceAfter: cnf["check"].(map[string]interface{})["deregisterCriticalServiceAfter"].(string), //check失败后30秒删除本服务，注销时间，相当于过期时间
			// GRPC:     fmt.Sprintf("%v:%v/%v", IP, r.Port, r.Service),// grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
		}
	}
	err := consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Println("register server error : ", err)
	}
}
