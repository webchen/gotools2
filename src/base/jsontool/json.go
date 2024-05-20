package jsontool

import (
	"bytes"
	"encoding/json"
	"os"
	//jsoniter "github.com/json-iterator/go"
)

//var json = jsoniter.ConfigCompatibleWithStandardLibrary

// LoadFromFile  从文件里面加载
func LoadFromFile(file string, v interface{}) {
	// 读取JSON文件内容 返回字节切片
	bytes, _ := os.ReadFile(file)
	// 将字节切片映射到指定结构上
	json.Unmarshal(bytes, &v)
}

// LoadFromByte  直接从byte转json对象
func LoadFromByte(bytes []byte, v interface{}) {
	json.Unmarshal(bytes, &v)
}

// LoadFromString 从string转json对象
func LoadFromString(str string, v interface{}) {
	json.Unmarshal([]byte(str), &v)
}

// MarshalToString interface to string
func MarshalToString(v interface{}) string {
	b, _ := json.Marshal(&v)
	return string(b)
}

// JSONStrFormat 格式化JSON字符串
func JSONStrFormat(jsonstr string) string {
	var str bytes.Buffer
	json.Indent(&str, []byte(jsonstr), "", "    ")
	return str.String()
}
