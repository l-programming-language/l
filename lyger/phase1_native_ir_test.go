package Lyger

import (
	"fmt"
	"reflect"
	"testing"
)

// Native coordination and source-action coordination have different ordering:
// the former applies one operation across operands; the latter runs the whole
// action body for each operand. These cases exercise both grammar and binding.
func TestPhase1NativeLifecycleCoordination(t *testing.T) {
	const apps = "Use IO's Print from `io` as \"Print\".\n\"message\" is `child`.\nCreate \"First\" as a New Basic App.\nThe App Prints message.\nCreate \"Second\" as a New Basic App.\nThe App Prints message.\nUse First as \"Job\".\n"
	for _, operation := range []struct{ root, kind string }{{"Build", "build"}, {"Run", "run"}, {"Stop", "stop"}, {"Await", "await"}} {
		t.Run(operation.root, func(t *testing.T) {
			r := phase1Require(t, apps+operation.root+" Second and Job.\n", nil)
			ops := phase1Lifecycle(t, r, []string{operation.kind, operation.kind})
			if phase1Canonical(t, r, ops[0]).ID != "Second" || phase1Canonical(t, r, ops[1]).ID != "First" {
				t.Fatal("coordinated targets lost source order or canonical alias identity")
			}
		})
	}
	for _, action := range []struct {
		root  string
		kinds []string
	}{
		{"Start", []string{"build", "run", "build", "run"}},
		{"Execute", []string{"build", "run", "await", "build", "run", "await"}},
	} {
		t.Run(action.root, func(t *testing.T) {
			r := phase1Require(t, apps+action.root+" Second and Job.\n", nil)
			ops := phase1Lifecycle(t, r, action.kinds)
			for index, op := range ops {
				want := "Second"
				if index >= len(ops)/2 {
					want = "First"
				}
				if phase1Canonical(t, r, op).ID != want {
					t.Fatalf("source action operand %d did not retain the complete per-App body", index)
				}
			}
		})
	}
}

func TestPhase1PassiveLifecycleTargets(t *testing.T) {
	for _, operation := range []struct{ form, kind string }{{"built", "build"}, {"run", "run"}, {"stopped", "stop"}, {"awaited", "await"}} {
		t.Run(operation.form, func(t *testing.T) {
			r := phase1Require(t, phase1Worker+fmt.Sprintf("Worker is %s.\n", operation.form), nil)
			ops := phase1Lifecycle(t, r, []string{operation.kind})
			if phase1Canonical(t, r, ops[0]).ID != "Worker" {
				t.Fatal("passive operation lost its grammatical subject target")
			}
		})
	}
}

func TestPhase1PassiveAwaitRetainsFollowingEffects(t *testing.T) {
	for _, tail := range []struct {
		name, source string
		kinds        []string
	}{
		{"following-imperative", "Worker is awaited and then show `joined|`.\n", []string{"await", "show"}},
		{"following-source-action", "Worker is awaited and then Start Worker.\n", []string{"await", "build", "run"}},
		{"following-passive", "Worker is awaited and then stopped.\n", []string{"await", "stop"}},
	} {
		t.Run(tail.name, func(t *testing.T) {
			r := phase1Require(t, phase1Worker+tail.source, nil)
			got := phase1Kinds(r.Compilation.Semantics.Instructions)
			if !reflect.DeepEqual(got, tail.kinds) {
				t.Fatalf("passive Await effects=%v want %v", got, tail.kinds)
			}
		})
	}
}
