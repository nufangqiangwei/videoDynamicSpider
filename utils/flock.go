package utils

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Flock struct {
	dir     string
	f       *os.File
	sysType string
}

func New(dir string) *Flock {
	return &Flock{
		dir:     dir,
		sysType: runtime.GOOS,
	}
}

// 加锁
func (l *Flock) Lock() error {
	if l.sysType != "linux" {
		return nil
	}

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
			return fmt.Errorf("cannot flock directory %s - %s", l.dir, err)
		case <-timetick.C:
			err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
			if err == nil {
				return nil
			}
		}
	}
}

// 释放锁
func (l *Flock) Unlock() error {
	if l.sysType != "linux" {
		return nil
	}
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
			flock := New(locked_file)
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
