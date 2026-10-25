package Lyger

import (
	"reflect"
	"testing"
)

func Test_importScan(t *testing.T) {
	tConfig := canonicalTestOptions()
	gram, err := loadGrammar(tConfig)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		content []byte
		want    []*importStatement
		wantErr bool
	}{ // Defining our test cases
		{"Basic", []byte(TEST_MAIN_BYTES), getWantedImportStatements("basic"), false},
	}
	// Iterating over our test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepared, err := PrepareSource("import-scan.l", tt.content)
			var got []*importStatement
			if err == nil {
				for _, statement := range prepared.Statements {
					var imported *importStatement
					var isImport bool
					isImport, imported, err = gram.isImportStatement(statement.Raw)
					if err != nil {
						break
					}
					if isImport {
						got = append(got, imported)
					}
				}
			}
			// Checking the error
			if (err != nil) != tt.wantErr {
				t.Errorf("scanForImports() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) { // Asserting whether we get the correct wanted value
				t.Errorf("scanForImports() = %v, want %v", got, tt.want)
			}
		})
	}
}
