package Lyger

// op is an AML structure type abstracting an Operation to generate source code for
type op struct {
	ID          string
	Prefix      string
	Type        string
	Operands    []string
	Operator    string
	Postfix     string
	AML         map[string]string // {"open": "Op{", "close": "}Op"}
	Compilation []byte
}

// prop is an AML structure type abstracting a Property to generate source code for
type prop struct {
	ID          string
	Key         string
	Type        string
	Value       string
	StructTags  string
	AML         map[string]string // {"open": "Prop{", "close": "}Prop"}
	Compilation []byte
}

// funct is an AML structure type abstracting a Function to generate source code for
type funct struct {
	ID          string
	Key         string
	Inputs      []*prop
	Outputs     []string
	Receiver    *prop
	AML         map[string]string // {"open": "Func{", "close": "}Func"}
	Compilation []byte
}

// def encapsulates and structures logeme definitions associated with Attributes and Objects
type def struct {
	ID          string
	Text        string
	Type        string // Syllable (`XY`), Word (`[X, Y]`), Sentence (`[[X],[Y]]`)
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
}

// attribute structures an abstracted part of an Object
type attribute struct {
	ID          string
	Definition  *def
	Children    []*attribute
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
}

// object structures a complete abstract file of code with attributes to be assembled by Logeme downstream
type object struct {
	Module      *sourceModule
	ID          string
	Meta        map[string]string
	Definition  *def
	Attributes  []*attribute
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
}

// class groups Objects and recursively nested Classes in source order
type class struct {
	Root        string
	Module      *sourceModule
	ID          string
	Meta        map[string]string
	Objects     []*object
	Children    []*class
	members     []sourceMember
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
}

// service structures a set of Classes and high-level Objects
type service struct {
	Target      *sourceTarget
	ID          string
	Meta        map[string]string
	Classes     []*class
	Objects     []*object
	Children    []*service
	members     []sourceMember
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
}

// App structures an entire functional Application for generation by Logeme and execution by Lyger
type App struct {
	ID          string
	buildID     string
	Status      string
	Meta        map[string]string
	Settings    map[string]string
	Services    []*service
	Objects     []*object
	AML         map[string]string
	Params      map[string]string
	Compilation []byte
	Output      string
	Error       string
}

func (a *App) buildName() string {
	if a.buildID != "" {
		return a.buildID
	}
	return a.ID
}
