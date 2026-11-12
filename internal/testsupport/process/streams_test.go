package process

import (
	"bytes"
	"testing"
)

func TestStreams(t *testing.T) {
	out, diagnostic := &stream{}, &stream{}
	first := []byte{' ', 0, '\n', 0xff}
	if _, callErr := out.Write(first); callErr != nil {
		t.Fatal(callErr)
	}
	if _, callErr := diagnostic.Write([]byte("err\n")); callErr != nil {
		t.Fatal(callErr)
	}
	snapshot := out.Bytes()
	if _, callErr := out.Write([]byte("tail")); callErr != nil {
		t.Fatal(callErr)
	}
	snapshot[0] = 'x'
	if !bytes.Equal(out.Bytes(), append(first, []byte("tail")...)) {
		t.Fatalf("stdout corrupted %q", out.Bytes())
	}
	if string(diagnostic.Bytes()) != "err\n" {
		t.Fatal("streams mixed")
	}
}
