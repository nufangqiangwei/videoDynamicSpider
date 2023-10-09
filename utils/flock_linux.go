//go:build linux

package utils

import (
	"fmt"
	"os"
	"syscall"
	"time"
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

// Lock 加锁
func (l *Flock) Lock() error {
	f, err := os.Open(l.dir)
	if err != nil {
		return err
	}
	l.f = f

	timeOut := time.After(time.Minute)
	timetick := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-timeOut:
			timetick.Stop()
			return fmt.Errorf("cannot flock directory %s - %s", l.dir, err)
		case <-timetick.C:
			err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			if err == nil {
				timetick.Stop()
				return nil
			}
		}
	}
}

// Unlock 释放锁
func (l *Flock) Unlock() error {
	defer l.f.Close()
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
