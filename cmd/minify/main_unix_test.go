//go:build linux || darwin || netbsd || solaris || openbsd || js || wasm
// +build linux darwin netbsd solaris openbsd js wasm

package main

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestIsDirUnix(t *testing.T) {
	cases := []struct {
		name     string
		dir      string
		expected bool
	}{
		{"SimpleFile", "file", false},
		{"FileInCurrentDirectory", "./file", false},
		{"FileInParentDirectory", "../file", false},
		{"FileInFullyQualifiedDirectory", "/path/to/file", false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := IsDir(c.dir)
			test.T(t, actual, c.expected)
		})
	}
}
