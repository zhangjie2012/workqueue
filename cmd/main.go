package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/zhangjie2012/workqueue/pkg/workqueue"
)

func PrintObj(obj interface{}) error {
	str, ok := obj.(string)

	if ok {
		fmt.Printf("string -- %s\n", str)
	} else {
		fmt.Printf("other type -- %v\n", obj)
	}

	return nil
}

func isClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func main() {
	wq := workqueue.NewWorkQueue(PrintObj)
	wq.Run()

	stopCh := make(chan bool)
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 100; i++ {
			// 字符串、整型
			v := r.Int31() % 2
			if v == 1 {
				wq.Enqueue(fmt.Sprintf("I'm string -> %d", v))
			} else {
				wq.Enqueue(v)
			}

			time.Sleep(10 * time.Millisecond)
		}
		stopCh <- true
	}()

	<-stopCh

	wq.ShutDown()
}
