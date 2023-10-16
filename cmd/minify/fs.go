package main

import (
	"io/fs"
	"os"
	"path/filepath"
)

func NewFS() fs.FS {
	return dirFS("")
}

type dirFS string

func (dir dirFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(string(dir), name))
}

func (dir dirFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(string(dir), name))
}
