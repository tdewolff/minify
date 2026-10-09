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
	hasComment := bytes.Contains(data, []byte("<!--"))

	w := &bytes.Buffer{}
	r := bytes.NewBuffer(data)
	var params map[string]string
	if !hasScript && !hasComment {
		// is </script> is present, it was not originally embedded in HTML
		params = map[string]string{"escape-html": "1"}
	}
	err := js.Minify(minify.New(), w, r, params)
	if err == nil && (!hasScript && bytes.Contains(bytes.ToLower(w.Bytes()), []byte("</script>")) || !hasComment && bytes.Contains(data, []byte("<!--"))) {
		panic("</script> or <!-- added")
	}
	return 1
}
