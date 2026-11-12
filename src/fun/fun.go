package fun

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/webchen/gotools2/src/base/dirtool"
	"github.com/webchen/gotools2/src/util/logs"
	"github.com/webchen/gotools2/src/util/model2"
)

func HttpPanic(message string, data interface{}) {
	obj := &model2.PanicMessage{
		Message: message,
		Data:    data,
	}
	panic(obj)
}

var reqSeq atomic.Uint64

// GetReqSeqId 返回进程内唯一的请求序号：秒级时间 + 严格递增序号。
func GetReqSeqId() string {
	n := reqSeq.Add(1)
	return time.Now().Format("20060102150405") + strconv.FormatUint(n, 10)
}

// 目录打包到ZIP
func DirToZip(directory string) string {
	n := dirtool.GetCWDPath() + "tmp" + string(os.PathSeparator) + GetReqSeqId() + ".zip"
	zipWriter, err := os.Create(n)
	if err != nil {
		logs.Error("DirToZip create zip file fail", err)
		return ""
	}
	defer zipWriter.Close()
	archive := zip.NewWriter(zipWriter)
	defer archive.Close()
	walkErr := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name, err = filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}
		headerWriter, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		// 不 defer：Walk 结束前及时释放句柄，避免大目录耗尽 fd
		_, err = io.Copy(headerWriter, f)
		f.Close()
		return err
	})
	if walkErr != nil {
		logs.Error("DirToZip walk fail", walkErr)
	}
	return n
}

var uEscSeq = []byte{'\\', 'u'}

// 中文unicode转中文字符串
// 把 body 里字面的 JSON 转义 \uXXXX 还原成 UTF-8 字符（只为日志可读）。
// 不含 \u 时直接返回原切片（零拷贝）；转义非法或不足 4 位时原样保留，不丢内容、不篡改原文。
func ZhToUnicode(raw []byte) ([]byte, error) {
	if !bytes.Contains(raw, uEscSeq) {
		return raw, nil
	}

	out := make([]byte, 0, len(raw))
	changed := false

	for i := 0; i < len(raw); {
		// 普通字节
		if raw[i] != '\\' {
			out = append(out, raw[i])
			i++
			continue
		}

		// 转义反斜杠：\\ 成对原样输出，避免把字面的 \\u4e2d 当成 unicode 转义
		if i+1 < len(raw) && raw[i+1] == '\\' {
			out = append(out, raw[i], raw[i+1])
			i += 2
			continue
		}

		// \uXXXX
		if i+5 < len(raw) && raw[i+1] == 'u' {
			if r, ok := hex4(raw[i+2 : i+6]); ok {
				// 代理对：\uD83D\uDE00 合并成一个字符
				if utf16.IsSurrogate(r) {
					if i+11 < len(raw) && raw[i+6] == '\\' && raw[i+7] == 'u' {
						if r2, ok2 := hex4(raw[i+8 : i+12]); ok2 {
							if merged := utf16.DecodeRune(r, r2); merged != utf8.RuneError {
								out = utf8.AppendRune(out, merged)
								changed = true
								i += 12
								continue
							}
						}
					}
					// 落单的代理项：原样保留，不输出非法 UTF-8
					out = append(out, raw[i:i+6]...)
					i += 6
					continue
				}

				out = utf8.AppendRune(out, r)
				changed = true
				i += 6
				continue
			}
		}

		// 其它情况（\n、\"、非法 \u 等）：只吐当前字节，其余按普通字节继续
		out = append(out, raw[i])
		i++
	}

	if !changed {
		return raw, nil
	}
	return out, nil
}

// hex4 解析 4 位十六进制，含非法字符返回 false
func hex4(b []byte) (rune, bool) {
	var r rune
	for i := 0; i < 4; i++ {
		var d rune
		switch c := b[i]; {
		case c >= '0' && c <= '9':
			d = rune(c - '0')
		case c >= 'a' && c <= 'f':
			d = rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = rune(c-'A') + 10
		default:
			return 0, false
		}
		r = r<<4 | d
	}
	return r, true
}
