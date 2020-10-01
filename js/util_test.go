package js

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestBinaryNumber(t *testing.T) {
	test.Bytes(t, binaryNumber([]byte("0b0"), 0), []byte("0"))
	test.Bytes(t, binaryNumber([]byte("0b1"), 0), []byte("1"))
	test.Bytes(t, binaryNumber([]byte("0b1001"), 0), []byte("9"))
	test.Bytes(t, binaryNumber([]byte("0b100000000000000000000000000000000000000000000000000000000000000"), 0), []byte("4611686018427387904"))
	test.Bytes(t, binaryNumber([]byte("0b1000000000000000000000000000000000000000000000000000000000000000"), 0), []byte("0b1000000000000000000000000000000000000000000000000000000000000000"))
}

func TestOctalNumber(t *testing.T) {
	test.Bytes(t, octalNumber([]byte("0o0"), 0), []byte("0"))
	test.Bytes(t, octalNumber([]byte("0o1"), 0), []byte("1"))
	test.Bytes(t, octalNumber([]byte("0o775"), 0), []byte("509"))
	test.Bytes(t, octalNumber([]byte("0o100000000000000000000"), 0), []byte("1152921504606846976"))
	test.Bytes(t, octalNumber([]byte("0o1000000000000000000000"), 0), []byte("0o1000000000000000000000"))
}

func TestHexadecimalNumber(t *testing.T) {
	test.Bytes(t, hexadecimalNumber([]byte("0x0"), 0), []byte("0"))
	test.Bytes(t, hexadecimalNumber([]byte("0x1"), 0), []byte("1"))
	test.Bytes(t, hexadecimalNumber([]byte("0xFE"), 0), []byte("254"))
	test.Bytes(t, hexadecimalNumber([]byte("0x1000000000"), 0), []byte("68719476736"))
	test.Bytes(t, hexadecimalNumber([]byte("0xd000000000"), 0), []byte("893353197568"))
	test.Bytes(t, hexadecimalNumber([]byte("0xe000000000"), 0), []byte("0xe000000000"))
	test.Bytes(t, hexadecimalNumber([]byte("0xE000000000"), 0), []byte("0xE000000000"))
	test.Bytes(t, hexadecimalNumber([]byte("0x10000000000"), 0), []byte("0x10000000000"))
}
