package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg = &sync.WaitGroup{}

func CreateListener(f func()) {
	wg.Add(1)
	go func() {
		exitChan := make(chan os.Signal, 1)
		signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		<-exitChan
		f()
		close(exitChan)
		wg.Done()
	}()
}

func AwaitAll() {
	wg.Wait()
}
