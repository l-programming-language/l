package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func TestStructuralServiceParagraphOwnership(t *testing.T) {
	for _, descriptor := range []string{"", "Main "} {
		for _, tc := range []struct {
			name, source        string
			apps, continuations int
			trace               []string
		}{
			{
				"Service-led ordinary paragraph",
				"    Create \"Library\" as a New Service.\nShow `visible`.\n",
				0, 0, []string{"show:visible"},
			},
			{
				"Service alias retains ordinary effects",
				"    Create \"Library\" as a New Service.\nUse Library as \"Alias\".\nShow `visible`.\n",
				0, 0, []string{"show:visible"},
			},
			{
				"Service keeps App continuation owner",
				"    Create \"Worker\" as a New Public Basic App.\nCreate \"Library\" as a New Service.\nShow `visible`.\nExit.\n\nRun Worker.\nAwait Worker.\nShow `unreachable`.\n",
				1, 1, []string{"run:Worker", "await:Worker", "show:visible", "exit"},
			},
			{
				"Service alias keeps App continuation owner",
				"    Create \"Worker\" as a New Public Basic App.\nCreate \"Library\" as a New Service.\nUse Library as \"Alias\".\nShow `visible`.\nExit.\n\nRun Worker.\nAwait Worker.\nShow `unreachable`.\n",
				1, 1, []string{"run:Worker", "await:Worker", "show:visible", "exit"},
			},
			{
				"later App does not change paragraph opener",
				"    Create \"Library\" as a New Service.\nCreate \"Worker\" as a New Basic App.\nShow `visible`.\n",
				1, 0, []string{"show:visible"},
			},
		} {
			t.Run(descriptor+tc.name, func(t *testing.T) {
				source := strings.ReplaceAll(tc.source, "New Service", "New "+descriptor+"Service")
				result := runnableServiceRequire(t, source, nil)
				plan, err := result.WorkspacePlan()
				if err != nil || len(plan.Packages()) != tc.apps || len(result.Compilation.Program.Apps) != tc.apps {
					t.Fatalf("structural Service changed App package inventory: %v", err)
				}
				if got := paragraphActivationTrace(t, result); !reflect.DeepEqual(got, tc.trace) {
					t.Errorf("structural declaration changed executable paragraph ownership: %v; want %v", got, tc.trace)
				}
				if len(result.Compilation.Semantics.Continuations) != tc.continuations {
					t.Fatal("structural declaration changed App continuation ownership")
				}
			})
		}
	}
}
