package main

import "C"
import (
	"os"
	"regexp"
	"unsafe"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"github.com/tdewolff/parse/v2/buffer"
)

var m *minify.M

func init() {
	m = minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
}

func goBytes(str *C.char, length C.longlong) []byte {
	return (*[1 << 32]byte)(unsafe.Pointer(str))[:length:length]
}

//export minifyString
func minifyString(cmediatype, cinput *C.char, input_length C.longlong, coutput *C.char, output_length *C.longlong) *C.char {
	mediatype := C.GoString(cmediatype) // copy
	input := goBytes(cinput, input_length)
	output := goBytes(coutput, input_length)

	out := buffer.NewStaticWriter(output[:0])
	if err := m.Minify(mediatype, out, buffer.NewReader(input)); err != nil {
		return C.CString(err.Error())
	} else if err := out.Close(); err != nil {
		return C.CString(err.Error())
	}
	*output_length = C.longlong(out.Len())
	return nil
}

//export minifyFile
func minifyFile(cmediatype, cinput, coutput *C.char) *C.char {
	mediatype := C.GoString(cmediatype) // copy
	input := C.GoString(cinput)
	output := C.GoString(coutput)

	fi, err := os.Open(input)
	if err != nil {
		return C.CString(err.Error())
	}

	fo, err := os.Create(output)
	if err != nil {
		return C.CString(err.Error())
	}

	if err := m.Minify(mediatype, fo, fi); err != nil {
		fi.Close()
		fo.Close()
		return C.CString(err.Error())
	} else if err := fi.Close(); err != nil {
		fo.Close()
		return C.CString(err.Error())
	} else if err := fo.Close(); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

func main() {}
