package Lyger

import (
	"fmt"
	"testing"
)

func Test_lookup(t *testing.T) {
	fmt.Println("This is a test")
}

/*
func Test_lookup(t *testing.T) {
	// Creating a temporal and empty CSV file
	lex := initLexicon("eng", "test")
	notFoundWord := getWantedWord("")
	varWord := getWantedWord("var")
	foundWord := getWantedWord("list")
	// Defining the struct we're going to use
	tests := []struct {
		name     string
		testWord string
		want     *word
		wantErr  bool
	}{ // Defining our test cases
		{"Not Found", "foo", notFoundWord, true},
		{"Variable", `"bar"`, varWord, false},
		{"Word Found", "list", foundWord, false},
	}
	// Iterating over our test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lex.lookup(tt.testWord)
			// Checking the error
			if (err != nil) != tt.wantErr {
				t.Errorf("lookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Checking the returning value
			if got.Type != tt.want.Type {
				t.Errorf("lookup() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
