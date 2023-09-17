//go:build windows

package utils

import (
	"os"
	"runtime"
)

type Flock struct {
	dir     string
	f       *os.File
	sysType string
}

func NewFlock(dir string) *Flock {
	return &Flock{
		dir:     dir,
		sysType: runtime.GOOS,
	}
}

// 加锁
func (l *Flock) Lock() error {
	return nil
}

// 释放锁
func (l *Flock) Unlock() error {
	return nil
}
