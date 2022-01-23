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
	dirs      map[string]bool
	paths     map[string]bool
	recursive bool
}

// NewWatcher returns a new Watcher.
func NewWatcher(recursive bool) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{watcher, map[string]bool{}, map[string]bool{}, recursive}, nil
}

// Close closes the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// AddPath adds a new path to watch.
func (w *Watcher) AddPath(root string) error {
	w.paths[root] = true

	info, err := os.Lstat(root)
	if err != nil {
		return err
	}

	if info.Mode().IsRegular() {
		root = filepath.Dir(root)
		if w.dirs[root] {
			return nil
		}
		if err := w.watcher.Add(root); err != nil {
			return err
		}
		w.dirs[root] = true
	} else if info.Mode().IsDir() && w.recursive {
		return WalkDir(DirFS("."), filepath.Clean(root), func(path string, d DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if w.dirs[path] {
					return SkipDir
				}
				if err := w.watcher.Add(path); err != nil {
					return err
				}
				w.dirs[path] = true
			}
			return nil
		})
	}
	return nil
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

				// check if changed file is being watched (as a file or indirectly in a dir)
				watched := false
				for path := range w.paths {
					if IsDir(path) {
						// file in w.paths
						if path == filepath.Clean(event.Name) {
							watched = true
							break
						}
					} else if _, err := filepath.Rel(path, event.Name); err == nil {
						// dir in w.paths
						watched = true
						break
					}
				}
				if !watched {
					break
				}

				if info, err := os.Lstat(event.Name); err == nil {
					if info.Mode().IsDir() && w.recursive {
						if event.Op&fsnotify.Create == fsnotify.Create {
							if err := w.AddPath(event.Name); err != nil {
								Error.Println(err)
							}
						}
					} else if info.Mode().IsRegular() {
						if event.Op&fsnotify.Write == fsnotify.Write {
							if t, ok := changetimes[event.Name]; !ok || 100*time.Millisecond < time.Since(t) {
								time.Sleep(100 * time.Millisecond) // wait to make sure write is finished
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
