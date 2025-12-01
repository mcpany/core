package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches configuration files for changes and sends a signal to reload.
type Watcher struct {
	watcher    *fsnotify.Watcher
	reloadChan chan<- struct{}
	done       chan struct{}
}

// NewWatcher creates a new configuration watcher.
func NewWatcher(reloadChan chan<- struct{}) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		watcher:    watcher,
		reloadChan: reloadChan,
		done:       make(chan struct{}),
	}, nil
}

// Watch starts watching the given configuration files.
func (w *Watcher) Watch(paths []string) error {
	for _, path := range paths {
		if err := w.watcher.Add(path); err != nil {
			return err
		}
	}
	go w.run()
	return nil
}

// Close stops the watcher.
func (w *Watcher) Close() {
	close(w.done)
	w.watcher.Close()
}

func (w *Watcher) run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("Configuration file modified: %s", event.Name)
				w.reloadChan <- struct{}{}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching config file: %v", err)
		case <-w.done:
			return
		}
	}
}
