package Lyger

import "testing"

func TestApplicationUnitsExcludeStructuralServices(t *testing.T) {
	result := requireSourceCompilation(t, "Create \"Host\" as a New Basic App.\nCreate \"Unused\" as a New Main Service.\nCreate \"First\" as a New Basic App.\nCreate \"Library\" as a New Service.\nRun First and Host.\nAwait Host and First.\n", nil)
	units, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].ID != "Host" || units[1].ID != "First" {
		t.Fatalf("structural Services changed the App graph: %+v", units)
	}
}
