package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/matryer/try"
)

// IsDir returns true if the passed string looks like it specifies a directory, false otherwise.
func IsDir(dir string) bool {
	if 0 < len(dir) && dir[len(dir)-1] == os.PathSeparator {
		return true
	}
	info, err := os.Lstat(dir)
	return err == nil && info.Mode().IsDir() && info.Mode()&os.ModeSymlink == 0
}

// SameFile returns true if the two file paths specify the same path.
// While Linux is case-preserving case-sensitive (and therefore a string comparison will work),
// Windows is case-preserving case-insensitive; we use os.SameFile() to work cross-platform.
func SameFile(filename1 string, filename2 string) (bool, error) {
	fi1, err := os.Stat(filename1)
	if err != nil {
		return false, err
	}

	fi2, err := os.Stat(filename2)
	if err != nil {
		return false, err
	}
	return os.SameFile(fi1, fi2), nil
}

func openInputFile(input string) (io.ReadCloser, error) {
	var r *os.File
	if input == "" {
		r = os.Stdin
	} else {
		err := try.Do(func(attempt int) (bool, error) {
			var ferr error
			r, ferr = os.Open(input)
			return attempt < 5, ferr
		})

		if err != nil {
			return nil, fmt.Errorf("open input file %q: %w", input, err)
		}
	}
	return r, nil
}

func openInputFiles(filenames []string, sep []byte) (*concatFileReader, error) {
	return newConcatFileReader(filenames, openInputFile, sep)
}

func openOutputFile(output string) (*os.File, error) {
	var w *os.File
	if output == "" {
		w = os.Stdout
	} else {
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0777); err != nil {
			return nil, fmt.Errorf("creating directory %q: %w", dir, err)
		}

		err := try.Do(func(attempt int) (bool, error) {
			var ferr error
			w, ferr = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			return attempt < 5, ferr
		})

		if err != nil {
			return nil, fmt.Errorf("open output file %q: %w", output, err)
		}
	}
	return w, nil
}

func createSymlink(input, output string) error {
	if _, err := os.Stat(output); err == nil {
		if err = os.Remove(output); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0777); err != nil {
		return err
	}
	if err := os.Symlink(input, output); err != nil {
		return err
	}
	return nil
}

type concatFileReader struct {
	filenames []string
	sep       []byte
	opener    func(string) (io.ReadCloser, error)

	cur     io.ReadCloser
	sepLeft int
}

func newConcatFileReader(filenames []string, opener func(string) (io.ReadCloser, error), sep []byte) (*concatFileReader, error) {
	var cur io.ReadCloser
	if 0 < len(filenames) {
		var filename string
		filename, filenames = filenames[0], filenames[1:]

		var err error
		if cur, err = opener(filename); err != nil {
			return nil, err
		}
	}
	return &concatFileReader{filenames, sep, opener, cur, 0}, nil
}

func (r *concatFileReader) Read(p []byte) (int, error) {
	m := r.writeSep(p) // write remaining separator
	if r.cur == nil {
		return m, io.EOF
	}
	n, err := r.cur.Read(p[m:])
	n += m

	// current reader is finished, load in the new reader
	if err == io.EOF {
		if err := r.cur.Close(); err != nil {
			return n, err
		}
		r.cur = nil

		if 0 < len(r.filenames) {
			var filename string
			filename, r.filenames = r.filenames[0], r.filenames[1:]
			if r.cur, err = r.opener(filename); err != nil {
				return n, err
			}
			r.sepLeft = len(r.sep)

			// if previous read returned (0, io.EOF), read from the new reader
			if n == 0 {
				return r.Read(p)
			}
			n += r.writeSep(p[n:])
		}
	}
	return n, err
}

func (r *concatFileReader) writeSep(p []byte) int {
	if 0 < r.sepLeft {
		m := copy(p, r.sep[len(r.sep)-r.sepLeft:])
		r.sepLeft -= m
		return m
	}
	return 0
}

func (r *concatFileReader) Close() error {
	if r.cur != nil {
		return r.cur.Close()
	}
	return nil
}
