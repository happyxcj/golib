package lib

import (
	"sync"
	"time"
)

var globalTimerPool sync.Pool

// 从对象池获取指定超时时间的定时器.
func acquireTimer(timeout time.Duration) *time.Timer {
	v := globalTimerPool.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t := v.(*time.Timer)
	t.Reset(timeout)
	return t
}

// 释放定时器资源到对象池
func releaseTimer(t *time.Timer) {
	if !t.Stop() {
		// 定时器如果还未触发则清除.
		select {
		case <-t.C:
		default:
		}
	}
	globalTimerPool.Put(t)
}
