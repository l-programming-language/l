package Lyger

import (
	"errors"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

type EntryData interface {
}

// val is a structure encapsulating an abstracted value
type val struct {
	Type valueType
	Data interface{}
}

// isDef checks if a text primitive is a definition or not
func (v *val) isDef() bool {
	if v.Type != valueText {
		return false
	}
	text, ok := v.Data.(string)
	if !ok {
		return false
	}
	frame, framed := splitDefinitionFrame(text)
	return framed && frame.Quoted
}

// primitive is instantiated to encapsulate any non-lexeme primitive type
type primitive struct {
	ID      string      `json:"id,omitempty"`
	Type    valueType   `json:"type,omitempty"`
	Thing   interface{} `json:"thing,omitempty"`
	Text    string      `json:"text,omitempty"`
	Number  float64     `json:"number,omitempty"`
	Truth   bool        `json:"truth,omitempty"`
	Bytes   []byte      `json:"bytes,omitempty"`
	Pointer string      `json:"pointer,omitempty"`
}

// value reads the selected representation, including valid nil Thing/Bytes data.
func (p *primitive) value() (val, error) {
	if p == nil {
		return val{}, lerrors.Internal.Errorf("cannot read a nil primitive")
	}
	switch p.Type {
	case valueThing:
		return val{Type: p.Type, Data: p.Thing}, nil
	case valueText:
		return val{Type: p.Type, Data: p.Text}, nil
	case valueNumber:
		return val{Type: p.Type, Data: p.Number}, nil
	case valueTruth:
		return val{Type: p.Type, Data: p.Truth}, nil
	case valueBytes:
		return val{Type: p.Type, Data: p.Bytes}, nil
	case valuePointer:
		return val{Type: p.Type, Data: p.Pointer}, nil
	case valueNull:
		return val{Type: p.Type}, nil
	case valueError:
		return val{Type: p.Type, Data: p.Bytes}, nil
	default:
		return val{}, lerrors.Internal.Errorf("primitive %s has no readable value", p.Type)
	}
}

func newPrimitive(kind valueType) *primitive {
	if category, ok := kind.StorageKind(); !ok || category != storagePrimitive {
		kind = valueUnknown
	}
	p := &primitive{Type: kind}
	if kind == valueBytes {
		p.Bytes = []byte{}
	}
	return p
}

// loadPrimitive inputs a byte slice and outputs a new loaded primitive
func loadPrimitive(tBytes []byte) (*primitive, error) {
	var rePrim primitive
	if isText(tBytes) {
		rePrim.Type = valueText
		rePrim.Text = string(tBytes)
	} else if isTruth(tBytes) {
		rePrim.Type = valueTruth
		rePrim.Truth = bytesToBool(tBytes)
	} else if isID(tBytes) {
		rePrim.ID = strings.ToLower(string(tBytes))
		rePrim.Type = valueID
	} else {
		number, err := strconv.ParseFloat(string(tBytes), 64)
		if err == nil {
			rePrim.Type = valueNumber
			rePrim.Number = number
			return &rePrim, nil
		}
		rePrim.Type = valueNone
		if errors.Is(err, strconv.ErrRange) {
			return &rePrim, lerrors.InvalidInput.Errorf("invalid Number %q: %w", tBytes, err)
		}
		return &rePrim, lerrors.InvalidInput.Errorf("%s", "newPrimitive unable to process input of "+string(tBytes))
	}
	return &rePrim, nil
}

// list structures an ordered dynamic array of in determinate length in L
type list struct {
	ID     string
	Values []interface{}
	Types  []string
}

// newList returns a newly initialized valueless list
func newList() *list {
	var v []interface{}
	return &list{
		Values: v,
		Types:  []string{},
	}
}

// dictionary structures a dynamic key-value hashmap in L
type dictionary struct {
	ID     string
	Values map[string]interface{}
	Types  map[string]string
}

// newDict returns a newly initialized valueless dictionary
func newDict() *dictionary {
	return &dictionary{
		Values: make(map[string]interface{}),
		Types:  make(map[string]string),
	}
}

// applicative structures an Application based Entry of a Record
type applicative struct {
	ID         string
	Type       valueType
	App        *App
	Service    *service
	Class      *class
	Object     *object
	Attribute  *attribute
	Definition *def
	Function   *funct
	Property   *prop
	Operation  *op
}

// value retains canonical payload identity; it never copies an application.
func (a *applicative) value() (val, error) {
	if a == nil {
		return val{}, lerrors.Internal.Errorf("cannot read a nil applicative")
	}
	switch a.Type {
	case valueApp:
		if a.App != nil {
			return val{Type: a.Type, Data: a.App}, nil
		}
	case valueService:
		if a.Service != nil {
			return val{Type: a.Type, Data: a.Service}, nil
		}
	case valueClass:
		if a.Class != nil {
			return val{Type: a.Type, Data: a.Class}, nil
		}
	case valueObject:
		if a.Object != nil {
			return val{Type: a.Type, Data: a.Object}, nil
		}
	case valueAttribute:
		if a.Attribute != nil {
			return val{Type: a.Type, Data: a.Attribute}, nil
		}
	case valueDefinition:
		if a.Definition != nil {
			return val{Type: a.Type, Data: a.Definition}, nil
		}
	case valueFunction:
		if a.Function != nil {
			return val{Type: a.Type, Data: a.Function}, nil
		}
	case valueProperty:
		if a.Property != nil {
			return val{Type: a.Type, Data: a.Property}, nil
		}
	case valueOperation:
		if a.Operation != nil {
			return val{Type: a.Type, Data: a.Operation}, nil
		}
	}
	return val{}, lerrors.Internal.Errorf("applicative %s has no readable value", a.Type)
}

// newApplicative returns a new valueless applicative struct initializes for an inputted type
func newApplicative(aType valueType) *applicative {
	a := applicative{Type: aType}
	switch a.Type {
	case valueApp:
		a.App = &App{}
	case valueService:
		a.Service = &service{}
	case valueClass:
		a.Class = &class{}
	case valueObject:
		a.Object = &object{}
	case valueAttribute:
		a.Attribute = &attribute{}
	case valueDefinition:
		a.Definition = &def{}
	case valueFunction:
		a.Function = &funct{}
	case valueProperty:
		a.Property = &prop{}
	case valueOperation:
		a.Operation = &op{}
	default:
		a.Type = valueUnknown
	}
	return &a
}

// iterative structures an iterable L type
type iterative struct {
	ID         string
	Type       valueType
	List       *list
	Dictionary *dictionary
	Map        *dictionary
}

func (i *iterative) value() (val, error) {
	if i == nil {
		return val{}, lerrors.Internal.Errorf("cannot read a nil iterative")
	}
	switch i.Type {
	case valueList:
		if i.List != nil {
			return val{Type: i.Type, Data: i.List}, nil
		}
	case valueMap:
		if i.Map != nil {
			return val{Type: i.Type, Data: i.Map}, nil
		}
	case valueDictionary:
		if i.Dictionary != nil {
			return val{Type: i.Type, Data: i.Dictionary}, nil
		}
	}
	return val{}, lerrors.Internal.Errorf("iterative %s has no readable value", i.Type)
}

// newIterative returns a new valueless iterative struct initializes for an inputted type
func newIterative(iType valueType) *iterative {
	i := iterative{Type: iType}
	switch i.Type {
	case valueList:
		i.List = newList()
	case valueMap:
		i.Map = newDict()
	case valueDictionary:
		i.Dictionary = newDict()
	default:
		i.Type = valueUnknown
	}
	return &i
}

// entry structures the data stored in a Record
type entry struct {
	ID          string
	Type        storageKind
	Primitive   *primitive
	Lexeme      *lexeme
	Applicative *applicative
	Iterative   *iterative
	Runtime     *runtimeReference
}

// value validates storage at the read boundary. Runtime is orthogonal metadata.
func (e *entry) value() (val, error) {
	if e == nil {
		return val{}, lerrors.Internal.Errorf("cannot read a nil entry")
	}
	payloads := 0
	for _, present := range []bool{e.Primitive != nil, e.Lexeme != nil, e.Applicative != nil, e.Iterative != nil} {
		if present {
			payloads++
		}
	}
	if payloads != 1 {
		return val{}, lerrors.Internal.Errorf("entry %q requires one storage payload, got %d", e.ID, payloads)
	}
	switch e.Type {
	case storagePrimitive:
		if e.Primitive != nil {
			return e.Primitive.value()
		}
	case storageLexeme:
		if e.Lexeme != nil {
			return val{Type: valueLexeme, Data: e.Lexeme}, nil
		}
	case storageApplicative:
		if e.Applicative != nil {
			return e.Applicative.value()
		}
	case storageIterative:
		if e.Iterative != nil {
			return e.Iterative.value()
		}
	}
	return val{}, lerrors.Internal.Errorf("entry %q has inconsistent %s storage", e.ID, e.Type)
}

// unresolved identifies valid transitional primitive storage without reading it.
func (e *entry) unresolved() bool {
	if e == nil || e.Type != storagePrimitive || e.Primitive == nil || e.Lexeme != nil || e.Applicative != nil || e.Iterative != nil {
		return false
	}
	switch e.Primitive.Type {
	case valueID, valueNone, valueUnknown:
		return true
	default:
		return false
	}
}

// init accepts an open lexical name or a concrete built-in value name.
func (e *entry) init(name string, category storageKind) {
	e.Type = category
	kind, _ := parseValueType(name)
	switch category {
	case storagePrimitive:
		e.Primitive = newPrimitive(kind)
	case storageApplicative:
		e.Applicative = newApplicative(kind)
	case storageIterative:
		e.Iterative = newIterative(kind)
	case storageLexeme:
		e.Lexeme = newLexeme(name)
	default:
		e.Type = storageUnknown
	}
}

// initEntry returns a newly initialized memory entry
func initEntry(eType string, eCat storageKind) *entry {
	newEntry := &entry{}
	newEntry.init(eType, eCat)
	return newEntry
}

// initLexicalEntry
func initLexicalEntry(definition []byte) (*entry, error) {
	newEntry := initEntry("generic", storageLexeme)
	err := newEntry.Lexeme.define(definition)
	if err != nil {
		return newEntry, err
	}
	return newEntry, nil
}

// record structures an entry for storage in the Memory of an executing L Program
type record struct {
	runtimeFunctionAlias bool
	runtimeType          *runtimeTypeRef
	storageScope         *lexicalScope
	members              map[string]*record
	ID                   string
	Type                 string
	Value                *entry
}

// compileMeta builds the final ID and if a lexeme, the POS if missing from type
func (r *record) compileMeta() error {
	if r == nil || r.Value == nil {
		return lerrors.Internal.Errorf("record metadata requires a value")
	}
	if r.Type == "lexeme.generic" {
		value, err := r.Value.value()
		if err != nil {
			return err
		}
		if value.Type != valueLexeme {
			return lerrors.Internal.Errorf("generic lexical record has %s value", value.Type)
		}
		lex := value.Data.(*lexeme) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
		if lex.POS != "" && lex.POS != "generic" {
			r.Type = "lexeme." + strings.ToLower(lex.POS) + ".generic"
		}
	}
	if r.Value.ID != "" && r.Value.ID != r.ID {
		r.ID += "." + r.Value.ID
	}
	return nil
}

// getTypeMap returns a map[string]string for all applicative, iterative, lexeme, and primitive types
func getTypeMap() map[string]string {
	m := make(map[string]string)
	// Primitives
	m["Null"] = "type.primitive.null"
	m["Thing"] = "type.primitive.thing"
	m["Text"] = "type.primitive.text"
	m["Number"] = "type.primitive.text"
	m["Byte"] = "type.primitive.text"
	m["Truth"] = "type.primitive.text"
	m["Pointer"] = "type.primitive.text"
	// Applicative
	m["Op"] = "type.applicative.operation"
	m["Prop"] = "type.applicative.property"
	m["Func"] = "type.applicative.function"
	m["Definition"] = "type.applicative.definition"
	m["Attribute"] = "type.applicative.attribute"
	m["Object"] = "type.applicative.object"
	m["Class"] = "type.applicative.class"
	m["Service"] = "type.applicative.service"
	m["App"] = "type.applicative.app"
	// Iteratives
	m["Dictionary"] = "type.iterative.dictionary"
	m["List"] = "type.iterative.list"
	return m
}
