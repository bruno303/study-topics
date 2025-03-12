package fs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func StartMonitor(filesToWatch []string, delay int, cb func() error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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
						fmt.Println("File modified:", event.Name)

						if delay > 0 {
							fmt.Printf("Waiting %d seconds before re-executing...\n", delay)
							time.Sleep(time.Duration(delay) * time.Second)
						}

						err := cb()
						if err != nil {
							log.Println("Error running command:", err)
							os.Exit(1)
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	files := make([]string, 0)
	for _, pattern := range filesToWatch {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Fatalf("Error parsing pattern %s: %v", pattern, err)
		}
		if len(matches) == 0 {
			log.Printf("Warning: No files match the pattern %s", pattern)
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
					log.Fatalf("Error walking directory %s: %v", match, err)
				}
			} else {
				files = append(files, match)
			}
		}
	}

	for _, file := range files {
		err := watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Watching file: %s\n", files)
	<-done
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
