package Lyger

// amlTags contains the AML tag content information need to generate headers for each AML Element type when compiling
type amlTags struct {
	Name     string            // Tag label of Name, i.e. "HW_App"
	Type     string            // Tag Type, i.e. BOOK, SERVICE, S, W, JSON, etc.
	Inputs   map[string][]byte // {"x": []byte("HW_App")}
	Contents [][]byte          // [OpenTag, CloseTag], i.e. [`<BOOK "x">`, `<BOOK/>`]
}

// aml s - syllabic sequence is an element containing the abstract structure for any AML "S" tag
type s struct {
	Name          string
	Type          string
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	Definition    *def               // Logeme definition from semantic lowering.
	CompiledBytes []byte
}

// aml word is an element containing the abstract structure for any AML "W" tag
type w struct {
	Name          string
	Type          string
	Syllables     []*s
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	Definition    *def               // Logeme definition from semantic lowering.
	CompiledBytes []byte
}

// aml sentence is an element containing the abstract structure for any AML "object" tag
type sentence struct {
	Name          string
	Type          string
	Words         []*w
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	CompiledBytes []byte
}

// aml text is an element containing the abstract structure for any AML "class" tag
type text struct {
	Name          string
	Type          string
	sentences     []*sentence
	children      []*text
	members       []amlSourceMember
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	CompiledBytes []byte
}

// aml chapter is an element containing the abstract structure for any AML "service" tag
type chapter struct {
	Name          string
	Type          string
	Texts         []*text
	sentences     []*sentence
	children      []*chapter
	members       []amlSourceMember
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	CompiledBytes []byte
}

// aml book is an element containing the abstract structure for any AML "app" tag
type book struct {
	Name          string
	Type          string
	Chapters      []*chapter
	Sentences     []*sentence
	Tags          *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache         map[string]*record // Source bindings shared with semantic memory.
	CompiledBytes []byte
}

// Program structures and processes a map of analysed statements into AML data for further assembly into source code
type Program struct {
	Name         string
	Type         string
	Tags         *amlTags           // &amlTags{Name: "syllables", Type: "S"...}
	Cache        map[string]*record // Source bindings shared with semantic memory.
	Books        []*book
	AMLBytes     []byte
	BackendBytes []byte
	Apps         []*App
}

// Make renders both AML views from the application books and metadata.
func (p *Program) Make() error {
	observable, backend, err := renderProgram(p, p.Apps)
	if err != nil {
		return err
	}
	p.AMLBytes, p.BackendBytes = observable, backend
	return nil
}
