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
func (w *Watcher) Watch(paths []string, reloadFunc func()) {
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
		err := w.watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-w.done
}

// Close stops the file watcher.
func (w *Watcher) Close() {
	close(w.done)
	w.watcher.Close()
}
