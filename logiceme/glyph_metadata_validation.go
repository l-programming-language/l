package Logiceme

import (
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// These are source inputs, not prepared implementation fields. Conjoinment
// roles, signature punctuation, safety/module headers and helper names are
// compiler-owned and must never be supplied by glyph metadata. Ordinary Class
// metadata is inherited separately; generic list interiors are glyph-local and
// checked at both the file boundary and direct preparation entry point.
var nativeSourceInputs = map[string]string{
	"A": "", "EA": "Contents Type",
	"BA": "Key Keys Type Mutable Parameter", "BO": "Key Keys Type Mutable",
	"DA": "", "KA": "", "HA": "", "GA": "", "MA": "", "PA": "",
	"E": "Statement", "DE": "Ops Op", "GE": "Key", "CE": "Ops Op", "ME": "",
	"I": "", "GI": "Key", "CI": "", "BI": "Default", "MI": "Key", "HI": "",
	"O": "Key Keys Value Values Type", "OB": "Key Keys Value Values Type Mutable Parameter",
	"OC": "Key Keys Value Values", "DOC": "Key Keys",
	"CO": "Key Keys", "COB": "Key Keys",
	"OG": "Key Keys Value Values", "DOG": "Key Keys", "PO": "Key Keys",
	"FO": "Key Keys Type", "FOB": "Key Keys Type Mutable",
	"KO": "Key Keys Type", "KOB": "Key Keys Type Mutable",
	"DO": "Key Keys Type", "DOB": "Key Keys Type Mutable Parameter",
	"GO": "Key Keys Type", "GOB": "Key Keys Type Mutable",
	"MO": "Key Keys Type", "MOB": "Key Keys Type Mutable Parameter",
	"BOD": "", "BOK": "", "BOH": "", "BOG": "", "BOM": "", "BOP": "", "BOC": "", "HOB": "",
	"U": "Prop Props Relation Shape", "MU": "", "FU": "Prop Props", "BU": "Prop Props",
	"W": "", "FW": "Func Key Inputs Outputs Receiver Lifetimes GenericTypes", "BW": "Func Inputs Outputs",
	"FWC": "Func Key Inputs Outputs Receiver Capture",
	"HW":  "Func Key Inputs Outputs Receiver", "DW": "",
	"X": "Key Iterable Index RangeArity", "CX": "Key Iterable RangeArity",
	"Y": "Key Props Lifetimes GenericTypes", "KY": "Key Receiver Inputs Lifetimes GenericTypes", "KYC": "Key Receiver Inputs",
	"MY": "Key Funcs Lifetimes GenericTypes", "HY": "Key Props Shape Lifetimes GenericTypes", "FY": "Key Receiver", "DY": "Key",
	"V": "Key", "MV": "Key", "DV": "Key", "CV": "Capture",
}

// These native list interiors belong only to their declaring/applying glyph.
// Reject inherited entries even when empty, sentinel-valued or locally overridden.
func validateGenericInputInheritance(system *SegmentalSystem) error {
	for _, name := range []string{"Lifetimes", "GenericTypes"} {
		if system.fileMeta.Get(name).Type != "NotFound" {
			return lerrors.InvalidInput.Errorf("Rust %s is glyph-local; inherited file metadata is unsupported", name)
		}
	}
	return nil
}

func validateGlyphMetadata(g *Glyph, ctx glyphContext) error {
	m := map[string]string(nil)
	if g.plan != nil {
		m = g.plan.input
	} else {
		var err error
		m, err = readGlyphInputs(g.InputString)
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
	}
	form := glyphForm(g)
	if _, exists := m["Deferred"]; exists {
		return lerrors.Unsupported.Errorf("Rust %s: Deferred is unsupported; Rust has no deferred execution operation", form)
	}
	inputs, supported := nativeSourceInputs[form]
	if !supported {
		return validateGlyphForm(g)
	}
	allowed := map[string]bool{"Imports": true, "Tags": true}
	for _, name := range strings.Fields(inputs) {
		allowed[name] = true
	}
	keys := make([]string, 0, len(m))
	for name := range m {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		if !allowed[name] {
			if name == "Mutable" {
				return invalidGlyph(g, "Mutable only applies to newly declared simple bindings; it cannot qualify assignments or a complete pattern")
			}
			if name == "Capture" {
				if form == "FW" || form == "BW" {
					return invalidGlyph(g, "Capture is unsupported by the FW/BW template interface")
				}
				return invalidGlyph(g, "Capture requires an anonymous callable or future")
			}
			if strings.HasPrefix(name, "Header") {
				return invalidGlyph(g, name+" is a compiler-owned header input")
			}
			return invalidGlyph(g, "unsupported metadata "+name+" for this form")
		}
	}
	for _, name := range []string{"Lifetimes", "GenericTypes"} {
		if value, supplied := m[name]; supplied {
			if strings.TrimSpace(value) == "" {
				return invalidGlyph(g, name+" must be nonempty when supplied")
			}
			if _, named := m["Key"]; form == "KY" && !named {
				return invalidGlyph(g, name+" requires a named KY with a local Key")
			}
		}
	}
	for _, pair := range [][2]string{{"Key", "Keys"}, {"Value", "Values"}, {"Op", "Ops"}, {"Prop", "Props"}} {
		_, first := m[pair[0]]
		_, second := m[pair[1]]
		if first && second {
			return invalidGlyph(g, pair[0]+" and "+pair[1]+" metadata are exclusive")
		}
	}
	if form == "FW" || form == "BW" {
		if err := validateCallableMetadata(g, m); err != nil {
			return err
		}
	}
	if parameter, exists := m["Parameter"]; exists {
		if parameter == "List" {
			if form != "BA" {
				return invalidGlyph(g, "Parameter List requires BA")
			}
			for _, name := range []string{"Key", "Keys", "Type", "Mutable"} {
				if _, exists := m[name]; exists {
					return invalidGlyph(g, "Parameter List cannot consume "+name)
				}
			}
			if ctx.role != "parameters" && ctx.role != "binder" {
				return invalidGlyph(g, "Parameter List requires a callable/type header or higher-ranked binder position")
			}
		} else {
			if parameter != "Type" && parameter != "Lifetime" && parameter != "Const" {
				return invalidGlyph(g, "Parameter must be List, Type, Lifetime or Const")
			}
			if ctx.role != "parameter" {
				return invalidGlyph(g, "Parameter requires a BA parameter-list parent")
			}
		}
	}
	if relation, exists := m["Relation"]; exists {
		if relation != "Requirements" && relation != "Bounds" {
			return invalidGlyph(g, "Relation must be Requirements or Bounds")
		}
		for _, name := range []string{"Shape", "Prop", "Props"} {
			if _, exists := m[name]; exists {
				return invalidGlyph(g, "Relation cannot consume "+name)
			}
		}
		if relation == "Requirements" && ctx.role != "requirements" {
			return invalidGlyph(g, "Requirements requires a callable/type header position")
		}
		if relation == "Bounds" && ctx.role != "bounds" {
			return invalidGlyph(g, "Bounds requires a Requirements parent")
		}
	}
	if form == "U" {
		if shape, exists := m["Shape"]; exists {
			switch shape {
			case "Atom", "Field", "Unit", "Positional", "Record":
			default:
				return invalidGlyph(g, "Shape must be Atom, Field, Unit, Positional or Record")
			}
		} else if _, exists := m["Relation"]; !exists {
			for _, name := range []string{"Prop", "Props"} {
				if _, exists := m[name]; exists {
					return invalidGlyph(g, "structural aggregation cannot consume "+name+" without a Shape")
				}
			}
		}
	}
	if form == "X" || form == "CX" {
		if m["Iterable"] == "" {
			for _, name := range []string{"Key", "Index", "RangeArity"} {
				if _, exists := m[name]; exists {
					return invalidGlyph(g, name+" requires an Iterable")
				}
			}
		} else if m["RangeArity"] != "" && m["RangeArity"] != "2" {
			if _, exists := m["Index"]; exists {
				return invalidGlyph(g, "Index requires RangeArity 2")
			}
			if m["RangeArity"] == "0" {
				if _, exists := m["Key"]; exists {
					return invalidGlyph(g, "Key is unused with RangeArity 0")
				}
			}
		}
	}
	return nil
}

// Validate the record interfaces actually consumed by the copied FW/BW cases.
// Keep the shared decoders and other Prop destinations unchanged: a recognized
// outer Func must not conceal fields that these templates would silently ignore.
func validateCallableMetadata(g *Glyph, signature map[string]string) error {
	checkFields := func(fields map[string]string, kind, allowed string) error {
		allowedFields := map[string]bool{}
		for _, key := range strings.Fields(allowed) {
			allowedFields[key] = true
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if !allowedFields[key] {
				return invalidGlyph(g, "unsupported "+kind+" field "+key)
			}
		}
		return nil
	}
	if raw, supplied := signature["Func"]; supplied {
		for _, key := range []string{"Key", "Inputs", "Outputs", "Receiver"} {
			if _, flat := signature[key]; flat {
				return invalidGlyph(g, "Func record and flat signature metadata are exclusive: conflicting "+key)
			}
		}
		record, err := readNativeRecord(raw, "Func")
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
		if err := checkFields(record, "Func", "Key Inputs Outputs Receiver"); err != nil {
			return err
		}
		signature = record
	}
	for _, raw := range readNativeList(signature["Inputs"]) {
		prop, err := readNativeRecord(raw, "Prop")
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
		if err := checkFields(prop, "input Prop", "Key Type"); err != nil {
			return err
		}
		if strings.TrimSpace(prop["Key"]) == "" || strings.TrimSpace(prop["Type"]) == "" {
			return invalidGlyph(g, "function parameters require name and type")
		}
	}
	for _, raw := range readNativeList(signature["Receiver"]) {
		prop, err := readNativeRecord(raw, "Prop")
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
		if err := checkFields(prop, "Receiver Prop", "Key Type Value StructTags"); err != nil {
			return err
		}
	}
	return nil
}
