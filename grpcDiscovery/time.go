package grpcDiscovery

import (
	"sync"
	"time"
)

type TimeLock struct {
	t        sync.Map
	TimeTick int64
}

func (t *TimeLock) GetGroupNameLock(groupName string) bool {
	v, ok := t.t.Load(groupName)
	if !ok {
		return false
	}
	return t.TimeTick+v.(int64) < time.Now().Unix()
}
func (t *TimeLock) LockByGroupName(groupName string) bool {
	v, ok := t.t.Load(groupName)
	if ok {
		if t.TimeTick+v.(int64) > time.Now().Unix() {
			// 尚未超时
			return false
		} else {
			// 已经超时了,重新加锁
			t.t.LoadOrStore(groupName, time.Now().Unix())
			return true
		}
	}

	t.t.Store(groupName, time.Now().Unix())
	return true
}
func (t *TimeLock) UnLock(groupName string) {
	t.t.Delete(groupName)
}
