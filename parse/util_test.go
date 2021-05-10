package parse

import (
	"bytes"
	"math/rand"
	"net/url"
	"regexp"
	"testing"

	"github.com/tdewolff/test"
)

func helperRandChars(n, m int, chars string) [][]byte {
	r := make([][]byte, n)
	for i := range r {
		for j := 0; j < m; j++ {
			r[i] = append(r[i], chars[rand.Intn(len(chars))])
		}
	}
	return r
}

func helperRandStrings(n, m int, ss []string) [][]byte {
	r := make([][]byte, n)
	for i := range r {
		for j := 0; j < m; j++ {
			r[i] = append(r[i], []byte(ss[rand.Intn(len(ss))])...)
		}
	}
	return r
}

////////////////////////////////////////////////////////////////

var wsSlices [][]byte
var entitySlices [][]byte
var encodedURLSlices [][]byte
var urlSlices [][]byte

func init() {
	wsSlices = helperRandChars(10000, 50, "abcdefg \n\r\f\t")
	entitySlices = helperRandStrings(100, 5, []string{"&quot;", "&#39;", "&#x027;", "    ", " ", "test"})
	encodedURLSlices = helperRandStrings(100, 5, []string{"%20", "%3D", "test"})
	urlSlices = helperRandStrings(100, 5, []string{"~", "\"", "<", "test"})
}

func TestCopy(t *testing.T) {
	foo := []byte("abc")
	bar := Copy(foo)
	foo[0] = 'b'
	test.String(t, string(foo), "bbc")
	test.String(t, string(bar), "abc")
}

func TestToLower(t *testing.T) {
	foo := []byte("Abc")
	bar := ToLower(foo)
	bar[1] = 'B'
	test.String(t, string(foo), "aBc")
	test.String(t, string(bar), "aBc")
}

func TestEqualFold(t *testing.T) {
	test.That(t, EqualFold([]byte("Abc"), []byte("abc")))
	test.That(t, !EqualFold([]byte("Abcd"), []byte("abc")))
	test.That(t, !EqualFold([]byte("Bbc"), []byte("abc")))
	test.That(t, !EqualFold([]byte("[]"), []byte("{}"))) // same distance in ASCII as 'a' and 'A'
}

func TestWhitespace(t *testing.T) {
	test.That(t, IsAllWhitespace([]byte("\t \r\n\f")))
	test.That(t, !IsAllWhitespace([]byte("\t \r\n\fx")))
}

func TestTrim(t *testing.T) {
	test.Bytes(t, TrimWhitespace([]byte("a")), []byte("a"))
	test.Bytes(t, TrimWhitespace([]byte(" a")), []byte("a"))
	test.Bytes(t, TrimWhitespace([]byte("a ")), []byte("a"))
	test.Bytes(t, TrimWhitespace([]byte(" ")), []byte(""))
}

func TestReplaceMultipleWhitespace(t *testing.T) {
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("  a")), []byte(" a"))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("a  ")), []byte("a "))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("a  b  ")), []byte("a b "))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("  a  b  ")), []byte(" a b "))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte(" a b  ")), []byte(" a b "))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("  a b ")), []byte(" a b "))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("   a")), []byte(" a"))
	test.Bytes(t, ReplaceMultipleWhitespace([]byte("a  b")), []byte("a b"))
}

func TestReplaceMultipleWhitespaceRandom(t *testing.T) {
	wsRegexp := regexp.MustCompile("[ \t\f]+")
	wsNewlinesRegexp := regexp.MustCompile("[ ]*[\r\n][ \r\n]*")
	for _, e := range wsSlices {
		reference := wsRegexp.ReplaceAll(e, []byte(" "))
		reference = wsNewlinesRegexp.ReplaceAll(reference, []byte("\n"))
		test.Bytes(t, ReplaceMultipleWhitespace(Copy(e)), reference, "in '"+string(e)+"'")
	}
}

func TestReplaceEntities(t *testing.T) {
	entitiesMap := map[string][]byte{
		"varphi": []byte("&phiv;"),
		"varpi":  []byte("&piv;"),
		"quot":   []byte("\""),
		"apos":   []byte("'"),
		"amp":    []byte("&"),
	}
	revEntitiesMap := map[byte][]byte{
		'\'': []byte("&#39;"),
	}
	var entityTests = []struct {
		entity   string
		expected string
	}{
		{"&#34;", `"`},
		{"&#039;", `&#39;`},
		{"&#x0022;", `"`},
		{"&#x27;", `&#39;`},
		{"&#160;", `&#160;`},
		{"&quot;", `"`},
		{"&apos;", `&#39;`},
		{"&#9191;", `&#9191;`},
		{"&#x23e7;", `&#9191;`},
		{"&#x23E7;", `&#9191;`},
		{"&#x23E7;", `&#9191;`},
		{"&#x270F;", `&#9999;`},
		{"&#x2710;", `&#x2710;`},
		{"&apos;&quot;", `&#39;"`},
		{"&#34", `&#34`},
		{"&#x22", `&#x22`},
		{"&apos", `&apos`},
		{"&amp;", `&`},
		{"&#39;", `&#39;`},
		{"&amp;amp;", `&amp;amp;`},
		{"&amp;#34;", `&amp;#34;`},
		{"&amp;a mp;", `&a mp;`},
		{"&amp;DiacriticalAcute;", `&amp;DiacriticalAcute;`},
		{"&amp;CounterClockwiseContourIntegral;", `&amp;CounterClockwiseContourIntegral;`},
		{"&amp;CounterClockwiseContourIntegralL;", `&CounterClockwiseContourIntegralL;`},
		{"&varphi;", "&phiv;"},
		{"&varpi;", "&piv;"},
		{"&varnone;", "&varnone;"},
	}
	for _, tt := range entityTests {
		t.Run(tt.entity, func(t *testing.T) {
			b := ReplaceEntities([]byte(tt.entity), entitiesMap, revEntitiesMap)
			test.T(t, string(b), tt.expected, "in '"+tt.entity+"'")
		})
	}
}

func TestReplaceEntitiesRandom(t *testing.T) {
	entitiesMap := map[string][]byte{
		"quot": []byte("\""),
		"apos": []byte("'"),
	}
	revEntitiesMap := map[byte][]byte{
		'\'': []byte("&#39;"),
	}

	quotRegexp := regexp.MustCompile("&quot;")
	aposRegexp := regexp.MustCompile("(&#39;|&#x027;)")
	for _, e := range entitySlices {
		reference := quotRegexp.ReplaceAll(e, []byte("\""))
		reference = aposRegexp.ReplaceAll(reference, []byte("&#39;"))
		test.Bytes(t, ReplaceEntities(Copy(e), entitiesMap, revEntitiesMap), reference, "in '"+string(e)+"'")
	}
}

func TestReplaceMultipleWhitespaceAndEntities(t *testing.T) {
	entitiesMap := map[string][]byte{
		"varphi": []byte("&phiv;"),
	}
	var entityTests = []struct {
		entity   string
		expected string
	}{
		{"  &varphi;  &#34; \n ", " &phiv; \"\n"},
	}
	for _, tt := range entityTests {
		t.Run(tt.entity, func(t *testing.T) {
			b := ReplaceMultipleWhitespaceAndEntities([]byte(tt.entity), entitiesMap, nil)
			test.T(t, string(b), tt.expected, "in '"+tt.entity+"'")
		})
	}
}

func TestReplaceMultipleWhitespaceAndEntitiesRandom(t *testing.T) {
	entitiesMap := map[string][]byte{
		"quot": []byte("\""),
		"apos": []byte("'"),
	}
	revEntitiesMap := map[byte][]byte{
		'\'': []byte("&#39;"),
	}

	wsRegexp := regexp.MustCompile("[ ]+")
	quotRegexp := regexp.MustCompile("&quot;")
	aposRegexp := regexp.MustCompile("(&#39;|&#x027;)")
	for _, e := range entitySlices {
		reference := wsRegexp.ReplaceAll(e, []byte(" "))
		reference = quotRegexp.ReplaceAll(reference, []byte("\""))
		reference = aposRegexp.ReplaceAll(reference, []byte("&#39;"))
		test.Bytes(t, ReplaceMultipleWhitespaceAndEntities(Copy(e), entitiesMap, revEntitiesMap), reference, "in '"+string(e)+"'")
	}
}

func TestPrintable(t *testing.T) {
	var tests = []struct {
		s         string
		printable string
	}{
		{"a", "a"},
		{"\x00", "0x00"},
		{"\x7F", "0x7F"},
		{"\u0800", "ࠀ"},
		{"\u200F", "U+200F"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			printable := ""
			for _, r := range tt.s {
				printable += Printable(r)
			}
			test.T(t, printable, tt.printable)
		})
	}
}

func TestDecodeURL(t *testing.T) {
	var urlTests = []struct {
		url      string
		expected string
	}{
		{"%20%3F%7E", " ?~"},
		{"%80", "%80"},
		{"%2B%2b", "++"},
		{"%' ", "%' "},
		{"a+b", "a b"},
	}
	for _, tt := range urlTests {
		t.Run(tt.url, func(t *testing.T) {
			b := DecodeURL([]byte(tt.url))
			test.T(t, string(b), tt.expected, "in '"+tt.url+"'")
		})
	}
}

func TestDecodeURLRandom(t *testing.T) {
	for _, e := range encodedURLSlices {
		reference, _ := url.QueryUnescape(string(e))
		test.Bytes(t, DecodeURL(Copy(e)), []byte(reference), "in '"+string(e)+"'")
	}
}

func TestEncodeURL(t *testing.T) {
	var urlTests = []struct {
		url      string
		expected string
	}{
		{"AZaz09-_.!~*'()", "AZaz09-_.!~*'()"},
		{"<>", "%3C%3E"},
		{"\u2318", "%E2%8C%98"},
		{"a b", "a+b"},
	}
	for _, tt := range urlTests {
		t.Run(tt.url, func(t *testing.T) {
			b := EncodeURL([]byte(tt.url), URLEncodingTable)
			test.T(t, string(b), tt.expected, "in '"+tt.url+"'")
		})
	}
}

func TestEncodeURLRandom(t *testing.T) {
	for _, e := range urlSlices {
		reference := url.QueryEscape(string(e))
		test.Bytes(t, EncodeURL(Copy(e), URLEncodingTable), []byte(reference), "in '"+string(e)+"'")
	}
}

////////////////////////////////////////////////////////////////

func BenchmarkBytesTrim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			bytes.TrimSpace(e)
		}
	}
}

func BenchmarkTrim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			TrimWhitespace(e)
		}
	}
}

func BenchmarkReplaceMultipleWhitespace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			ReplaceMultipleWhitespace(e)
		}
	}
}

func BenchmarkWhitespaceTable(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if IsWhitespace(c) {
					n++
				}
			}
		}
	}
}

func BenchmarkWhitespaceIf1(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if c == ' ' {
					n++
				}
			}
		}
	}
}

func BenchmarkWhitespaceIf2(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if c == ' ' || c == '\n' {
					n++
				}
			}
		}
	}
}

func BenchmarkWhitespaceIf3(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if c == ' ' || c == '\n' || c == '\r' {
					n++
				}
			}
		}
	}
}

func BenchmarkWhitespaceIf4(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
					n++
				}
			}
		}
	}
}

func BenchmarkWhitespaceIf5(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			for _, c := range e {
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
					n++
				}
			}
		}
	}
}
