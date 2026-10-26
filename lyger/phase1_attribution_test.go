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
		for _, call := range nativetest.RustCalls(fn.Body) {
			if call.Name != "set_source" {
				continue
			}
			if len(call.Args) != 1 || len(call.Args[0]) != 1 {
				t.Fatal("source attribution requires one literal location")
			}
			location, err := nativetest.DecodeRustLiteral(call.Args[0][0].Text)
			if err != nil {
				t.Fatal(err)
			}
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
