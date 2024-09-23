package fun

import (
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/webchen/gotools2/src/base/dirtool"
	"github.com/webchen/gotools2/src/util/model2"
)

func HttpPanic(message string, data interface{}) {
	obj := &model2.PanicMessage{
		Message: message,
		Data:    data,
	}
	panic(obj)
}

// 生成随机数
func GetReqSeqId() string {
	strTime := time.Now().Format("20060102150405")
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	num := fmt.Sprintf("%d", r1.Intn(1000)+1000)
	return strTime + num
}

// 目录打包到ZIP
func DirToZip(directory string) string {
	n := dirtool.GetCWDPath() + "tmp" + string(os.PathSeparator) + GetReqSeqId() + ".zip"
	zipWriter, _ := os.Create(n)
	defer zipWriter.Close()
	archive := zip.NewWriter(zipWriter)
	defer archive.Close()
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, _ := zip.FileInfoHeader(info)
		header.Method = zip.Deflate
		header.Name, _ = filepath.Rel(directory, path)
		if info.IsDir() {
			header.Name += "/"
		}
		headerWriter, _ := archive.CreateHeader(header)
		if info.IsDir() {
			return nil
		}
		f, _ := os.Open(path)
		defer f.Close()
		_, _ = io.Copy(headerWriter, f)
		return nil
	})
	return n
}

// 中文unicode转中文字符串
func ZhToUnicode(raw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}
