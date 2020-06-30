package js

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestBinaryNumber(t *testing.T) {
	test.Bytes(t, binaryNumber([]byte("0b0")), []byte("0"))
	test.Bytes(t, binaryNumber([]byte("0b1")), []byte("1"))
	test.Bytes(t, binaryNumber([]byte("0b1001")), []byte("9"))
	test.Bytes(t, binaryNumber([]byte("0b100000000000000000000000000000000000000000000000000000000000000")), []byte("4611686018427387904"))
	test.Bytes(t, binaryNumber([]byte("0b1000000000000000000000000000000000000000000000000000000000000000")), []byte("0b1000000000000000000000000000000000000000000000000000000000000000"))
}

func TestOctalNumber(t *testing.T) {
	test.Bytes(t, octalNumber([]byte("0o0")), []byte("0"))
	test.Bytes(t, octalNumber([]byte("0o1")), []byte("1"))
	test.Bytes(t, octalNumber([]byte("0o775")), []byte("509"))
	test.Bytes(t, octalNumber([]byte("0o100000000000000000000")), []byte("1152921504606846976"))
	test.Bytes(t, octalNumber([]byte("0o1000000000000000000000")), []byte("0o1000000000000000000000"))
}
