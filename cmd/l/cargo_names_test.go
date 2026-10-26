package main

import (
	"strings"
	"testing"
)

func TestOrchestratorPackageName(t *testing.T) {
	for _, reserved := range []string{"build", "deps", "examples", "incremental"} {
		for _, stem := range []string{reserved, strings.ToUpper(reserved), strings.ToUpper(reserved[:1]) + reserved[1:]} {
			t.Run("reserved/"+stem, func(t *testing.T) {
				if got := orchestratorPackageName(stem, nil); got != "l_entry" {
					t.Fatalf("reserved stem %q produced Cargo target %q; want l_entry", stem, got)
				}
			})
		}
	}
	for _, tc := range []struct {
		name, stem, want string
		packages         []string
	}{
		{name: "normal", stem: "main", want: "main"},
		{name: "normal case", stem: "Worker", want: "Worker"},
		{name: "hyphen", stem: "worker-project", want: "worker-project"},
		{name: "underscore", stem: "_entry", want: "_entry"},
		{name: "reserved prefix", stem: "build_app", want: "build_app"},
		{name: "reserved suffix", stem: "deps2", want: "deps2"},
		{name: "empty", want: "l_entry"},
		{name: "spaces", stem: "worker project", want: "l_entry"},
		{name: "leading hyphen", stem: "-worker", want: "l_entry"},
		{name: "leading digit", stem: "42worker", want: "l_entry"},
		{name: "non ASCII", stem: "日本", want: "l_entry"},
		{name: "package collision", stem: "worker", packages: []string{"worker"}, want: "l_entry"},
		{name: "case folded package collision", stem: "Worker", packages: []string{"WORKER"}, want: "l_entry"},
		{name: "reserved fallback occupied", stem: "build", packages: []string{"L_ENTRY", "l_entry_1"}, want: "l_entry_2"},
		{name: "invalid fallback occupied", stem: "worker project", packages: []string{"l_entry", "L_ENTRY_1"}, want: "l_entry_2"},
		{name: "package fallback occupied", stem: "worker", packages: []string{"worker", "l_entry", "l_entry_1"}, want: "l_entry_2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var packages []*sourceBundle
			for _, name := range tc.packages {
				packages = append(packages, &sourceBundle{packagePlan: &cargoPackage{name: name}})
			}
			if got := orchestratorPackageName(tc.stem, packages); got != tc.want {
				t.Fatalf("stem %q with packages %q produced %q; want %q", tc.stem, tc.packages, got, tc.want)
			}
		})
	}
}

func TestOrchestratorPackageNameReservesBinaryTargets(t *testing.T) {
	for _, tc := range []struct {
		name, stem, want string
		packages         []*cargoPackage
	}{
		{"explicit target", "entry", "l_entry", []*cargoPackage{{name: "Worker", binary: "entry"}}},
		{"case equivalent target", "ENTRY", "l_entry", []*cargoPackage{{name: "Worker", binary: "entry"}}},
		{"occupied target fallbacks", "build", "l_entry_2", []*cargoPackage{{name: "First", binary: "L_ENTRY"}, {name: "Second", binary: "l_entry_1"}}},
		{"distinct hyphen and underscore", "work-er", "work-er", []*cargoPackage{{name: "Worker", binary: "work_er"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var packages []*sourceBundle
			for _, p := range tc.packages {
				packages = append(packages, &sourceBundle{packagePlan: p})
			}
			if got := orchestratorPackageName(tc.stem, packages); got != tc.want {
				t.Fatalf("stem %q produced %q; want %q", tc.stem, got, tc.want)
			}
		})
	}
}
