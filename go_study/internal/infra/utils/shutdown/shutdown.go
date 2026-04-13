package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	wg                    = &sync.WaitGroup{}
	callbacksMu           = &sync.Mutex{}
	callbacks    []func() = make([]func(), 0)
	shuttingDown          = false
	signals               = defaultSignals()
)

func defaultSignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT}
}

func ConfigureGracefulShutdown() {
	go func() {
		exitChan := make(chan os.Signal, 1)
		signal.Notify(exitChan, signals...)
		defer signal.Stop(exitChan)

		<-exitChan
		Trigger()
	}()
}

func Trigger() {
	callbacksMu.Lock()
	if shuttingDown {
		callbacksMu.Unlock()
		return
	}

	shuttingDown = true
	registeredCallbacks := append([]func(){}, callbacks...)
	callbacksMu.Unlock()

	for _, f := range registeredCallbacks {
		go f()
	}
}

func CreateListener(f func()) bool {
	callbacksMu.Lock()
	defer callbacksMu.Unlock()

	if shuttingDown {
		return false
	}

	wg.Add(1)
	callbacks = append(callbacks, func() {
		f()
		wg.Done()
	})

	return true
}

func AwaitAll() {
	wg.Wait()
}

func ResetForTests() {
	callbacksMu.Lock()
	defer callbacksMu.Unlock()

	callbacks = make([]func(), 0)
	shuttingDown = false
	signals = defaultSignals()
	wg = &sync.WaitGroup{}
}
