package main

import (
	"reflect"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func TestContractsProfileRequestsAgree(t *testing.T) {
	input := contractCLIInput(t)
	for _, name := range []string{"NoHeapAllocation", "NoBlocking", "NoExternalIO", "NoPanic", "OwnedLifecycle"} {
		t.Run(name, func(t *testing.T) {
			flag := scopedCLIReport(t, input, 1, "check", "--profile", name)
			document := contracts.CheckRequests{SchemaVersion: contracts.SchemaVersion, Kind: "contract_check_requests", Producer: "test", Version: "1", Checks: []contracts.CheckRequest{{ID: "whole", Scope: "whole_entry", Profile: contracts.Profile{Name: name}}}}
			path := scopedWrite(t, t.TempDir(), "requests.json", document)
			request := scopedCLIReport(t, input, 1, "check", "--request-file", path)
			if len(flag.Profiles) != 1 || len(request.Requests) != 1 || flag.Profiles[0].Profile.Scope != "whole_region" || flag.Profiles[0].Status != request.Requests[0].Status {
				t.Fatalf("flag/request disagreement: %+v / %+v", flag.Profiles, request.Requests)
			}
			obligations := request.Requests[0].Obligations
			if !reflect.DeepEqual(flag.Profiles[0].Obligations, obligations[len(obligations)-len(flag.Profiles[0].Obligations):]) {
				t.Fatal("equivalent requests changed profile claims, assumptions or reasons")
			}
			document.Checks[0].Profile.Scope = "whole-region"
			if err := document.Validate(); err == nil {
				t.Fatal("obsolete scope spelling accepted")
			}
		})
	}
	if _, err := parseContractsCommand([]string{"check", "--profile", "PreservesOnFailure", input}); err == nil {
		t.Fatal("failure profile accepted without explicit request parameters")
	}
	if _, err := parseContractsCommand([]string{"check", "--profile", "unknown", input}); err == nil {
		t.Fatal("unknown profile accepted")
	}
}
