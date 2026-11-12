package base

const (
	// LockedFlag 入锁状态
	LockedFlag int32 = 1

	// UnlockedFlag 未锁状态
	UnlockedFlag int32 = 0
)

// TryMutex 支持非阻塞尝试的互斥锁。
// 用容量 1 的 channel 表示令牌：持有者取走令牌，释放时归还；
// 不再对 sync.Mutex 内部状态做 unsafe CAS，状态与锁本身始终一致。
type TryMutex struct {
	ch chan struct{}
}

// NewTryMutex 新建对象
func NewTryMutex() *TryMutex {
	m := &TryMutex{ch: make(chan struct{}, 1)}
	m.ch <- struct{}{}
	return m
}

// Lock 加锁（阻塞）
func (m *TryMutex) Lock() {
	<-m.ch
}

// Unlock 解锁；未持锁时调用会 panic，与 sync.Mutex 语义一致
func (m *TryMutex) Unlock() {
	select {
	case m.ch <- struct{}{}:
	default:
		panic("sync: unlock of unlocked try mutex")
	}
}

// TryLock 尝试加锁，成功返回 true，已被占用返回 false
func (m *TryMutex) TryLock() bool {
	select {
	case <-m.ch:
		return true
	default:
		return false
	}
}

// IsLocked 是否已被持有
func (m *TryMutex) IsLocked() bool {
	return len(m.ch) == 0
}
