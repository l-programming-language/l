package Lyger

import (
	"fmt"
	"testing"
)

func Test_initGrammar(t *testing.T) {
	fmt.Println("This is a test")
}

/*
func Test_checkStatement(t *testing.T) {
	// Creating a temporal and empty CSV file
	gram := initGrammar(Options{Language: "eng"})
	passiveTokens := getTestTokens(gram, "passive")
	passiveStatement := getWantedCheckStatement("passive")
	//createTokens := getTestTokens(gram,"create")
	//defineTokens := getTestTokens(gram,"define")
	// Defining the struct we're going to use
	tests := []struct {
		name    string
		tokens  []*token
		want    *checkedStatement
		wantErr bool
	}{ // Defining our test cases
		{"Passive", passiveTokens, passiveStatement, false},
		//{"Create", `"bar"`, createTokens, false},
		//{"Define", "list", defineTokens, false},
	}
	// Iterating over our test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gram.checkStatement(tt.tokens)
			// Checking the error
			if (err != nil) != tt.wantErr {
				t.Errorf("checkStatement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Checking the returning value
			if got.category != tt.want.category {
				t.Errorf("checkStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
