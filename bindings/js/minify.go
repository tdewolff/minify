package main

import "C"
import (
	"fmt"
	"os"
	"regexp"
	"strconv"
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
	minifyConfig(nil, nil, 0)
}

func goBytes(str *C.char, length, capacity C.longlong) []byte {
	// somehow address space on 32-bit system is smaller than 1<<30
	return (*[1 << 28]byte)(unsafe.Pointer(str))[:length:capacity]
}

func goStringArray(carr **C.char, length C.longlong) []string {
	if length == 0 {
		return []string{}
	}

	strs := make([]string, length)
	arr := unsafe.Slice(carr, length)
	for i := 0; i < int(length); i++ {
		strs[i] = C.GoString(arr[i])
	}
	return strs
}

//export minifyConfig
func minifyConfig(ckeys **C.char, cvals **C.char, length C.longlong) *C.char {
	keys := goStringArray(ckeys, length)
	vals := goStringArray(cvals, length)

	cssMinifier := &css.Minifier{}
	htmlMinifier := &html.Minifier{}
	jsMinifier := &js.Minifier{}
	jsonMinifier := &json.Minifier{}
	svgMinifier := &svg.Minifier{}
	xmlMinifier := &xml.Minifier{}

	var err error
	for i := 0; i < len(keys); i++ {
		switch keys[i] {
		case "css-precision":
			var precision int64
			precision, err = strconv.ParseInt(vals[i], 10, 64)
			cssMinifier.Precision = int(precision)
		case "html-keep-comments":
			htmlMinifier.KeepComments, err = strconv.ParseBool(vals[i])
		case "html-keep-conditional-comments":
			htmlMinifier.KeepConditionalComments, err = strconv.ParseBool(vals[i])
		case "html-keep-special-comments":
			htmlMinifier.KeepSpecialComments, err = strconv.ParseBool(vals[i])
		case "html-keep-default-attr-vals":
			htmlMinifier.KeepDefaultAttrVals, err = strconv.ParseBool(vals[i])
		case "html-keep-document-tags":
			htmlMinifier.KeepDocumentTags, err = strconv.ParseBool(vals[i])
		case "html-keep-end-tags":
			htmlMinifier.KeepEndTags, err = strconv.ParseBool(vals[i])
		case "html-keep-whitespace":
			htmlMinifier.KeepWhitespace, err = strconv.ParseBool(vals[i])
		case "html-keep-quotes":
			htmlMinifier.KeepQuotes, err = strconv.ParseBool(vals[i])
		case "js-precision":
			var precision int64
			precision, err = strconv.ParseInt(vals[i], 10, 64)
			jsMinifier.Precision = int(precision)
		case "js-keep-var-names":
			jsMinifier.KeepVarNames, err = strconv.ParseBool(vals[i])
		case "js-version":
			var version int64
			version, err = strconv.ParseInt(vals[i], 10, 64)
			jsMinifier.Version = int(version)
		case "json-precision":
			var precision int64
			precision, err = strconv.ParseInt(vals[i], 10, 64)
			jsonMinifier.Precision = int(precision)
		case "json-keep-numbers":
			jsonMinifier.KeepNumbers, err = strconv.ParseBool(vals[i])
		case "svg-keep-comments":
			svgMinifier.KeepComments, err = strconv.ParseBool(vals[i])
		case "svg-precision":
			var precision int64
			precision, err = strconv.ParseInt(vals[i], 10, 64)
			svgMinifier.Precision = int(precision)
		case "xml-keep-whitespace":
			xmlMinifier.KeepWhitespace, err = strconv.ParseBool(vals[i])
		default:
			return C.CString(fmt.Sprintf("unknown config key: %s", keys[i]))
		}
		if err != nil {
			if err.(*strconv.NumError).Func == "ParseInt" {
				err = fmt.Errorf("\"%s\" is not an integer", vals[i])
			} else if err.(*strconv.NumError).Func == "ParseBool" {
				err = fmt.Errorf("\"%s\" is not a boolean", vals[i])
			}
			return C.CString(fmt.Sprintf("bad config value for %s: %v", keys[i], err))
		}
	}

	m = minify.New()
	m.Add("text/css", cssMinifier)
	m.Add("text/html", htmlMinifier)
	m.Add("image/svg+xml", svgMinifier)
	m.AddRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma|j|live)script(1\\.[0-5])?$|^module$"), jsMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]json$"), jsonMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)
	return nil
}

//export minifyString
func minifyString(cmediatype, cinput *C.char, input_length C.longlong, coutput *C.char, output_length *C.longlong) *C.char {
	mediatype := C.GoString(cmediatype)                    // copy
	input := goBytes(cinput, input_length, input_length+1) // +1 for NULL byte used in parser
	output := goBytes(coutput, input_length, input_length)

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
