package Logiceme

import (
	"fmt"
	"strconv"
	"strings"
)

func projectChildRole(g *Glyph, roles map[int]string, i int, slot, role string) {
	g.appendChildSlot(slot, i)
	roles[i] = role
}
func projectTypedChildRoles(g *Glyph, roles map[int]string, typ, slot, role string, min, max int) error {
	ids := g.childIndices(typ)
	if len(ids) < min || len(ids) > max {
		return invalidGlyph(g, fmt.Sprintf("%s requires %d..%d %s children", slot, min, max, typ))
	}
	for _, i := range ids {
		projectChildRole(g, roles, i, slot, role)
	}
	return nil
}
func requireSingleProp(g *Glyph) (Prop, error) {
	m := g.plan.input
	raw := m["Props"]
	if raw == "" {
		raw = m["Prop"]
	}
	p, e := readNativeProps(raw)
	if e != nil || len(p) != 1 {
		return Prop{}, invalidGlyph(g, "requires exactly one Prop")
	}
	return p[0], nil
}

// Receiver.Type historically records its implementation owner. Preserve that
// shorthand, but keep an explicitly typed self receiver in the signature.
func formatReceiver(receiver Prop, owner string) (string, error) {
	if receiver.Key == "self" {
		if receiver.Type == "" || receiver.Type == "Self" || receiver.Type == owner {
			return "self", nil
		}
		return "self: " + receiver.Type, nil
	}
	if receiver.Key == "&self" || receiver.Key == "&mut self" {
		equivalent := "&Self"
		if receiver.Key == "&mut self" {
			equivalent = "&mut Self"
		}
		if receiver.Type != "" && receiver.Type != "Self" && receiver.Type != owner && receiver.Type != equivalent {
			return "", fmt.Errorf("receiver %s conflicts with explicit Type %s", receiver.Key, receiver.Type)
		}
		return receiver.Key, nil
	}
	if receiver.Type == "" {
		return receiver.Key, nil
	}
	return receiver.Key + ": " + receiver.Type, nil
}
func prepareInstantiation(g *Glyph, ctx glyphContext, roles map[int]string, headers func() error) error {
	m := g.plan.input
	form := glyphForm(g)
	use := func(i int, slot, role string) { projectChildRole(g, roles, i, slot, role) }
	if form == "HY" || form == "MY" || form == "FY" {
		if e := headers(); e != nil {
			return e
		}
	}
	switch form {
	case "DY":
		if m["Key"] == "" {
			if e := projectTypedChildRoles(g, roles, "Calculation", "Atom", "native", 1, 1); e != nil {
				return e
			}
			if len(g.Conjoinments) != 1 {
				return invalidGlyph(g, "native atom cannot also have arguments")
			}
		} else {
			if e := projectTypedChildRoles(g, roles, "Calculation", "Prefix", ctx.role, 0, 1); e != nil {
				return e
			}
			for i := range g.Conjoinments {
				c := &g.Conjoinments[i]
				switch c.Vowels[0].Symbol {
				case "Y":
					use(i, "Arguments", "argument")
				case "O":
					if glyphForm(c) != "MO" {
						return invalidGlyph(g, "associated arguments require MO")
					}
					use(i, "Arguments", "argument")
				}
			}
		}
		g.setSourceMeta("PathSeparator", "<")
		if ctx.role == "callee" || ctx.role == "expression" || ctx.role == "future" {
			g.setSourceMeta("PathSeparator", "::<")
		}
	case "KY", "KYC":
		if receiver, supplied := m["Receiver"]; supplied {
			if strings.TrimSpace(receiver) == "" {
				return invalidGlyph(g, "Receiver must be nonempty when supplied")
			}
			g.setSourceMeta("Receiver", receiver)
		} else if old := g.Meta.get("Receiver"); old.metaVal == "Void" && old.metaType != "String" {
			// Equal bytes can still carry inherited native-source provenance.
			// Replace only this glyph's record; SetAdd keeps equal-value types.
			g.Meta.Metas[old.metaIndex] = &MetaObj{Key: "Receiver", Value: "Void", Type: "String"}
		} else {
			g.Meta.SetAdd("Receiver", "Void", "String")
		}
		if form == "KY" {
			if key, supplied := m["Key"]; supplied {
				if strings.TrimSpace(key) == "" {
					return invalidGlyph(g, "Key must be nonempty when supplied")
				}
				if len(g.Conjoinments) != 0 {
					return invalidGlyph(g, "named KY excludes a callable child")
				}
				g.appendNativeAtoms("Callee", key)
			} else {
				if _, supplied := m["Receiver"]; supplied {
					return invalidGlyph(g, "anonymous KY excludes Receiver")
				}
				if len(g.Conjoinments) != 1 || glyphForm(&g.Conjoinments[0]) != "BW" {
					return invalidGlyph(g, "anonymous KY requires exactly one BW Operation")
				}
				use(0, "Callee", "callee")
			}
		} else if m["Key"] != "" {
			g.appendNativeAtoms("Callee", m["Key"])
		} else {
			if len(g.Conjoinments) != 1 {
				return invalidGlyph(g, "invocation requires exactly one anonymous Operation or DY callee")
			}
			c := &g.Conjoinments[0]
			if glyphForm(c) != "DY" && glyphForm(c) != "FW" && glyphForm(c) != "FWC" {
				return invalidGlyph(g, "callee requires DY or anonymous FW/FWC")
			}
			cm, _ := readGlyphInputs(c.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
			if c.Vowels[0].Symbol == "W" {
				callable, err := readNativeFunc(cm)
				if err != nil {
					return invalidGlyph(g, err.Error())
				}
				if m["Receiver"] != "" || callable.Key != "" {
					return invalidGlyph(g, "anonymous FW child cannot have Key or Receiver")
				}
			}
			if m["Receiver"] != "" && (strings.Contains(cm["Key"], "::") || len(c.childIndices("Calculation")) > 0) {
				return invalidGlyph(g, "receiver callee requires an unqualified method segment")
			}
			use(0, "Callee", "callee")
		}
		g.setSourceMeta("Inputs", strings.Join(readNativeList(m["Inputs"]), ", "))
		g.setSourceMeta("ClosureCallee", "false")
		if len(g.Conjoinments) == 1 && g.Conjoinments[0].Vowels[0].Symbol == "W" {
			g.setSourceMeta("ClosureCallee", "true")
		}
	case "FY":
		ids := g.childIndices("Instantiation")
		if len(ids) > 2 {
			return invalidGlyph(g, "implementation requires one inherent or two trait/type targets")
		}
		if len(ids) > 0 {
			if m["Key"] != "" || m["Receiver"] != "" {
				return invalidGlyph(g, "native and composed implementation targets are exclusive")
			}
			for _, i := range ids {
				use(i, "Targets", "type")
			}
		} else {
			if m["Receiver"] == "" {
				return invalidGlyph(g, "implementation requires Receiver target")
			}
			if m["Key"] != "" {
				g.appendNativeAtoms("Targets", m["Key"])
			}
			g.appendNativeAtoms("Targets", m["Receiver"])
		}
		for _, i := range g.childIndices("Operation") {
			f := glyphForm(&g.Conjoinments[i])
			if f != "FW" && f != "FWC" {
				return invalidGlyph(g, "implementation requires FW/FWC methods")
			}
			use(i, "Methods", "method")
		}
	case "HY", "MY":
		if m["Key"] == "" {
			return invalidGlyph(g, "type declaration requires Key")
		}
		if form == "HY" {
			shape := m["Shape"]
			if shape == "" {
				shape = "Record"
			}
			if shape != "Record" && shape != "Sum" {
				return invalidGlyph(g, "HY Shape must be Record or Sum")
			}
			g.setSourceMeta("Shape", shape)
			for _, i := range g.childIndices("Description") {
				if _, ok := roles[i]; ok {
					continue
				}
				wanted := "BU"
				if shape == "Sum" {
					wanted = "FU"
				}
				if glyphForm(&g.Conjoinments[i]) != wanted {
					return invalidGlyph(g, "type field/variant has wrong role; expected "+wanted)
				}
				use(i, "Fields", "field")
			}
			p, e := readNativeProps(m["Props"])
			if e != nil {
				return invalidGlyph(g, e.Error())
			}
			if len(p) > 0 && len(g.plan.slots["Fields"]) > 0 {
				return invalidGlyph(g, "native Props and composed fields are exclusive")
			}
			if shape == "Sum" && len(p) > 0 {
				return invalidGlyph(g, "Sum variants require FU definitions")
			}
			for _, f := range p {
				if f.Value != "" {
					return invalidGlyph(g, "struct field declaration cannot supply a value; definitions own explicit initialization")
				}
				g.appendNativeAtoms("Fields", f.Key+": "+f.Type)
			}
		} else {
			for _, i := range g.childIndices("Operation") {
				f := glyphForm(&g.Conjoinments[i])
				if f != "HW" && f != "FW" && f != "FWC" {
					return invalidGlyph(g, "trait requires HW signatures or FW/FWC default methods")
				}
				use(i, "Methods", "method")
			}
			if m["Funcs"] != "" && len(g.plan.slots["Methods"]) > 0 {
				return invalidGlyph(g, "Funcs and composed methods are exclusive")
			}
			for _, r := range readNativeList(m["Funcs"]) {
				fields, e := readNativeRecord(r, "Func")
				if e != nil {
					return invalidGlyph(g, e.Error())
				}
				if _, supplied := fields["Receiver"]; supplied {
					return invalidGlyph(g, "Receiver is unsupported by the MY Funcs template interface")
				}
				if e := validateCallableMetadata(g, map[string]string{"Func": r}); e != nil {
					return e
				}
				f, e := readNativeFunc(fields)
				if e != nil {
					return invalidGlyph(g, e.Error())
				}
				if strings.TrimSpace(f.Key) == "" {
					return invalidGlyph(g, "trait signatures require named methods")
				}
				var p []string
				if !f.Receiver.IsVoid() {
					receiver, err := formatReceiver(f.Receiver, "Self")
					if err != nil {
						return invalidGlyph(g, err.Error())
					}
					p = append(p, receiver)
				}
				for _, a := range f.Inputs {
					p = append(p, a.Key+": "+a.Type)
				}
				sig := "fn " + f.Key + "(" + strings.Join(p, ", ") + ")"
				if len(f.Outputs) > 0 {
					sig += " -> " + nativeTuple(f.Outputs)
				}
				g.appendNativeAtoms("Methods", sig+";\n")
			}
		}
	case "Y":
		if m["Key"] == "" {
			return invalidGlyph(g, "record construction requires Key")
		}
		p, e := readNativeProps(m["Props"])
		if e != nil {
			return invalidGlyph(g, e.Error())
		}
		for _, f := range p {
			if f.Key == "" || f.Value == "" {
				return invalidGlyph(g, "construction fields require name and value")
			}
			g.appendNativeAtoms("Fields", f.Key+": "+f.Value)
		}
	}
	// The copied native-record interfaces must receive a collection even when
	// omitted. Inspect the merged metadata so inherited records remain intact.
	collection := ""
	switch form {
	case "Y", "HY":
		collection = "Props"
	case "MY":
		collection = "Funcs"
	}
	if collection != "" {
		if g.Meta.Get(collection).Type == "NotFound" {
			g.Meta.SetAdd(collection, "&[]&", "Array")
		}
		raw := g.Meta.Get(collection).Value
		var count int
		var err error
		if collection == "Props" {
			var records []Prop
			records, err = readNativeProps(raw)
			count = len(records)
		} else {
			var records []Func
			records, err = readNativeFuncs(raw)
			count = len(records)
		}
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
		if count == 0 && (form == "HY" || form == "MY") {
			// These wrappers do not forward composed members or headers. An
			// empty native shell must not silently discard that composition.
			for i := range g.Conjoinments {
				return invalidGlyph(g, "empty "+collection+" cannot discard composed child "+glyphForm(&g.Conjoinments[i]))
			}
			if form == "HY" && m["Shape"] == "Sum" {
				return invalidGlyph(g, "empty Props requires absent Shape or Shape: Record; Shape: Sum is unsupported")
			}
		}
	}
	return nil
}
func prepareDescription(g *Glyph, ctx glyphContext, roles map[int]string) error {
	m := g.plan.input
	form := glyphForm(g)
	use := func(i int, slot, role string) { projectChildRole(g, roles, i, slot, role) }
	if form == "MU" {
		for i := range g.Conjoinments {
			if g.Conjoinments[i].Vowels[0].Symbol != "O" {
				return invalidGlyph(g, "MU updates require explicit Definition children")
			}
			use(i, "Updates", "statement")
		}
		return nil
	}
	if form == "FU" || form == "BU" {
		p, e := requireSingleProp(g)
		if e != nil {
			return e
		}
		if p.Key == "" || p.Value != "" || p.StructTags != "" {
			return invalidGlyph(g, "field/variant requires name without value or attributes")
		}
		g.setSourceMeta("Key", p.Key)
		if form == "FU" {
			if p.Type != "" {
				return invalidGlyph(g, "variant payload uses U, not Prop.Type")
			}
			return projectTypedChildRoles(g, roles, "Description", "Payload", "payload", 0, 1)
		}
		if p.Type != "" {
			if len(g.Conjoinments) > 0 {
				return invalidGlyph(g, "field type sources are exclusive: native or composed child")
			}
			g.appendNativeAtoms("TypeValue", p.Type)
			return nil
		}
		return projectTypedChildRoles(g, roles, "Instantiation", "TypeValue", "type", 1, 1)
	}
	if relation := m["Relation"]; relation != "" {
		switch relation {
		case "Requirements":
			if len(g.Conjoinments) == 0 {
				return invalidGlyph(g, "requirements list requires one or more bounds")
			}
			for i := range g.Conjoinments {
				c := &g.Conjoinments[i]
				cm, _ := readGlyphInputs(c.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
				if glyphForm(c) != "U" || cm["Relation"] != "Bounds" {
					return invalidGlyph(g, "requirements children must be U Bounds")
				}
				use(i, "Predicates", "bounds")
			}
		case "Bounds":
			if e := projectTypedChildRoles(g, roles, "Description", "Subject", "subject", 1, 1); e != nil {
				return e
			}
			if e := projectTypedChildRoles(g, roles, "Instantiation", "Bounds", "type", 1, 1000); e != nil {
				return e
			}
			if e := projectTypedChildRoles(g, roles, "Declaration", "Binder", "binder", 0, 1); e != nil {
				return e
			}
			for _, i := range g.plan.slots["Binder"] {
				c := &g.Conjoinments[i]
				cm, _ := readGlyphInputs(c.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
				if glyphForm(c) != "BA" || cm["Parameter"] != "List" {
					return invalidGlyph(g, "higher-ranked binder requires lifetime BA list")
				}
				for _, p := range c.Conjoinments {
					pm, _ := readGlyphInputs(p.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
					if pm["Parameter"] != "Lifetime" || glyphForm(&p) != "BA" {
						return invalidGlyph(g, "higher-ranked binder permits only lifetime parameters")
					}
				}
			}
		default:
			return invalidGlyph(g, "unknown Relation")
		}
		return nil
	}
	shape := m["Shape"]
	if shape == "" {
		if ctx.role == "pattern" || ctx.role == "payload" {
			return invalidGlyph(g, "neutral description requires explicit Shape")
		}
		for i := range g.Conjoinments {
			if g.Conjoinments[i].Vowels[0].Symbol != "W" {
				return invalidGlyph(g, "structural aggregation requires Operations")
			}
			use(i, "Bodies", "body")
		}
		return nil
	}
	var p Prop
	if m["Prop"] != "" || m["Props"] != "" {
		var e error
		p, e = requireSingleProp(g)
		if e != nil {
			return e
		}
	}
	switch shape {
	case "Atom":
		if p.Key != "" || p.Type != "" || p.StructTags != "" {
			return invalidGlyph(g, "atomic pattern uses Prop.Value only")
		}
		if p.Value != "" {
			if len(g.Conjoinments) > 0 {
				return invalidGlyph(g, "native and composed atom are exclusive")
			}
			g.appendNativeAtoms("Atom", p.Value)
		} else {
			if e := projectTypedChildRoles(g, roles, "Instantiation", "Atom", "type", 1, 1); e != nil {
				return e
			}
		}
	case "Field":
		if p.Key == "" || p.Type != "" || p.Value != "" || p.StructTags != "" {
			return invalidGlyph(g, "field pattern requires only Prop.Key; no Type, Value or attributes")
		}
		g.setSourceMeta("Key", p.Key)
		return projectTypedChildRoles(g, roles, "Description", "Pattern", "pattern", 1, 1)
	case "Unit":
		if len(g.Conjoinments) > 0 {
			return invalidGlyph(g, "unit payload has no children")
		}
	case "Positional", "Record":
		if p.Value != "" || p.Key != "" || p.StructTags != "" {
			return invalidGlyph(g, "constructor description uses Prop.Type only")
		}
		if ctx.role == "payload" {
			for i := range g.Conjoinments {
				wanted := "Instantiation"
				if shape == "Record" {
					wanted = "Description"
				}
				if g.Conjoinments[i].Vowels[0].getSegmentName() != wanted || (shape == "Record" && glyphForm(&g.Conjoinments[i]) != "BU") {
					return invalidGlyph(g, "FU variant payload requires the correct child field role")
				}
				use(i, "Fields", "type")
			}
		} else {
			if p.Type != "" {
				g.appendNativeAtoms("Constructor", p.Type)
			}
			ids := g.childIndices("Instantiation")
			if len(ids) > 1 || len(ids) > 0 && p.Type != "" {
				return invalidGlyph(g, "multiple constructors")
			}
			for _, i := range ids {
				use(i, "Constructor", "type")
			}
			if shape == "Record" && p.Type == "" && len(ids) == 0 {
				return invalidGlyph(g, "record pattern requires constructor")
			}
			seen := map[string]bool{}
			rest := false
			for _, i := range g.childIndices("Description") {
				c := &g.Conjoinments[i]
				cm, _ := readGlyphInputs(c.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
				if shape == "Record" {
					raw := cm["Prop"]
					if raw == "" {
						raw = cm["Props"]
					}
					cp, _ := readNativeProps(raw) //nolint:errcheck // Pattern-shape validation below preserves the existing empty-props diagnostic for malformed metadata.
					if cm["Shape"] == "Atom" && len(cp) == 1 && cp[0].Value == ".." {
						if rest {
							return invalidGlyph(g, "duplicate rests in record pattern")
						}
						rest = true
					} else {
						if rest {
							return invalidGlyph(g, "rest pattern must be last")
						}
						if cm["Shape"] != "Field" || len(cp) != 1 {
							return invalidGlyph(g, "record child requires Field description or explicit rest")
						}
						if seen[cp[0].Key] {
							return invalidGlyph(g, "duplicate field pattern")
						}
						seen[cp[0].Key] = true
					}
				}
				use(i, "Fields", "pattern")
			}
		}
	default:
		return invalidGlyph(g, "unknown Shape")
	}
	g.setSourceMeta("SingletonComma", "false")
	if shape == "Positional" && ctx.role != "payload" && len(g.plan.slots["Fields"]) == 1 && len(g.plan.slots["Constructor"])+len(g.plan.atoms["Constructor"]) == 0 {
		g.setSourceMeta("SingletonComma", "true")
	}
	return nil
}

// Use the same presence rule as Evaluate's keyed branch. In particular, a
// supplied native subject named Void differs from the absent String sentinel.
func isKeyedEvaluation(g *Glyph) bool {
	li := LucidInterpreter{Meta: g.Meta, SegmentalSystem: g.SegmentalSystem}
	return li.interpretCondition("if Meta.Key != Void") == "true"
}

func prepareEvaluation(g *Glyph, ctx glyphContext, roles map[int]string, names *generatedNames) error {
	m := g.plan.input
	form := glyphForm(g)
	use := func(i int, slot, role string) { projectChildRole(g, roles, i, slot, role) }
	switch form {
	case "I":
		if m["Key"] != "" || ctx.role == "pattern" {
			return invalidGlyph(g, "I is only a condition adapter")
		}
		return projectTypedChildRoles(g, roles, "Calculation", "Value", "condition", 1, len(g.Conjoinments))
	case "MI":
		count := 0
		if m["Key"] != "" {
			g.appendNativeAtoms("Subject", m["Key"])
			count++
		}
		for _, i := range g.childIndices("Calculation") {
			use(i, "Subject", "expression")
			count++
		}
		if count != 1 {
			return invalidGlyph(g, "match requires exactly one subject")
		}
		for _, i := range g.childIndices("Evaluation") {
			if glyphForm(&g.Conjoinments[i]) != "HI" {
				return invalidGlyph(g, "match requires HI arms")
			}
			use(i, "Arms", "arm")
		}
	case "HI":
		if ctx.parent != "MI" {
			return invalidGlyph(g, "HI requires MI parent")
		}
		if e := projectTypedChildRoles(g, roles, "Description", "Pattern", "pattern", 1, 1); e != nil {
			return e
		}
		if e := projectTypedChildRoles(g, roles, "Calculation", "Guard", "condition", 0, 1); e != nil {
			return e
		}
		return projectTypedChildRoles(g, roles, "Operation", "Body", "body", 0, 1)
	case "GI":
		if _, local := m["Key"]; !local {
			inherited := g.SegmentalSystem.fileMeta.Get("Key")
			if inherited.Type == "SourceString" {
				// The legacy merge ignores a type change when bytes are equal.
				// Restore provenance on this glyph's private record only.
				key := g.Meta.get("Key")
				g.Meta.Metas[key.metaIndex] = &MetaObj{Key: "Key", Value: inherited.Value, Type: inherited.Type}
			}
		}
		conds, bodies := g.childIndices("Calculation"), g.childIndices("Operation")
		role := "condition"
		if isKeyedEvaluation(g) {
			if strings.TrimSpace(g.Meta.Get("Key").Value) == "" {
				return invalidGlyph(g, "keyed selection requires a nonempty Key")
			}
			if len(conds) == 0 && len(bodies) == 0 {
				return invalidGlyph(g, "keyed selection requires at least one Calculation/Operation pair")
			}
			if len(conds) != len(bodies) {
				return invalidGlyph(g, "keyed selection requires one Operation per Calculation")
			}
			role = "pattern"
		} else if len(bodies) < len(conds) || len(bodies) > len(conds)+1 {
			return invalidGlyph(g, "selection requires one Operation per condition and optional fallback")
		}
		for _, i := range conds {
			use(i, "Conditions", role)
		}
		for _, i := range bodies {
			use(i, "Bodies", "body")
		}
		g.setSourceMeta("Selector", names.fresh("value"))
	case "CI":
		if !ctx.callable {
			return invalidGlyph(g, "select requires callable context")
		}
		defaults := 0
		for i := range g.Conjoinments {
			c := &g.Conjoinments[i]
			if glyphForm(c) != "BI" {
				return invalidGlyph(g, "select accepts BI child branches only")
			}
			cm, err := readGlyphInputs(c.InputString)
			if err != nil {
				return invalidGlyph(c, err.Error())
			}
			if cm["Default"] == "True" {
				defaults++
			}
			roles[i] = "select"
		}
		if defaults > 1 {
			return invalidGlyph(g, "select has multiple defaults")
		}
	case "BI":
		if ctx.parent != "CI" {
			return invalidGlyph(g, "BI requires CI parent")
		}
		if v, ok := m["Default"]; ok && v != "True" && v != "False" {
			return invalidGlyph(g, "Default must be True or False")
		}
		bodies := g.childIndices("Operation")
		if len(bodies) > 1 {
			return invalidGlyph(g, "select branch accepts at most one W body")
		}
		for _, i := range bodies {
			if glyphForm(&g.Conjoinments[i]) != "W" {
				return invalidGlyph(g, "select branch body requires W")
			}
			use(i, "Body", "body")
		}
		ids := append(g.childIndices("Calculation"), g.childIndices("Definition")...)
		if len(ids) != 1 {
			return invalidGlyph(g, "select branch requires exactly one OC/DOC Definition")
		}
		i := ids[0]
		c := &g.Conjoinments[i]
		if c.Vowels[0].Symbol == "E" {
			return invalidGlyph(g, "select branch requires an OC/DOC Definition; wrap the Calculation in DOC with an explicit key")
		}
		cf := glyphForm(c)
		if cf != "OC" && cf != "DOC" {
			return invalidGlyph(g, "select branch requires OC/DOC; use an explicit O in the body for outer assignments")
		}
		use(i, "Guard", "selection")
	}
	return nil
}
func prepareIteration(g *Glyph, ctx glyphContext, roles map[int]string) error {
	m := g.plan.input
	form := glyphForm(g)
	if e := projectTypedChildRoles(g, roles, "Operation", "Body", "body", 1, 1); e != nil {
		return e
	}
	if m["Iterable"] != "" {
		if len(g.Conjoinments) != 1 {
			if form == "X" {
				return invalidGlyph(g, "iterator does not accept condition children")
			}
			return invalidGlyph(g, "iterator does not accept initializer, condition or step")
		}
		arity := m["RangeArity"]
		if arity == "" {
			arity = "2"
		}
		maxArity := 2
		if form == "CX" {
			maxArity = 1
		}
		n, e := strconv.Atoi(arity)
		if e != nil || n < 0 || n > maxArity {
			return invalidGlyph(g, "RangeArity requires 0, 1 or 2; CX requires 0 or 1")
		}
		binding := "_"
		if n > 0 {
			if m["Key"] == "" {
				return invalidGlyph(g, "range arity requires Key")
			}
			binding = m["Key"]
		}
		if n == 2 {
			index := m["Index"]
			if index == "" {
				index = "_"
			}
			binding = "(" + index + ", " + binding + ")"
		}
		g.setSourceMeta("Binding", binding)
		if form == "CX" && !ctx.async {
			return invalidGlyph(g, "channel iteration requires async context")
		}
		return nil
	}
	if form == "CX" {
		return invalidGlyph(g, "channel iteration requires Iterable")
	}
	calculations := g.childIndices("Calculation")
	evaluations := g.childIndices("Evaluation")
	definitions := g.childIndices("Definition")
	if len(calculations) > 0 && len(evaluations) > 0 && len(definitions) > 0 {
		return invalidGlyph(g, "three-part iteration is unsupported; move initialization and steps into explicit compositions")
	}
	conditions := append(append(calculations, evaluations...), definitions...)
	if len(conditions) > 1 {
		return invalidGlyph(g, "iteration accepts at most one condition")
	}
	if len(evaluations) == 1 && glyphForm(&g.Conjoinments[evaluations[0]]) != "I" {
		return invalidGlyph(g, "iteration Evaluation condition requires I")
	}
	if len(definitions) == 1 {
		definition := glyphForm(&g.Conjoinments[definitions[0]])
		if definition != "OG" && definition != "DOG" {
			return invalidGlyph(g, "iteration Definition condition requires OG or DOG")
		}
	}
	if len(conditions) == 1 {
		projectChildRole(g, roles, conditions[0], "Condition", "condition")
	}
	return nil
}
