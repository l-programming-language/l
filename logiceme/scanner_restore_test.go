package Logiceme

import (
	"bytes"
	"testing"
)

func TestRestoreDefinitionsRetainsOrderAndOwnership(t *testing.T) {
	source := []byte("A A\xff\r\n")
	for _, tc := range []struct {
		definitions []DefinitionStore
		want        string
	}{
		{nil, "A A\xff\r\n"},
		{[]DefinitionStore{{[]byte("missing"), []byte("unused")}}, "A A\xff\r\n"},
		{[]DefinitionStore{{[]byte("A"), []byte("B")}, {[]byte("B"), []byte("C")}}, "C A\xff\r\n"},
	} {
		input := bytes.Clone(source)
		got := RestoreDefinitionBytes(tc.definitions, input)
		if string(got) != tc.want {
			t.Fatalf("restored bytes = %q, want %q", got, tc.want)
		}
		got[0] = '!'
		if (input[0] == '!') != (len(tc.definitions) == 0) {
			t.Fatal("definition restoration changed input ownership")
		}
	}
}
