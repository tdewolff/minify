// +build linux darwin netbsd solaris openbsd js wasm

package main

import (
	"os"
	"syscall"
)

func getOwnership(info os.FileInfo) (int, int, bool) {
	if stat_t, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat_t.Uid), int(stat_t.Gid), true
	}
	return 0, 0, false
}
