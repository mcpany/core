package config

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors configuration files for changes and triggers a reload.
type Watcher struct {
	watcher *fsnotify.Watcher
	done    chan bool
	mu      sync.Mutex
	timer   *time.Timer
}

// NewWatcher creates a new file watcher.
//
// Returns:
//   - A pointer to a new Watcher.
//   - An error if the watcher creation fails.
func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher: watcher,
		done:    make(chan bool),
	}, nil
}

// Watch starts monitoring the specified configuration paths.
//
// Parameters:
//   - paths: A slice of file or directory paths to watch.
//   - reloadFunc: The function to call when a change is detected.
//
// Returns:
//   - An error if watching fails.
func (w *Watcher) Watch(paths []string, reloadFunc func()) error {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					w.mu.Lock()
					if w.timer != nil {
						w.timer.Stop()
					}
					w.timer = time.AfterFunc(1*time.Second, func() {
						log.Println("Reloading configuration...")
						reloadFunc()
					})
					w.mu.Unlock()
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-w.done:
				return
			}
		}
	}()

	for _, path := range paths {
		if isURL(path) {
			continue
		}
		err := w.watcher.Add(path)
		if err != nil {
			return err
		}
	}

	<-w.done
	return nil
}

// Close stops the file watcher and releases resources.
func (w *Watcher) Close() {
	close(w.done)
	_ = w.watcher.Close()
}
