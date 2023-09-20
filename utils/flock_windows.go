//go:build windows

package utils

import (
	"os"
	"syscall"
)

type Flock struct {
	dir string
	f   *os.File
}

func NewFlock(dir string) *Flock {
	return &Flock{
		dir: dir,
	}
}

// 加锁
func (l *Flock) Lock() error {
	syscall.ForkLock.Lock()
	return nil
}

// 释放锁
func (l *Flock) Unlock() error {
	syscall.ForkLock.Unlock()
	return nil
}
