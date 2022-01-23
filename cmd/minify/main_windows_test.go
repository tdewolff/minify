//go:build windows
// +build windows

package main

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestIsDir_Windows(t *testing.T) {
	cases := []struct {
		name     string
		dir      string
		expected bool
	}{
		{"SimpleFile", "file", false},
		{"FileInCurrentDirectory", ".\\file", false},
		{"FileInParentDirectory", "..\\file", false},
		{"FileInFullyQualifiedDirectory", "c:\\path\to\\file", false},
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
