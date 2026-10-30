package Lyger

// These structural payloads retain distinct pointer identities. Operation and
// property bindings live in selected definitions, not a second AML schema here.
type op struct{ _ byte }
type prop struct{ _ byte }

// funct retains the resolved source declaration behind a Function value.
type funct struct {
	Source *sourceFunction
}

// def encapsulates and structures logeme definitions associated with Attributes and Objects
type def struct {
	ID          string
	Text        string
	Type        string // Syllable (`XY`), Word (`[X, Y]`), Sentence (`[[X],[Y]]`)
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
