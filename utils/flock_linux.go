//go:build linux
// +build linux

package utils

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
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

// 加锁
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
				//go func() {
				//	c := time.NewTicker(time.Minute)
				//	select {
				//	case <-c.C:
				//		l.Unlock()
				//	}
				//	c.Stop()
				//}()
				return nil
			}
		}
	}
}

// 释放锁
func (l *Flock) Unlock() error {
	defer l.f.Close()
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}

func main() {
	test_file_path, _ := os.Getwd()
	locked_file := test_file_path
	wg := sync.WaitGroup{}
	rand.Seed(time.Now().Unix())
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(num int) {
			flock := NewFlock(locked_file)
			err := flock.Lock()
			if err != nil {
				wg.Done()
				fmt.Println(err.Error())
				return
			}
			sleepTime := rand.Intn(20)
			fmt.Printf("休眠%d秒\n", sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
			fmt.Printf("output : %d\n", num)
			flock.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
	time.Sleep(2 * time.Second)
}
