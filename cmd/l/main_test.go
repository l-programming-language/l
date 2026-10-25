package main

import (
	"flag"

	//"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

func Test_getCliData(t *testing.T) {
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string   // The name of the test
		want    *cliData // What inputFile instance we want our function to return.
		wantErr bool     // whether we want an error.
		osArgs  []string // The command arguments used for this test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{"No parameters", &cliData{"build", "main.l", "ubuntu"}, false, []string{"cmd", "build"}},
		{"Main file parameter", &cliData{"build", "./test/main.l", "ubuntu"}, false, []string{"cmd", "build", "./test/main.l"}},
	}
	// Iterating over the previous test slice
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Saving the original os.Args reference
			actualOsArgs := os.Args
			// defer function will run after the test is done
			defer func() {
				os.Args = actualOsArgs                                           // Restoring the original os.Args reference
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Resetting the Flag command line. So that we can parse flags again
			}()

			os.Args = tt.osArgs // Setting the specific command args for this test
			command, err := parseCommand(os.Args[1:])
			var got *cliData
			if command != nil {
				got = command.CLI
			}
			if (err != nil) != tt.wantErr { // Asserting whether we get the correct error value
				t.Errorf("getCliData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) { // Asserting whether we get the correct wanted value
				t.Errorf("getCliData() = %v, want %v", got, tt.want)
			}
		})
	}
}
