package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	wg                 = &sync.WaitGroup{}
	callbacks []func() = make([]func(), 0)
)

func ConfigureGracefulShutdown() {
	go func() {
		exitChan := make(chan os.Signal, 1)
		signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		<-exitChan
		close(exitChan)
		for _, f := range callbacks {
			go f()
		}
	}()
}

func CreateListener(f func()) {
	wg.Add(1)
	callbacks = append(callbacks, func() {
		f()
		wg.Done()
	})
}

func AwaitAll() {
	wg.Wait()
}
