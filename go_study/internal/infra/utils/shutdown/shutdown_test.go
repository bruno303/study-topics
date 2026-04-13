package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

var shutdownTestMu sync.Mutex

func setupShutdownState(t *testing.T) {
	t.Helper()

	shutdownTestMu.Lock()

	ResetForTests()
	signals = []os.Signal{syscall.SIGWINCH}

	t.Cleanup(func() {
		signal.Reset(syscall.SIGWINCH)
		ResetForTests()

		shutdownTestMu.Unlock()
	})
}

func startSignalLoop(t *testing.T) func() {
	t.Helper()

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = syscall.Kill(os.Getpid(), syscall.SIGWINCH)
			}
		}
	}()

	return func() {
		close(stop)
	}
}

func waitForChannel(t *testing.T, ch <-chan struct{}, message string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal(message)
	}
}

func TestCreateListener_WhenRegisteredBeforeShutdown_ExecutesCallbackAndAwaitAllReturns(t *testing.T) {
	setupShutdownState(t)

	executed := make(chan struct{})
	registered := CreateListener(func() {
		close(executed)
	})
	if !registered {
		t.Fatal("expected listener registration to succeed before shutdown")
	}

	ConfigureGracefulShutdown()
	stopSignals := startSignalLoop(t)
	defer stopSignals()

	waitForChannel(t, executed, "expected pre-shutdown listener callback to execute")

	awaited := make(chan struct{})
	go func() {
		AwaitAll()
		close(awaited)
	}()

	waitForChannel(t, awaited, "expected AwaitAll to return after listener callback completion")
}

func TestCreateListener_WhenRegisteredAfterShutdownStart_IsIgnoredAndDoesNotBlockAwaitAll(t *testing.T) {
	setupShutdownState(t)

	started := make(chan struct{})
	release := make(chan struct{})
	lateCalled := make(chan struct{})

	registered := CreateListener(func() {
		close(started)
		<-release
	})
	if !registered {
		t.Fatal("expected initial listener registration to succeed")
	}

	ConfigureGracefulShutdown()
	stopSignals := startSignalLoop(t)
	defer stopSignals()

	waitForChannel(t, started, "expected initial listener callback to start")

	lateRegistered := CreateListener(func() {
		close(lateCalled)
	})
	if lateRegistered {
		t.Fatal("expected late listener registration to be rejected")
	}

	close(release)

	awaited := make(chan struct{})
	go func() {
		AwaitAll()
		close(awaited)
	}()

	waitForChannel(t, awaited, "expected AwaitAll to return without waiting for late listener")

	select {
	case <-lateCalled:
		t.Fatal("expected late listener registration to be ignored")
	default:
	}
}

func TestTrigger_WhenCalledMultipleTimes_ExecutesListenersOnceAndAwaitAllReturns(t *testing.T) {
	setupShutdownState(t)

	called := make(chan struct{}, 2)
	release := make(chan struct{})

	registered := CreateListener(func() {
		called <- struct{}{}
		<-release
	})
	if !registered {
		t.Fatal("expected listener registration to succeed before trigger")
	}

	Trigger()
	waitForChannel(t, called, "expected listener callback to execute after trigger")

	Trigger()

	awaited := make(chan struct{})
	go func() {
		AwaitAll()
		close(awaited)
	}()

	select {
	case <-awaited:
		t.Fatal("expected AwaitAll to wait until triggered listener completes")
	default:
	}

	close(release)
	waitForChannel(t, awaited, "expected AwaitAll to return after triggered listener completion")

	select {
	case <-called:
		t.Fatal("expected listener callback to run only once")
	default:
	}
}
