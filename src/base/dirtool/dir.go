package dirtool

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var configPath = ""

// PathExist ， 判断文件是否存在
func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func PathIsExists(path string) bool {
	b, err := PathExist(path)
	if err != nil {
		return false
	}
	return b
}

// MustCreateDir , 创建文件夹，不返回错误
func MustCreateDir(path string) {
	exist, err := PathExist(path)
	if err != nil {
		log.Fatalln(path, err)
	}
	if !exist {
		os.MkdirAll(path, 0777)
	}
}

// GetCWDPath ，获取项目CWD目录，带 "/"
func GetCWDPath() string {
	pwd, _ := os.Getwd()
	return pwd + string(os.PathSeparator)
}

// GetBasePath ，获取项目的根目录，带 "/"
func GetBasePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic("getBasePath error --> " + err.Error())
	}
	return filepath.Dir(ex) + string(os.PathSeparator)
}

// GetParentDirectory 获取上层目录
func GetParentDirectory(dirctory string) string {
	return dirctory[0:strings.LastIndex(dirctory, string(os.PathSeparator))]
}

// GetConfigPath ，获取项目的配置目录
func GetConfigPath() string {
	if configPath != "" {
		return configPath
	}

	configPath = GetBasePath() + "config" + string(os.PathSeparator)
	if !PathIsExists(configPath) {
		configPath = GetCWDPath() + "config" + string(os.PathSeparator)
	}
	fmt.Println("config path : ", configPath)
	return configPath
}
