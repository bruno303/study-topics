package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/fsnotify/fsnotify"
)

type Monitor struct {
	FilesToWatch []string
	Delay        int
	Callback     func() error
}

func NewMonitor(filesToWatch []string, delay int, callback func() error) *Monitor {
	return &Monitor{
		FilesToWatch: filesToWatch,
		Delay:        delay,
		Callback:     callback,
	}
}

func (m *Monitor) Start(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Log().Error(ctx, "Error creating file watcher:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		var (
			debounceTimer *time.Timer
			debounceWait  = 500 * time.Millisecond
		)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceTimer != nil {
						debounceTimer.Stop()
					}

					debounceTimer = time.AfterFunc(debounceWait, func() {
						log.Log().Info(ctx, "File modified:", event.Name)

						if m.Delay > 0 {
							log.Log().Debug(ctx, "Waiting %d seconds before re-executing...\n", m.Delay)
							time.Sleep(time.Duration(m.Delay) * time.Second)
						}

						err := m.Callback()
						if err != nil {
							log.Log().Error(ctx, "Error running command:", err)
							os.Exit(1)
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Log().Error(ctx, "Error:", err)
			}
		}
	}()

	files := make([]string, 0)
	for _, pattern := range m.FilesToWatch {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Log().Error(ctx, fmt.Sprintf("Error parsing pattern %s", pattern), err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			log.Log().Warn(ctx, "Warning: No files match the pattern %s", pattern)
		}
		for _, match := range matches {
			// If it's a directory, recursively add all files
			if isDir(match) {
				err := filepath.Walk(match, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						files = append(files, path)
					}
					return nil
				})
				if err != nil {
					log.Log().Error(ctx, fmt.Sprintf("Error walking directory %s", match), err)
					os.Exit(1)
				}
			} else {
				files = append(files, match)
			}
		}
	}

	for _, file := range files {
		err := watcher.Add(file)
		if err != nil {
			log.Log().Error(ctx, "Error watching file", err)
			os.Exit(1)
		}
	}

	log.Log().Info(ctx, "Watching file: %s\n", files)
	<-done
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
