/*
Author: Connor Sanders
Version: 0.0.1
Released: 10/15/2020

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
		Logiceme Evaluator V0.0.1
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package Logiceme

// Segmental
type Segmental struct {
	Uuid        string
	Symbol      string
	Type        string
	SubSegments []Segmental
	Segment     Segment
	Scalars     []Scalar
	Content     [][]byte
	Vector      Vector
	Output      []byte
	Count       int
	Exponent    int
	Polarity    bool
}

// Process
type Process struct {
	Uuid         string
	Operands     []string
	Operators    []string
	Segments     []Segmental
	SubProcesses []Process
	Vector       Vector
	Output       []byte
	Count        int
	Exponent     int
	Polarity     bool
}

// Processor
type Processor struct {
	Uuid       string
	Name       string
	Type       string
	Definition []byte
	Processes  []Process
	Operands   []string
	Operators  []string
	Matrix     Matrix
	Output     []byte
	Count      int
	Exponent   int
	Polarity   bool
}
