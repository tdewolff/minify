package main

/*
#include <string.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	jsonmin "github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"github.com/tdewolff/parse/v2/buffer"
)

type minifyOptions struct {
	Type                        string `json:"type"`
	CSSPrecision                int    `json:"cssPrecision"`
	CSSVersion                  int    `json:"cssVersion"`
	HTMLKeepComments            bool   `json:"htmlKeepComments"`
	HTMLKeepConditionalComments bool   `json:"htmlKeepConditionalComments"`
	HTMLKeepDefaultAttrvals     bool   `json:"htmlKeepDefaultAttrvals"`
	HTMLKeepDocumentTags        bool   `json:"htmlKeepDocumentTags"`
	HTMLKeepEndTags             bool   `json:"htmlKeepEndTags"`
	HTMLKeepQuotes              bool   `json:"htmlKeepQuotes"`
	HTMLKeepSpecialComments     bool   `json:"htmlKeepSpecialComments"`
	HTMLKeepWhitespace          bool   `json:"htmlKeepWhitespace"`
	JSKeepVarNames              bool   `json:"jsKeepVarNames"`
	JSPrecision                 int    `json:"jsPrecision"`
	JSVersion                   int    `json:"jsVersion"`
	JSONKeepNumbers             bool   `json:"jsonKeepNumbers"`
	JSONPrecision               int    `json:"jsonPrecision"`
	SVGKeepComments             bool   `json:"svgKeepComments"`
	SVGPrecision                int    `json:"svgPrecision"`
	XMLKeepWhitespace           bool   `json:"xmlKeepWhitespace"`
}

var (
	jsMediatypePattern   = regexp.MustCompile("^(application|text)/(x-)?(java|ecma|j|live)script(1\\.[0-5])?$|^module$")
	jsonMediatypePattern = regexp.MustCompile("[/+]json$")
	xmlMediatypePattern  = regexp.MustCompile("[/+]xml$")
)

type minifyResult struct {
	Error string `json:"error"`
	Data  string `json:"data"`
}

func goBytes(str *C.char, length, capacity C.longlong) []byte {
	// address space on 32-bit system is smaller than 1<<30
	return (*[1 << 28]byte)(unsafe.Pointer(str))[:length:capacity]
}

func parseOptions(cOptionsJson *C.char) (minifyOptions, error) {
	opts := minifyOptions{}
	if cOptionsJson == nil {
		return opts, nil
	}

	raw := strings.TrimSpace(C.GoString(cOptionsJson))
	if raw == "" {
		return opts, nil
	}

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()

	rawOpts := map[string]interface{}{}
	if err := decoder.Decode(&rawOpts); err != nil {
		return opts, err
	}

	opts.Type = parseString(rawOpts["type"])
	opts.CSSPrecision = parseInt(rawOpts["cssPrecision"])
	opts.CSSVersion = parseInt(rawOpts["cssVersion"])
	opts.HTMLKeepComments = parseBool(rawOpts["htmlKeepComments"])
	opts.HTMLKeepConditionalComments = parseBool(rawOpts["htmlKeepConditionalComments"])
	opts.HTMLKeepDefaultAttrvals = parseBool(rawOpts["htmlKeepDefaultAttrvals"])
	opts.HTMLKeepDocumentTags = parseBool(rawOpts["htmlKeepDocumentTags"])
	opts.HTMLKeepEndTags = parseBool(rawOpts["htmlKeepEndTags"])
	opts.HTMLKeepQuotes = parseBool(rawOpts["htmlKeepQuotes"])
	opts.HTMLKeepSpecialComments = parseBool(rawOpts["htmlKeepSpecialComments"])
	opts.HTMLKeepWhitespace = parseBool(rawOpts["htmlKeepWhitespace"])
	opts.JSKeepVarNames = parseBool(rawOpts["jsKeepVarNames"])
	opts.JSPrecision = parseInt(rawOpts["jsPrecision"])
	opts.JSVersion = parseInt(rawOpts["jsVersion"])
	opts.JSONKeepNumbers = parseBool(rawOpts["jsonKeepNumbers"])
	opts.JSONPrecision = parseInt(rawOpts["jsonPrecision"])
	opts.SVGKeepComments = parseBool(rawOpts["svgKeepComments"])
	opts.SVGPrecision = parseInt(rawOpts["svgPrecision"])
	opts.XMLKeepWhitespace = parseBool(rawOpts["xmlKeepWhitespace"])

	return opts, nil
}

func parseString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func parseInt(v interface{}) int {
	switch n := v.(type) {
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return i
		}
	}
	return 0
}

func parseBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		val := strings.TrimSpace(strings.ToLower(b))
		return val == "true" || val == "1" || val == "yes"
	}
	return false
}

func buildResult(err error, data string) *C.char {
	res := minifyResult{}
	if err != nil {
		res.Error = err.Error()
	} else {
		res.Data = data
	}
	b, _ := json.Marshal(res)
	return C.CString(string(b))
}

func newMinifier(opts minifyOptions) (*minify.M, error) {
	m := minify.New()

	cssMinifier := &css.Minifier{
		Precision: opts.CSSPrecision,
		Version:   opts.CSSVersion,
	}
	htmlMinifier := &html.Minifier{
		KeepComments:            opts.HTMLKeepComments,
		KeepConditionalComments: opts.HTMLKeepConditionalComments,
		KeepDefaultAttrVals:     opts.HTMLKeepDefaultAttrvals,
		KeepDocumentTags:        opts.HTMLKeepDocumentTags,
		KeepEndTags:             opts.HTMLKeepEndTags,
		KeepQuotes:              opts.HTMLKeepQuotes,
		KeepSpecialComments:     opts.HTMLKeepSpecialComments,
		KeepWhitespace:          opts.HTMLKeepWhitespace,
	}
	jsMinifier := &js.Minifier{
		Precision:    opts.JSPrecision,
		KeepVarNames: opts.JSKeepVarNames,
		Version:      opts.JSVersion,
	}
	jsonMinifier := &jsonmin.Minifier{
		Precision:   opts.JSONPrecision,
		KeepNumbers: opts.JSONKeepNumbers,
	}
	svgMinifier := &svg.Minifier{
		KeepComments: opts.SVGKeepComments,
		Precision:    opts.SVGPrecision,
	}
	xmlMinifier := &xml.Minifier{
		KeepWhitespace: opts.XMLKeepWhitespace,
	}

	m.Add("text/css", cssMinifier)
	m.Add("text/html", htmlMinifier)
	m.Add("image/svg+xml", svgMinifier)
	m.AddRegexp(jsMediatypePattern, jsMinifier)
	m.AddRegexp(jsonMediatypePattern, jsonMinifier)
	m.AddRegexp(xmlMediatypePattern, xmlMinifier)
	m.Add("importmap", jsonMinifier)
	m.Add("speculationrules", jsonMinifier)

	aspMinifier := *htmlMinifier
	aspMinifier.TemplateDelims = html.ASPTemplateDelims
	m.Add("text/asp", &aspMinifier)
	m.Add("text/x-ejs-template", &aspMinifier)

	phpMinifier := *htmlMinifier
	phpMinifier.TemplateDelims = html.PHPTemplateDelims
	m.Add("application/x-httpd-php", &phpMinifier)

	tmplMinifier := *htmlMinifier
	tmplMinifier.TemplateDelims = html.GoTemplateDelims
	m.Add("text/x-template", &tmplMinifier)
	m.Add("text/x-go-template", &tmplMinifier)
	m.Add("text/x-mustache-template", &tmplMinifier)
	m.Add("text/x-handlebars-template", &tmplMinifier)

	return m, nil
}

func resolveType(t string) (string, error) {
	trimmed := strings.TrimSpace(t)
	if trimmed == "" {
		return "", nil
	}

	switch trimmed {
	case "text/css",
		"text/html",
		"image/svg+xml",
		"importmap",
		"speculationrules",
		"text/asp",
		"text/x-ejs-template",
		"application/x-httpd-php",
		"text/x-template",
		"text/x-go-template",
		"text/x-mustache-template",
		"text/x-handlebars-template":
		return trimmed, nil
	}

	if jsMediatypePattern.MatchString(trimmed) || jsonMediatypePattern.MatchString(trimmed) || xmlMediatypePattern.MatchString(trimmed) {
		return trimmed, nil
	}
	return "", fmt.Errorf("invalid type %q", trimmed)
}

//export MinifyString
func MinifyString(cData *C.char, cOptionsJson *C.char) *C.char {
	opts, err := parseOptions(cOptionsJson)
	if err != nil {
		return buildResult(err, "")
	}

	var dataBytes []byte
	if cData != nil {
		dataLen := C.longlong(C.strlen(cData))
		dataBytes = goBytes(cData, dataLen, dataLen+1) // +1 for NULL byte
	}
	if len(dataBytes) == 0 {
		return buildResult(errors.New("data is required"), "")
	}
	mediatype, err := resolveType(opts.Type)
	if err != nil {
		return buildResult(err, "")
	}
	if mediatype == "" {
		return buildResult(errors.New("type is required"), "")
	}

	m, err := newMinifier(opts)
	if err != nil {
		return buildResult(err, "")
	}

	outBuf := buffer.NewWriter(make([]byte, 0, len(dataBytes)))
	if err := m.Minify(mediatype, outBuf, buffer.NewReader(dataBytes)); err != nil {
		return buildResult(err, "")
	}

	return buildResult(nil, string(outBuf.Bytes()))
}

//export FreeCString
func FreeCString(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

func main() {}
