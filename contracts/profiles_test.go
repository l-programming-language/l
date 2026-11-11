package contracts

import "testing"

func TestProfileDescriptorEnforcement(t *testing.T) {
	for name, kind := range map[string]EffectKind{"NoHeapAllocation": "allocate", "NoBlocking": "block", "NoExternalIO": "io", "NoPanic": "panic"} {
		t.Run(name, func(t *testing.T) {
			profile := Profile{Name: name}
			summary := DerivedSummary{Paths: []PathSummary{{Outcome: Normal, Reachable: true}}}
			check := func(want Status) {
				t.Helper()
				if got := CheckProfile(summary, profile)[0]; got.Status != want {
					t.Fatalf("%s: want %s, got %+v", name, want, got)
				}
			}
			check(Established)
			profile.Scope = "source_operation"
			check(Established)
			request := CheckRequests{SchemaVersion: SchemaVersion, Kind: "contract_check_requests", Producer: "test", Version: "1", Checks: []CheckRequest{{ID: "operation", Scope: "published-operation", Profile: profile}}}
			if err := request.Validate(); err != nil {
				t.Fatalf("published operation boundary rejected: %v", err)
			}
			summary.Paths[0].Effects = []Effect{{Kind: kind, Definite: true}}
			check(Contradicted)
			summary.Paths[0].Effects = nil
			summary.Paths[0].Unknowns = []UnknownReason{{Components: []string{"effects"}}}
			check(Unresolved)
			summary.Paths[0].Unknowns[0].ExcludedEffects = []EffectKind{kind}
			check(Established)
			summary.Paths[0].Unknowns[0].Components = []string{"outcomes"}
			check(Unresolved)
			summary.Paths[0].Unknowns = nil
			profile.Scope = "whole-region"
			check(Unresolved)
		})
	}
	for _, name := range []string{"PreservesOnFailure", "unknown"} {
		if AllEstablished(CheckProfile(DerivedSummary{Paths: []PathSummary{{Outcome: Normal}}}, Profile{Name: name})) {
			t.Fatalf("%s bypassed its dedicated checker", name)
		}
	}
}
