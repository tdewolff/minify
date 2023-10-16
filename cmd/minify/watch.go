package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher is a wrapper for watching file changes in directories.
type Watcher struct {
	watcher    *fsnotify.Watcher
	dirs       map[string]bool
	paths      map[string]bool
	ignoreNext map[string]bool
	recursive  bool
}

// NewWatcher returns a new Watcher.
func NewWatcher(recursive bool) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		watcher:    watcher,
		dirs:       map[string]bool{},
		paths:      map[string]bool{},
		ignoreNext: map[string]bool{},
		recursive:  recursive,
	}, nil
}

// Close closes the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// IgnoreNext ignores the next change on a path.
func (w *Watcher) IgnoreNext(path string) {
	path = filepath.Clean(path)
	w.ignoreNext[path] = true
}

// AddPath adds a new path to watch.
func (w *Watcher) AddPath(path string) error {
	path = filepath.Clean(path)
	w.paths[path] = true

	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.Mode().IsRegular() {
		root := filepath.Dir(path)
		if w.dirs[root] {
			return nil
		}
		if err := w.watcher.Add(root); err != nil {
			return err
		}
		w.dirs[root] = true
	} else if info.Mode().IsDir() && w.recursive {
		return fs.WalkDir(os.DirFS("."), path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if w.dirs[path] {
					return fs.SkipDir
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

		// prevent reminifying the first time for files with input==output
		for path := range w.paths {
			if info, err := os.Lstat(path); err == nil && info.Mode().IsRegular() {
				changetimes[path] = time.Now()
			}
		}

		for w.watcher.Events != nil && w.watcher.Errors != nil {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					w.watcher.Events = nil
					break
				}

				filename := filepath.Clean(event.Name)
				if w.ignoreNext[filename] {
					w.ignoreNext[filename] = false
					changetimes[filename] = time.Now()
					continue
				}

				// check if changed file is being watched (as a file or indirectly in a dir)
				watched := false
				for path := range w.paths {
					if !IsDir(path) {
						if path == filename {
							watched = true
							break
						}
					} else if _, err := filepath.Rel(path, filename); err == nil {
						watched = true
						break
					}
				}
				if !watched {
					break
				}

				if info, err := os.Lstat(filename); err == nil {
					if info.Mode().IsDir() && w.recursive {
						if event.Op&fsnotify.Create == fsnotify.Create {
							if err := w.AddPath(filename); err != nil {
								Error.Println(err)
							}
						}
					} else if info.Mode().IsRegular() {
						if event.Op&fsnotify.Write == fsnotify.Write {
							if t, ok := changetimes[filename]; !ok || 100*time.Millisecond < time.Since(t) {
								time.Sleep(100 * time.Millisecond) // wait to make sure write is finished
								files <- event.Name
								changetimes[filename] = time.Now()
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
