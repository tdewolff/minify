package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher is a wrapper for watching file changes in directories.
type Watcher struct {
	watcher   *fsnotify.Watcher
	paths     map[string]bool
	recursive bool
}

// NewWatcher returns a new Watcher.
func NewWatcher(recursive bool) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{watcher, make(map[string]bool), recursive}, nil
}

// Close closes the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// AddPath adds a new path to watch.
func (w *Watcher) AddPath(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	if info.Mode().IsRegular() {
		root = filepath.Dir(root)
		if w.paths[root] {
			return nil
		}
		if err := w.watcher.Add(root); err != nil {
			return err
		}
		w.paths[root] = true
		return nil
	} else if !w.recursive {
		if w.paths[root] {
			return nil
		}
		if err := w.watcher.Add(root); err != nil {
			return err
		}
		w.paths[root] = true
		return nil
	} else {
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsDir() {
				if !validDir(info) || w.paths[path] {
					return filepath.SkipDir
				}
				if err := w.watcher.Add(path); err != nil {
					return err
				}
				w.paths[path] = true
			}
			return nil
		})
	}
}

// Run watches for file changes.
func (w *Watcher) Run() chan string {
	files := make(chan string, 10)
	go func() {
		changetimes := map[string]time.Time{}
		for w.watcher.Events != nil && w.watcher.Errors != nil {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					w.watcher.Events = nil
					break
				}
				if info, err := os.Stat(event.Name); err == nil {
					if validDir(info) {
						if event.Op&fsnotify.Create == fsnotify.Create {
							if err := w.AddPath(event.Name); err != nil {
								Error.Println(err)
							}
						}
					} else if validFile(info) {
						if event.Op&fsnotify.Write == fsnotify.Write {
							if t, ok := changetimes[event.Name]; !ok || 100*time.Millisecond < time.Now().Sub(t) {
								files <- event.Name
								changetimes[event.Name] = time.Now()
							}
						}
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					w.watcher.Errors = nil
					break
				}
				Error.Println(err)
			}
		}
		close(files)
	}()
	return files
}
