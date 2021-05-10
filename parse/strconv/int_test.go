package strconv

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/tdewolff/test"
)

func TestParseInt(t *testing.T) {
	intTests := []struct {
		i        string
		expected int64
	}{
		{"5", 5},
		{"99", 99},
		{"999", 999},
		{"-5", -5},
		{"+5", 5},
		{"9223372036854775807", 9223372036854775807},
		{"-9223372036854775807", -9223372036854775807},
		{"-9223372036854775808", -9223372036854775808},
	}
	for _, tt := range intTests {
		t.Run(fmt.Sprint(tt.i), func(t *testing.T) {
			i, n := ParseInt([]byte(tt.i))
			test.T(t, n, len(tt.i))
			test.T(t, i, tt.expected)
		})
	}
}

func TestParseIntError(t *testing.T) {
	intTests := []struct {
		i        string
		n        int
		expected int64
	}{
		{"a", 0, 0},
		{"+", 0, 0},
		{"9223372036854775808", 0, 0},
		{"-9223372036854775809", 0, 0},
		{"18446744073709551620", 0, 0},
	}
	for _, tt := range intTests {
		t.Run(fmt.Sprint(tt.i), func(t *testing.T) {
			i, n := ParseInt([]byte(tt.i))
			test.T(t, n, tt.n)
			test.T(t, i, tt.expected)
		})
	}
}

func TestLenInt(t *testing.T) {
	lenIntTests := []struct {
		i        int64
		expected int
	}{
		{0, 1},
		{1, 1},
		{10, 2},
		{99, 2},
		{9223372036854775807, 19},
		{-9223372036854775808, 19},

		// coverage
		{100, 3},
		{1000, 4},
		{10000, 5},
		{100000, 6},
		{1000000, 7},
		{10000000, 8},
		{100000000, 9},
		{1000000000, 10},
		{10000000000, 11},
		{100000000000, 12},
		{1000000000000, 13},
		{10000000000000, 14},
		{100000000000000, 15},
		{1000000000000000, 16},
		{10000000000000000, 17},
		{100000000000000000, 18},
		{1000000000000000000, 19},
	}
	for _, tt := range lenIntTests {
		t.Run(fmt.Sprint(tt.i), func(t *testing.T) {
			test.T(t, LenInt(tt.i), tt.expected)
		})
	}
}

////////////////////////////////////////////////////////////////

var num []int64

func TestMain(t *testing.T) {
	for j := 0; j < 1000; j++ {
		num = append(num, rand.Int63n(1000))
	}
}

func BenchmarkLenIntLog(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			n += int(math.Log10(math.Abs(float64(num[j])))) + 1
		}
	}
}

func BenchmarkLenIntSwitch(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			n += LenInt(num[j])
		}
	}
}
