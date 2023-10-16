package main

import (
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/tdewolff/test"
)

func TestCreateTasks(t *testing.T) {
	fsys := fstest.MapFS{
		"a.js":     {},
		"dir/b.js": {},
	}

	tests := []struct {
		input, output string
		tasks         map[string]string
	}{
		// root file
		{"a.js", "", map[string]string{"a.js": ""}},
		{"a.js", ".", map[string]string{"a.js": "a.js"}},
		{"a.js", "./", map[string]string{"a.js": "a.js"}},
		{"a.js", "out", map[string]string{"a.js": "out"}},
		{"a.js", "out/", map[string]string{"a.js": "out/a.js"}},

		// nested file
		{"dir/b.js", "", map[string]string{"dir/b.js": ""}},
		{"dir/b.js", ".", map[string]string{"dir/b.js": "b.js"}},
		{"dir/b.js", "./", map[string]string{"dir/b.js": "b.js"}},
		{"dir/b.js", "out", map[string]string{"dir/b.js": "out"}},
		{"dir/b.js", "out/", map[string]string{"dir/b.js": "out/b.js"}},

		// directory
		{"dir", "", map[string]string{"dir/b.js": ""}},
		{"dir", ".", map[string]string{"dir/b.js": "dir/b.js"}},
		{"dir", "./", map[string]string{"dir/b.js": "dir/b.js"}},
		{"dir", "out/", map[string]string{"dir/b.js": "out/dir/b.js"}},
		{"dir/", "out/", map[string]string{"dir/b.js": "out/b.js"}},
	}

	recursive = true
	for _, tt := range tests {
		t.Run(tt.input+" => "+tt.output, func(t *testing.T) {
			tasks, _, err := createTasks(fsys, []string{tt.input}, tt.output)
			test.Error(t, err)
			if len(tasks) != len(tt.tasks) {
				test.Fail(t, fmt.Sprintf("missing %v", tt.tasks))
			}
			for _, task := range tasks {
				if dst, ok := tt.tasks[task.srcs[0]]; !ok || dst != task.dst {
					test.Fail(t, fmt.Sprintf("unexpected %s => %s", task.srcs[0], task.dst))
				}
			}
		})
	}
}
