package base

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// rotateWriter 日志文件写入器：写满 maxSize 后把当前文件移到 back 目录并重开新文件。
// 写入与轮转共用一把锁，避免"轮转时关闭文件、写入还在写"导致的丢行。
type rotateWriter struct {
	mu       sync.Mutex
	f        *os.File
	path     string // <logdir>/<name>.log
	name     string // <name>
	ext      string // ".log"
	backDir  string // <logdir>/back
	maxSize  int64  // 单文件上限（字节），<=0 不轮转
	maxCount int    // back 目录保留份数，0 表示不留备份
	size     int64  // 当前文件已写字节
}

// newRotateWriter 打开日志文件并初始化写入器
func newRotateWriter(path string, maxSize int64, maxCount int) (*rotateWriter, error) {
	backDir := filepath.Join(filepath.Dir(path), "back")
	if err := os.MkdirAll(backDir, 0777); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return nil, err
	}
	var size int64
	if st, err := f.Stat(); err == nil {
		size = st.Size()
	}
	ext := filepath.Ext(path)
	return &rotateWriter{
		f:        f,
		path:     path,
		name:     strings.TrimSuffix(filepath.Base(path), ext),
		ext:      ext,
		backDir:  backDir,
		maxSize:  maxSize,
		maxCount: maxCount,
		size:     size,
	}, nil
}

// Write 实现 io.Writer：写入前判阈值，超出则先轮转
func (w *rotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		if err := w.reopenLocked(); err != nil {
			return 0, err
		}
	}
	if w.maxSize > 0 && w.size > 0 && w.size+int64(len(p)) > w.maxSize {
		w.rotateLocked()
	}
	if w.f == nil {
		return 0, os.ErrClosed
	}
	n, err := w.f.Write(p)
	w.size += int64(n)
	return n, err
}

// Close 关闭当前文件
func (w *rotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

// reopenLocked 重新打开当前日志文件（调用方需持锁）
func (w *rotateWriter) reopenLocked() error {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	w.f = f
	if st, err := f.Stat(); err == nil {
		w.size = st.Size()
	}
	return nil
}

// rotateLocked 把当前文件移到 back 目录、保留 maxCount 份、重开新文件（调用方需持锁）
func (w *rotateWriter) rotateLocked() {
	if w.f != nil {
		w.f.Close()
		w.f = nil
	}
	ts := time.Now().Format("20060102150405")
	target := filepath.Join(w.backDir, w.name+"_"+ts+w.ext)
	for i := 2; pathIsExists(target); i++ {
		target = filepath.Join(w.backDir, w.name+"_"+ts+"_"+strconv.Itoa(i)+w.ext)
	}
	if err := os.Rename(w.path, target); err != nil {
		// 移走失败：原文件仍在，重新打开继续追加，不做清理
		_ = w.reopenLocked()
		return
	}
	w.size = 0
	_ = w.reopenLocked()
	w.pruneLocked()
}

// pruneLocked 按文件名里的时间戳保留最新的 maxCount 份备份（调用方需持锁）
func (w *rotateWriter) pruneLocked() {
	entries, err := os.ReadDir(w.backDir)
	if err != nil {
		return
	}
	prefix := w.name + "_"
	var backups []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, prefix) && strings.HasSuffix(n, w.ext) {
			backups = append(backups, n)
		}
	}
	// 文件名以 14 位时间戳开头，字典序即时间顺序
	sort.Strings(backups)
	keep := w.maxCount
	if keep < 0 {
		keep = 0
	}
	for len(backups) > keep {
		os.Remove(filepath.Join(w.backDir, backups[0]))
		backups = backups[1:]
	}
}

// pathIsExists 文件是否存在
func pathIsExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
