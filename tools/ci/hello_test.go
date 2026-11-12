package main

import (
	"strings"
	"testing"
)

func TestHelloOutcomesRejectIncompleteEvidence(t *testing.T) {
	var events []nativeEvent
	for _, group := range []struct {
		pkg   string
		tests []string
	}{
		{"cmd/l", []string{"TestCoreHelloWorld", "TestCoreHelloWorldRejectsRuntimeEffects", "TestCoreHelloWorldUsesSelectedPrint", "TestCompilerInitialize", "TestCompilerCompile"}},
		{"lyger", []string{"TestDemoCompilationRetainsSemanticProgram"}},
		{"logiceme", []string{"TestLucidTemplatesFrozen"}},
	} {
		pkg := "github.com/l-programming-language/l/" + group.pkg
		events = append(events, nativeEvent{Action: "start", Package: pkg})
		for _, name := range group.tests {
			events = append(events, nativeEvent{Action: "run", Package: pkg, Test: name},
				nativeEvent{Action: "run", Package: pkg, Test: name + "/case"},
				nativeEvent{Action: "pass", Package: pkg, Test: name + "/case"},
				nativeEvent{Action: "pass", Package: pkg, Test: name})
		}
		events = append(events, nativeEvent{Action: "pass", Package: pkg})
	}
	var good []byte
	for _, event := range events {
		good = append(good, testJSON(t, event)...)
		good = append(good, '\n')
	}
	if err := validateHelloOutcomes(good); err != nil {
		t.Fatal(err)
	}
	firstRun := string(testJSON(t, events[1])) + "\n"
	childPass := string(testJSON(t, events[3])) + "\n"
	parentPass := string(testJSON(t, events[4])) + "\n"
	for name, bad := range map[string]string{
		"empty":                   "",
		"malformed":               "not json",
		"truncated":               string(good[:len(good)-3]),
		"missing root":            strings.ReplaceAll(string(good), "TestCompilerCompile", "UnrequestedTest"),
		"pass without run":        strings.Replace(string(good), firstRun, "", 1),
		"duplicate test":          strings.Replace(string(good), firstRun, firstRun+firstRun, 1),
		"skip":                    strings.Replace(string(good), `"Action":"pass"`, `"Action":"skip"`, 1),
		"fail":                    strings.Replace(string(good), `"Action":"pass"`, `"Action":"fail"`, 1),
		"unfinished child":        strings.Replace(string(good), childPass, "", 1),
		"parent before child":     strings.Replace(string(good), childPass+parentPass, parentPass+childPass, 1),
		"after package completed": string(good) + string(testJSON(t, events[len(events)-1])),
		"no matching tests":       string(testJSON(t, events[0])) + string(testJSON(t, nativeEvent{Action: "pass", Package: events[0].Package})),
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateHelloOutcomes([]byte(bad)); err == nil {
				t.Fatal("incomplete or invalid hello evidence accepted")
			}
		})
	}
}
