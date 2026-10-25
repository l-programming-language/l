package utilities

import (
	"os"
	"testing"
)

func Test_checkIfValidFile(t *testing.T) {
	// Creating a temporal and empty CSV file
	tmpfile, err := os.CreateTemp("", "main*.l")
	if err != nil {
		panic(err) // This should never happen
	}
	// Once all the tests are done. We delete the temporal file
	t.Cleanup(func() {
		if err := tmpfile.Close(); err != nil {
			t.Errorf("close temporary file: %v", err)
		}
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("remove temporary file: %v", err)
		}
	})
	// Defining the struct we're going to use
	tests := []struct {
		name     string
		filename string
		want     bool
		wantErr  bool
	}{ // Defining our test cases
		{"File does exist", tmpfile.Name(), true, false},
		{"File does not exist", "nowhere/main.l", false, true},
		{"File is not l", "main.txt", false, true},
	}
	// Iterating over our test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckIfValidFile(tt.filename)
			// Checking the error
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIfValidFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Checking the returning value
			if got != tt.want {
				t.Errorf("CheckIfValidFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
