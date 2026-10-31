// Author: Connor Sanders

package Logiceme

import (
	"bytes"
	"strings"
)

type Definitional struct {
	Definition      []byte
	SegmentalSystem SegmentalSystem
	Content         []byte
	Inputs          []string
}

// Compute Syllable Definitional Scalar
// I.E. `BA~` and `O`
// [Determinate, Vowel, Featural-META-Featural, Conjugate]
func (def *Definitional) Compute() error {
	glyph, err := def.SegmentalSystem.initializeGlyphChecked(def.Definition, nil, nil)
	if err != nil {
		return err
	}
	if err := glyph.Compile(); err != nil {
		return err
	}
	def.Content = glyph.Content
	return nil
}

// Syllable
// Logiceme of Glyphs
type Syllable struct {
	Uuid      string
	Name      string
	Type      string
	Processor Processor
	Meta      MetaDict
	Tensor    Tensor
	Content   []byte
	Count     int
	Exponent  int
	Polarity  bool
	// Rust EA forbids even explicitly written identity modifiers, which the
	// public numeric fields alone cannot distinguish from an unmodified form.
	hasExplicitModifiers bool
}

// Build Syllable
//
//	[`BA~`, `I~Hello World!~`, `-BA~`, `BA~`]
func (syl *Syllable) Build() {
	tensor := Tensor{
		Name:        syl.Name,
		Type:        "Scalar",
		Subject:     []byte(syl.Name),
		SubjectType: "Syllable",
		Order:       1,
		Count:       syl.Count,
		Exponent:    syl.Exponent,
		Polarity:    syl.Polarity,
	}
	proc := Processor{
		Name:       syl.Name,
		Type:       "Syllable",
		Definition: []byte(syl.Name),
		Count:      syl.Count,
		Exponent:   syl.Exponent,
		Polarity:   syl.Polarity,
	}
	syl.Tensor = tensor
	syl.Processor = proc
}

// Word
// Attribute/Resource
type Word struct {
	Uuid      string
	Name      string
	Type      string
	Syllables []Syllable
	Meta      MetaDict
	Tensor    Tensor
	Content   []byte
}

// Build Word
func (w *Word) Build() {
	syllables := make([]Syllable, len(w.Syllables))
	wTensor := Tensor{
		Name:        w.Name,
		Type:        "Vector",
		Subject:     []byte(w.Name),
		SubjectType: "Word",
		Order:       2,
		Dimensions:  make([]Tensor, len(w.Syllables)),
		Count:       1,
		Exponent:    1,
		Polarity:    false,
	}
	for index, syllable := range w.Syllables {
		syllable.Build()
		wTensor.Dimensions[index] = syllable.Tensor
		syllables[index] = syllable
	}
	w.Syllables = syllables
	w.Tensor = wTensor
}

// Sentence
// Sentence = Object
type Sentence struct {
	Uuid            string   `json:"Uuid,omitempty"`
	Name            string   `json:"Name,omitempty"`
	Type            string   `json:"Type,omitempty"`
	Words           []Word   `json:"Words,omitempty"`
	Meta            MetaDict `json:"Meta,omitempty"`
	Dependencies    []string `json:"Dependencies,omitempty"`
	EnvDependencies []string `json:"EnvDependencies,omitempty"`
	Tensor          Tensor   `json:"Tensor,omitempty"`
	Content         []byte   `json:"Content,omitempty"`
}

func compileDependenceMeta(cleanedDepend string, mDepStrs []string, mType string) (string, []string) {
	var mDeps []string
	for _, mDepStr := range mDepStrs {
		if mType == "Env" {
			cleanedEnvStr := strings.Replace(strings.Replace(strings.Replace(mDepStr, "<E{", "", 1), "}E>", "", 1), " ", "", -1)
			sEnvStrs := strings.Split(cleanedEnvStr, "|")
			mDeps = append(mDeps, sEnvStrs...)
		} else if mType == "Alias" {
			cleanedEnvStr := strings.Replace(strings.Replace(strings.Replace(mDepStr, "<A{", "", 1), "}A>", "", 1), " ", "", -1)
			mDeps = append(mDeps, cleanedEnvStr)
		}
		cleanedDepend = strings.Replace(cleanedDepend, mDepStr, "", 1)
	}
	if mType == "Alias" {
		if len(mDeps) > 0 {
			cleanedDepend = mDeps[0] + " " + cleanedDepend
		}
	}
	return cleanedDepend, mDeps
}

func removeDependencyAlias(deps []string) []string {
	var reDeps []string
	for _, dep := range deps {
		if strings.Contains(dep, " ") {
			sDep := strings.Split(dep, " ")
			reDeps = append(reDeps, sDep[1])
		} else {
			reDeps = append(reDeps, dep)
		}
	}
	return reDeps
}

func (s *Sentence) compileDependencies() {
	importStrs := lucidParser(string(s.Content), "Imports")
	for _, importStr := range importStrs {
		updatedImportStr := importStr
		dependencies := lucidParser(importStr, "Dependencies")
		for _, depend := range dependencies {
			var eDeps []string
			// TODO: ADD Internal Package Links, I.E. !! instead of @@, Here as well!
			cleanedDepend := strings.Replace(depend, "@@", "", -1)
			alias := lucidParser(cleanedDepend, "Alias")
			envDeps := lucidParser(cleanedDepend, "EnvDependencies")
			cleanedDepend, eDeps = compileDependenceMeta(cleanedDepend, envDeps, "Env")
			cleanedDepend, _ = compileDependenceMeta(cleanedDepend, alias, "Alias")
			s.Dependencies = append(s.Dependencies, cleanedDepend)
			s.EnvDependencies = append(s.EnvDependencies, eDeps...)
			updatedImportStr = strings.Replace(updatedImportStr, depend, cleanedDepend, -1)
		}
		updatedImportStr = strings.Replace(updatedImportStr, "@", "", -1)
		s.Content = bytes.Replace(s.Content, []byte(importStr), []byte(updatedImportStr), 1)
	}
	removeDuplicateHeaders(&s.Dependencies)
	removeDuplicateHeaders(&s.EnvDependencies)
	s.Dependencies = removeDependencyAlias(s.Dependencies)
}

// Compile Sentence
func (s *Sentence) Assemble(meta MetaDict) error {
	combinedMeta := s.Meta.Combine(meta)
	if err := s.Tensor.Compile(); err != nil {
		return err
	}
	s.Meta = combinedMeta
	s.Content = s.Tensor.Content
	return nil
}

// Link Sentence
func (s *Sentence) Link() {
	s.Content = s.Tensor.Content
	s.compileDependencies()
}

// Build Sentence
func (s *Sentence) Compile() {
	words := make([]Word, len(s.Words))
	sTensor := Tensor{
		Name:        s.Name,
		Type:        "Matrix",
		Subject:     []byte(s.Name),
		SubjectType: "Sentence",
		Order:       3,
		Dimensions:  make([]Tensor, len(s.Words)),
		Count:       1,
		Exponent:    1,
		Polarity:    false,
	}
	for index, word := range s.Words {
		word.Build()
		sTensor.Dimensions[index] = word.Tensor
		words[index] = word
	}
	s.Words = words
	s.Tensor = sTensor
}

// Text = Class
type Text struct {
	Members   []sourceMember
	Name      string
	Outline   MetaDict
	Context   MetaDict
	Sentences []Sentence
	SubTexts  []Text
	Content   [][]byte
}

// Chapter = Service
type Chapter struct {
	Members     []sourceMember
	Name        string
	Outline     MetaDict
	Context     MetaDict
	Sentences   []Sentence
	Texts       []Text
	SubChapters []Chapter
	Content     [][]byte
}

// Book = App
type Book struct {
	Name         string
	Program      []byte
	Meta         MetaDict
	Settings     MetaDict
	Outline      Outline
	Content      []Chapter
	DataSources  []DataSource
	Integrations []Integration
}
