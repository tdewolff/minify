// +build gofuzz

package fuzz

import "github.com/tdewolff/minify/v2/parse"

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	newData := parse.ReplaceEntities(data, map[string][]byte{
		"test":  []byte("&t;"),
		"test3": []byte("&test;"),
		"test5": []byte("&#5;"),
		"quot":  []byte("\""),
		"apos":  []byte("'"),
	}, map[byte][]byte{
		'\'': []byte("&#34;"),
		'"':  []byte("&#39;"),
	})
	if len(newData) > len(data) {
		panic("output longer than input")
	}
	return 1
}
