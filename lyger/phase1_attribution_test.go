package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"strconv"
	"strings"
	"testing"
)

func TestPhase1RuntimeSourceAttribution(t *testing.T) {
	source := phase1Worker + "Use Worker as \"Job\".\nStart Job.\nUse `effect.l`.\n"
	result := phase1Require(t, source, map[string]string{"effect.l": "Await Job.\nShow Worker's Output.\n"})
	generated, _ := phase1Generated(t, result)
	var locations []string
	for _, fn := range generated.Functions() {
		locationsByBinding := make(map[string]string)
		for _, call := range nativetest.RustCalls(fn.Body) {
			// The source-owned attribution word constructs an owned String,
			// then passes that local to the runtime's source-location operation.
			if call.Name == "from" && call.At >= 5 && fn.Body[call.At-2].Text == "String" && fn.Body[call.At-1].Text == "::" && fn.Body[call.At-3].Text == "=" && fn.Body[call.At-5].Text == "let" && len(call.Args) == 1 && len(call.Args[0]) == 1 {
				if location, err := nativetest.DecodeRustLiteral(call.Args[0][0].Text); err == nil {
					locationsByBinding[fn.Body[call.At-4].Text] = location
				}
			}
			if call.Name != "app_set_source" {
				continue
			}
			if len(call.Args) != 2 || len(call.Args[1]) != 1 {
				t.Fatal("source attribution requires a receiver and one location binding")
			}
			name := call.Args[1][0].Text
			location, found := locationsByBinding[name]
			if !found {
				t.Fatal("source attribution did not consume a preceding literal location binding")
			}
			delete(locationsByBinding, name)
			locations = append(locations, location)
		}
	}

	line := strings.Count(source[:strings.Index(source, "Start Job.")], "\n") + 1
	want := []string{
		"/main.l:" + strconv.Itoa(line) + ":1: build Worker",
		"/main.l:" + strconv.Itoa(line) + ":1: run Worker",
		"/effect.l:1:1: await Worker",
		"/effect.l:2:1: show Worker",
	}
	if len(locations) != len(want) {
		t.Fatalf("source locations = %q, want four ordered executing effects", locations)
	}
	for i := range want {
		if !strings.HasSuffix(locations[i], want[i]) {
			t.Errorf("effect %d attribution = %q, want suffix %q", i, locations[i], want[i])
		}
	}
}
