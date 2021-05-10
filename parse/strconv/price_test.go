package strconv

import (
	"fmt"
	"testing"

	"github.com/tdewolff/test"
)

func TestAppendPrice(t *testing.T) {
	priceTests := []struct {
		price    int64
		dec      bool
		expected string
	}{
		{0, false, "0"},
		{0, true, "0.00"},
		{100, true, "1.00"},
		{-100, true, "1.00"},
		{100000, false, "1,000"},
		{100000, true, "1,000.00"},
		{123456789012, true, "1,234,567,890.12"},
		{9223372036854775807, true, "92,233,720,368,547,758.07"},
		{-9223372036854775808, true, "92,233,720,368,547,758.08"},
		{149, false, "1"},
		{150, false, "2"},
	}

	for _, tt := range priceTests {
		t.Run(fmt.Sprint(tt.price), func(t *testing.T) {
			price := AppendPrice(make([]byte, 0, 4), tt.price, tt.dec, ',', '.')
			test.String(t, string(price), tt.expected, "for", tt.price)
		})
	}

	// coverage
}
