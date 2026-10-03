//go:build gofuzz
// +build gofuzz

package fuzz

import (
	"bytes"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	hasScript := bytes.Contains(bytes.ToLower(data), []byte("</script>"))

	w := &bytes.Buffer{}
	r := bytes.NewBuffer(data)
	var params map[string]string
	if !hasScript {
		// is </script> is present, it was not originally embedded in HTML
		params = map[string]string{"escape-html": "1"}
	}
	_ = js.Minify(minify.New(), w, r, params)
	if hasScript != bytes.Contains(bytes.ToLower(w.Bytes()), []byte("</script>")) {
		panic("</script> added or removed")
	}
	return 1
}
