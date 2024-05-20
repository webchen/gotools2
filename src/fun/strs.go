package fun

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// Md5 32位MD5
func Md5(strs string) string {
	w := md5.New()
	io.WriteString(w, strs)
	//将str写入到w中
	return fmt.Sprintf("%x", w.Sum(nil))
}

func Case2Cams(s string) string {
	r, _ := regexp.Compile(`_[0-9a-zA-Z]`)
	rep := r.ReplaceAllStringFunc(s, strings.ToUpper)
	rep = strings.ReplaceAll(rep, "_", "")
	return rep
}

func MapCase2Cams(m map[string]string) map[string]string {
	mm := make(map[string]string)
	for k, v := range m {
		mm[Case2Cams(k)] = v
	}
	return mm
}

// RandString 随机字符串
func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		b := r.Intn(25) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
