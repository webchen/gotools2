package base

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	// LockedFlag 入锁状态
	LockedFlag int32 = 1

	// UnlockedFlag 未锁状态
	UnlockedFlag int32 = 0
)

// TryMutex 锁对象
type TryMutex struct {
	in     sync.Mutex
	status *int32
}

// NewTryMutex 新建对象
func NewTryMutex() *TryMutex {
	status := UnlockedFlag
	return &TryMutex{
		status: &status,
	}
}

// Lock 加锁
func (m *TryMutex) Lock() {
	m.in.Lock()
}

// Unlock 解锁
func (m *TryMutex) Unlock() {
	m.in.Unlock()
	atomic.AddInt32(m.status, UnlockedFlag)
}

// TryLock 尝试加锁
func (m *TryMutex) TryLock() bool {
	if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.in)), UnlockedFlag, LockedFlag) {
		atomic.AddInt32(m.status, LockedFlag)
		return true
	}
	return false
}

// IsLocked 是否已锁
func (m *TryMutex) IsLocked() bool {
	return atomic.LoadInt32(m.status) == LockedFlag
}
