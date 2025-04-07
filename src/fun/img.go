package fun

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/webchen/gotools2/src/base/dirtool"

	"github.com/nfnt/resize"
)

func SaveFileToLocal(url, save_path string) string {
	ss := HttpGet(url)
	d := dirtool.GetCWDPath() + "tmp" + string(os.PathSeparator) + save_path
	dirtool.MustCreateDir(d)
	p := d + "/" + filepath.Base(url)
	os.Remove(p) // 不管是否存在，都直接删
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	ext := strings.ToLower(path.Ext(p))
	nss := ""
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		nss = string(CompressImageResource([]byte(ss), ext))
	}
	_, err := io.WriteString(f, nss)
	if err != nil {
		return ""
	}
	return p
}

// 压缩图片，控制在2M以内，仅仅支持 jpg/jpeg/png
func CompressImageResource(data []byte, ext string) []byte {
	if len(data) < 1024*1024*2 {
		return data
	}
	var img image.Image
	var err error
	// 很奇怪，直接用 image.Decode，png会报格式错误，所以分开判断
	if ext == ".png" {
		img, err = png.Decode(bytes.NewReader(data))
	} else {
		img, err = jpeg.Decode(bytes.NewReader(data))
	}
	if err != nil {
		return data
	}

	// 修改图片的大小
	m := resize.Resize(0, 0, img, resize.Lanczos3)
	buf := bytes.Buffer{}

	// 修改图片的质量
	err = jpeg.Encode(&buf, m, &jpeg.Options{Quality: 80})
	if err != nil {
		return data
	}
	if buf.Len() > len(data) {
		return data
	}
	return buf.Bytes()
}
