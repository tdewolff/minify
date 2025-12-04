package main

/*
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct {
	const char *mediatype;
	const char *data;
	int32_t cssPrecision;
	int32_t cssVersion;
	bool htmlKeepComments;
	bool htmlKeepConditionalComments;
	bool htmlKeepDefaultAttrvals;
	bool htmlKeepDocumentTags;
	bool htmlKeepEndTags;
	bool htmlKeepQuotes;
	bool htmlKeepSpecialComments;
	bool htmlKeepWhitespace;
	bool jsKeepVarNames;
	int32_t jsPrecision;
	int32_t jsVersion;
	bool jsonKeepNumbers;
	int32_t jsonPrecision;
	bool svgKeepComments;
	int32_t svgPrecision;
	bool xmlKeepWhitespace;
} MinifyOptions;

typedef struct {
	char *error;
	char *data;
} MinifyResult;
*/
import "C"
import (
	"errors"
	"fmt"
	"regexp"
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
	Type                        string
	Data                        string
	CSSPrecision                int
	CSSVersion                  int
	HTMLKeepComments            bool
	HTMLKeepConditionalComments bool
	HTMLKeepDefaultAttrvals     bool
	HTMLKeepDocumentTags        bool
	HTMLKeepEndTags             bool
	HTMLKeepQuotes              bool
	HTMLKeepSpecialComments     bool
	HTMLKeepWhitespace          bool
	JSKeepVarNames              bool
	JSPrecision                 int
	JSVersion                   int
	JSONKeepNumbers             bool
	JSONPrecision               int
	SVGKeepComments             bool
	SVGPrecision                int
	XMLKeepWhitespace           bool
}

var (
	jsMediatypePattern   = regexp.MustCompile("^(application|text)/(x-)?(java|ecma|j|live)script(1\\.[0-5])?$|^module$")
	jsonMediatypePattern = regexp.MustCompile("[/+]json$")
	xmlMediatypePattern  = regexp.MustCompile("[/+]xml$")
)

func parseOptions(opts *C.MinifyOptions) (minifyOptions, error) {
	if opts == nil {
		return minifyOptions{}, errors.New("options are required")
	}

	return minifyOptions{
		Type:                        strings.TrimSpace(C.GoString(opts.mediatype)),
		Data:                        C.GoString(opts.data),
		CSSPrecision:                int(opts.cssPrecision),
		CSSVersion:                  int(opts.cssVersion),
		HTMLKeepComments:            bool(opts.htmlKeepComments),
		HTMLKeepConditionalComments: bool(opts.htmlKeepConditionalComments),
		HTMLKeepDefaultAttrvals:     bool(opts.htmlKeepDefaultAttrvals),
		HTMLKeepDocumentTags:        bool(opts.htmlKeepDocumentTags),
		HTMLKeepEndTags:             bool(opts.htmlKeepEndTags),
		HTMLKeepQuotes:              bool(opts.htmlKeepQuotes),
		HTMLKeepSpecialComments:     bool(opts.htmlKeepSpecialComments),
		HTMLKeepWhitespace:          bool(opts.htmlKeepWhitespace),
		JSKeepVarNames:              bool(opts.jsKeepVarNames),
		JSPrecision:                 int(opts.jsPrecision),
		JSVersion:                   int(opts.jsVersion),
		JSONKeepNumbers:             bool(opts.jsonKeepNumbers),
		JSONPrecision:               int(opts.jsonPrecision),
		SVGKeepComments:             bool(opts.svgKeepComments),
		SVGPrecision:                int(opts.svgPrecision),
		XMLKeepWhitespace:           bool(opts.xmlKeepWhitespace),
	}, nil
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

func setResult(out *C.MinifyResult, err error, data string) {
	if out == nil {
		return
	}

	if out.error != nil {
		C.free(unsafe.Pointer(out.error))
		out.error = nil
	}
	if out.data != nil {
		C.free(unsafe.Pointer(out.data))
		out.data = nil
	}

	if err != nil {
		out.error = C.CString(err.Error())
		return
	}

	out.data = C.CString(data)
}

//export Minify
func Minify(cOptions *C.MinifyOptions, cResult *C.MinifyResult) {
	opts, err := parseOptions(cOptions)
	if err != nil {
		setResult(cResult, err, "")
		return
	}

	dataBytes := []byte(opts.Data)
	if len(dataBytes) == 0 {
		setResult(cResult, errors.New("data is required"), "")
		return
	}
	mediatype, err := resolveType(opts.Type)
	if err != nil {
		setResult(cResult, err, "")
		return
	}
	if mediatype == "" {
		setResult(cResult, errors.New("type is required"), "")
		return
	}

	m, err := newMinifier(opts)
	if err != nil {
		setResult(cResult, err, "")
		return
	}

	outBuf := buffer.NewWriter(make([]byte, 0, len(dataBytes)))
	if err := m.Minify(mediatype, outBuf, buffer.NewReader(dataBytes)); err != nil {
		setResult(cResult, err, "")
		return
	}

	setResult(cResult, nil, string(outBuf.Bytes()))
}

//export FreeCString
func FreeCString(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

func main() {}
