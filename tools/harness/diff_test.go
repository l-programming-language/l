package main

import (
	"testing"
)

func TestCompareBytesPreservesByteContract(t *testing.T) {
	for _, tc := range []struct {
		name, expected, actual string
		offset                 int
	}{
		{"trailing newline", "Hello World", "Hello World\n", 11},
		{"carriage return", "one\ntwo", "one\r\ntwo", 3},
		{"UTF-8 apostrophe", "CoreLib's", "CoreLib’s", 7},
		{"missing byte", "abc", "ab", 2},
		{"first byte", "a", "b", 0},
		{"empty", "", "x", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := compareBytes([]byte(tc.expected), []byte(tc.actual))
			if r.Equal || r.FirstDifference == nil || *r.FirstDifference != tc.offset {
				t.Fatalf("wrong exact-byte comparison: %+v", r)
			}
			if r.ExpectedSHA256 == r.ActualSHA256 || r.ExpectedLength != len(tc.expected) || r.ActualLength != len(tc.actual) {
				t.Fatalf("wrong hashes or lengths: %+v", r)
			}
		})
	}
	r := compareBytes([]byte("same\x00bytes"), []byte("same\x00bytes"))
	if !r.Equal || r.FirstDifference != nil || r.ExpectedSHA256 != r.ActualSHA256 {
		t.Fatalf("equal bytes differ: %+v", r)
	}
}
