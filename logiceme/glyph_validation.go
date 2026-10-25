package Logiceme

import (
	"fmt"
	"strings"
)

// glyphChildError carries compiler-owned structural provenance independently
// from application/file context, which may contain arbitrary user names.
type glyphChildError struct {
	form  string
	index int
	cause error
}

func (e *glyphChildError) Error() string {
	return fmt.Sprintf("%s child %d: %v", e.form, e.index, e.cause)
}

func (e *glyphChildError) Unwrap() error { return e.cause }

// validatePreparedGlyph checks logical roles after the ordered child projection
// has been prepared. Native type, borrow, trait and exhaustiveness checking
// belongs to rustc; these rules only reject malformed logeme composition.
func validatePreparedGlyph(g *Glyph, ctx glyphContext) error {
	m := g.plan.input
	form := glyphForm(g)
	if form == "HW" && ctx.parent != "MY" {
		return invalidGlyph(g, "bodyless HW signature requires MY trait association")
	}
	if form == "HOB" {
		if _, exists := m["Mutable"]; exists {
			return invalidGlyph(g, "HOB pattern binding cannot apply Mutable to the complete pattern")
		}
	}
	if form == "FW" || form == "BW" || form == "FWC" || form == "HW" {
		if ctx.role == "method" {
			method, err := readNativeFunc(m)
			if err != nil {
				return invalidGlyph(g, err.Error())
			}
			if method.Key == "" {
				return invalidGlyph(g, "implementation and trait members require named methods")
			}
		}
		if _, exists := m["Func"]; exists {
			for _, key := range []string{"Key", "Inputs", "Outputs", "Receiver"} {
				if _, flat := m[key]; flat {
					return invalidGlyph(g, "Func record and flat signature metadata are exclusive: conflicting "+key)
				}
			}
		}
	}
	if form == "FW" || form == "FWC" || form == "HW" || form == "FY" {
		for _, index := range g.plan.slots["Parameters"] {
			for i := range g.Conjoinments[index].Conjoinments {
				parameter := &g.Conjoinments[index].Conjoinments[i]
				if glyphForm(parameter) != "BA" {
					return invalidGlyph(g, "generic parameter defaults are not allowed at this function or implementation destination")
				}
			}
		}
	}
	if m["Mutable"] == "True" {
		named := 0
		for _, key := range inputKeys(m) {
			if key == "_" {
				continue
			}
			if !isNativeIdentifier(key) {
				return invalidGlyph(g, "Mutable requires simple names; discard _ and patterns cannot be mutable bindings")
			}
			named++
		}
		if named == 0 {
			return invalidGlyph(g, "Mutable requires a named binding; discard _ cannot be mutable")
		}
	}
	if form == "MO" && ctx.role == "argument" {
		if len(g.Conjoinments) != 1 || glyphForm(&g.Conjoinments[0]) != "DY" {
			return invalidGlyph(g, "associated equality argument requires exactly one DY type child; invocation is not a type")
		}
	}
	if (ctx.role == "type" || ctx.role == "argument") && g.Vowels[0].Symbol == "Y" && form != "DY" {
		return invalidGlyph(g, "type/associated argument requires DY, not an invocation or item child "+form)
	}
	if (ctx.parent == "KO" || ctx.parent == "KOB") && (form == "FW" || form == "FWC") {
		f, err := readNativeFunc(m)
		if err == nil && f.Key != "" {
			return invalidGlyph(g, "named function item cannot supply an expression value; requires anonymous callable")
		}
	}
	if form == "MO" && ctx.parent == "" {
		return invalidGlyph(g, "associated equality MO requires a path argument position; executable definition requires callable scope")
	}
	if parameter := m["Parameter"]; parameter != "" && parameter != "List" {
		if ctx.role != "parameter" {
			return invalidGlyph(g, "parameter requires a BA parameter-list parent")
		}
		keys := inputKeys(m)
		if len(keys) != 1 {
			return invalidGlyph(g, "parameter requires exactly one name")
		}
		if _, exists := m["Mutable"]; exists {
			return invalidGlyph(g, "parameter declarations do not accept Mutable")
		}
		switch parameter {
		case "Type":
			if m["Type"] != "" {
				return invalidGlyph(g, "Type parameter cannot supply a const Type annotation")
			}
		case "Lifetime":
			if !strings.HasPrefix(keys[0], "'") || !isNativeIdentifier(strings.TrimPrefix(keys[0], "'")) {
				return invalidGlyph(g, "lifetime parameter requires one lifetime name")
			}
			if m["Type"] != "" {
				return invalidGlyph(g, "lifetime parameter cannot supply Type")
			}
		case "Const":
			if m["Type"] == "" && g.childCount("Instantiation") == 0 {
				return invalidGlyph(g, "const parameter requires Type")
			}
		default:
			return invalidGlyph(g, "parameter kind must be Type, Lifetime or Const")
		}
	}
	if form == "BA" && m["Parameter"] == "List" {
		seen := map[string]bool{}
		for i := range g.Conjoinments {
			cm := g.Conjoinments[i].plan.input
			for _, key := range inputKeys(cm) {
				if seen[key] {
					return invalidGlyph(g, "parameter list has duplicate name "+key)
				}
				seen[key] = true
			}
		}
	}
	if form == "DY" && m["Key"] == "" && len(g.Conjoinments) == 1 {
		atom := g.Conjoinments[0].plan.input["Statement"]
		if isNativeItemAtom(atom) {
			return invalidGlyph(g, "native type/path atom cannot contain a whole item declaration")
		}
	}
	if form == "CE" {
		// A trailing comma is an empty operand in the existing Op contract,
		// rather than optional Rust list punctuation. Inspect the framed source
		// before the general list reader discards that absent final value.
		raw := m["Ops"]
		if raw == "" {
			raw = m["Op"]
		}
		for _, record := range readNativeList(raw) {
			fields, err := readNativeRecord(record, "Op")
			if err != nil {
				return invalidGlyph(g, err.Error())
			}
			operands := strings.TrimSpace(unframeNativeList(fields["Operands"]))
			if strings.HasSuffix(operands, ",") {
				return invalidGlyph(g, "communication requires nonempty operands; empty final operand")
			}
		}
	}
	if form == "U" && ctx.role == "payload" {
		if m["Shape"] != "Unit" && m["Shape"] != "Positional" && m["Shape"] != "Record" {
			return invalidGlyph(g, "FU payload requires Unit, Positional or Record description")
		}
		if raw := m["Prop"] + m["Props"]; raw != "" {
			return invalidGlyph(g, "FU payload does not accept a constructor or Prop metadata")
		}
	}
	return nil
}

// An unambiguous item introducer is a structural role error. fn/impl/const are
// deliberately excluded: they can introduce native type or expression atoms.
// This does not parse native expressions or predict whether a type resolves.
func isNativeItemAtom(atom string) bool {
	words := strings.Fields(atom)
	if len(words) == 0 {
		return false
	}
	for _, word := range words {
		switch word {
		case "pub", "async", "unsafe":
			continue
		case "struct", "enum", "trait", "mod", "use", "type", "static":
			return true
		}
		return false
	}
	return false
}
