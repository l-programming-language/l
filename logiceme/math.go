// Author: Connor Sanders

package Logiceme

// Tensor represents a linear object upon which the evaluation can be conducted.
type Tensor struct {
	Name            string
	Type            string
	SegmentalSystem SegmentalSystem
	Subject         []byte
	SubjectType     string
	Order           int
	Content         []byte
	ContentType     string
	Dimensions      []Tensor
	Count           int
	Exponent        int
	Polarity        bool
}

// Compile a Tensor
func (t *Tensor) Compile() error {
	t.Content = nil
	compiledDimensions := []Tensor{}
	if t.Order == 1 && t.Type == "Scalar" {
		// Compile Scalar Bytes
		sylDef := Definitional{
			Definition:      t.Subject,
			SegmentalSystem: t.SegmentalSystem,
			Content:         []byte(""),
			Inputs:          []string{""},
		}
		if err := sylDef.Compute(); err != nil {
			return err
		}
		t.Content = append(t.Content, sylDef.Content...)
	} else {
		// Recursively Compile SubTensors
		for _, dimension := range t.Dimensions {
			dimension.SegmentalSystem = t.SegmentalSystem
			if err := dimension.Compile(); err != nil {
				return err
			}
			compiledDimensions = append(compiledDimensions, dimension)
			t.Content = append(t.Content, dimension.Content...)
		}
		t.Dimensions = compiledDimensions
	}
	return nil
}

// Scalar 1-Tensor
type Scalar struct {
	Name         string
	Type         string
	Uuid         string
	Description  string
	ScalarTensor Tensor
}

// Vector 2-Tensor
type Vector struct {
	Name         string
	Type         string
	Uuid         string
	Description  string
	VectorTensor Tensor
}

// Matrix 3-Tensor
type Matrix struct {
	Name         string
	Type         string
	Uuid         string
	Description  string
	MatrixTensor Tensor
}
