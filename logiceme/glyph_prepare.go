package Logiceme

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// A role projection references original ordered children. Native slices contain
// only supplied atoms or prepared signatures; executable behavior stays composed.
type glyphPlan struct {
	input      map[string]string
	slots      map[string][]int
	atoms      map[string][]string
	role       string
	statement  bool
	visibility string
	lowered    *Glyph
}
type glyphContext struct {
	parent, role    string
	callable, async bool
	receiverOwner   string
}
type generatedNames struct {
	used   string
	counts map[string]int
}

func (n *generatedNames) fresh(base string) string {
	for {
		v := fmt.Sprintf("__l_%s_%d", base, n.counts[base])
		n.counts[base]++
		if !strings.Contains(n.used, v) {
			n.used += " " + v
			return v
		}
	}
}

var supportedGlyphForms = strings.Fields("A EA BA DA KA HA GA MA PA E DE GE CE ME I GI CI BI MI HI O OB OC OG CO COB FO FOB KO KOB BO BOD BOK BOH BOG BOM BOP BOC DO DOB DOC DOG GO GOB MO MOB HOB PO U MU FU BU W FW BW FWC HW DW X CX Y KY KYC MY HY FY DY V MV DV CV")
var unsupportedGlyphForms = map[string]string{"FA": "BA Parameter: List", "DU": "BA or an explicit O default", "FI": "U Relation: Requirements", "DI": "U Relation: Bounds", "CY": "DY in an expression role", "BY": "MO associated equality", "GY": "HY Shape: Sum", "KU": "FU with a U payload", "HU": "FU with a U payload", "PI": "U Shape: Positional", "KI": "U Shape: Record", "GU": "U Shape: Field", "HO": "HOB", "CW": "FWC", "MW": "FW inside FY/MY"}

func validateGlyphForm(g *Glyph) error {
	f := glyphForm(g)
	if to, ok := unsupportedGlyphForms[f]; ok {
		return lerrors.Unsupported.Errorf("Rust form %s is unsupported; use %s", f, to)
	}
	for _, v := range supportedGlyphForms {
		if v == f {
			return nil
		}
	}
	return lerrors.Unsupported.Errorf("unsupported Rust syllable/coda %s", f)
}
func invalidGlyph(g *Glyph, reason string) error {
	if strings.HasPrefix(reason, "expected ") && strings.HasSuffix(reason, " record") {
		return lerrors.Syntax.Errorf("Rust %s: malformed or unclosed %s", glyphForm(g), reason)
	}
	return lerrors.InvalidInput.Errorf("Rust %s: %s", glyphForm(g), reason)
}
func (g *Glyph) setSourceMeta(k, v string) {
	kind := "String"
	if v == "Void" {
		kind = "SourceString"
	}
	// The dictionary treats an unchanged value as a no-op, including
	// its type. Rust must distinguish the supplied identifier Void from the
	// absent-value sentinel even though their text is identical.
	if old := g.Meta.get(k); old.metaVal == v && old.metaType != kind {
		g.Meta.Metas[old.metaIndex] = &MetaObj{Key: k, Value: v, Type: kind}
		return
	}
	g.Meta.SetAdd(k, v, kind)
}
func (g *Glyph) hasSourceMeta(k string) bool {
	value := g.Meta.Get(k)
	return value.Value != "" && (value.Value != "Void" || value.Type == "SourceString")
}
func (g *Glyph) appendChildSlot(k string, indices ...int) {
	g.plan.slots[k] = append(g.plan.slots[k], indices...)
}
func (g *Glyph) appendNativeAtoms(k string, values ...string) {
	g.plan.atoms[k] = append(g.plan.atoms[k], values...)
}
func (g *Glyph) childIndices(typ string) []int {
	var out []int
	for i := range g.Conjoinments {
		if g.Conjoinments[i].Vowels[0].getSegmentName() == typ {
			out = append(out, i)
		}
	}
	return out
}
func (g *Glyph) childCount(typ string) int { return len(g.childIndices(typ)) }
func isNativeIdentifier(s string) bool {
	return regexp.MustCompile(`^(r#)?[A-Za-z_][A-Za-z0-9_]*$`).MatchString(s) && s != "_"
}
func inputKeys(m map[string]string) []string {
	if v, ok := m["Keys"]; ok {
		return readNativeList(v)
	}
	if v, ok := m["Key"]; ok {
		return []string{v}
	}
	return nil
}
func inputValues(m map[string]string) []string {
	if v, ok := m["Values"]; ok {
		return readNativeList(v)
	}
	if v, ok := m["Value"]; ok {
		return []string{v}
	}
	return nil
}
func nativeTuple(v []string) string {
	if len(v) == 1 {
		return v[0]
	}
	return "(" + strings.Join(v, ", ") + ")"
}
func prepareGlyph(g *Glyph, ctx glyphContext, names *generatedNames) error {
	if g.initializationErr != nil {
		return g.initializationErr
	}
	if err := validateGenericInputInheritance(g.SegmentalSystem); err != nil {
		return err
	}
	if e := validateGlyphForm(g); e != nil {
		return e
	}
	m, e := readGlyphInputs(g.InputString)
	if e != nil {
		return invalidGlyph(g, e.Error())
	}
	form := glyphForm(g)
	g.Type = g.Vowels[0].getSegmentName()
	g.plan = &glyphPlan{input: m, slots: map[string][]int{}, atoms: map[string][]string{}, role: ctx.role}
	if ctx.parent == "" && (form == "FW" || form == "HY" || form == "MY") {
		g.plan.visibility = g.SegmentalSystem.itemVisibility
	}
	if err := validateGlyphMetadata(g, ctx); err != nil {
		return err
	}
	g.Meta = MetaDict{}
	for _, k := range strings.Fields("Key Type Parameter Relation Shape AssignTarget Iterable Index RangeArity Default Lifetimes GenericTypes") {
		g.Meta.SetAdd(k, "Void", "String")
	}
	for _, k := range strings.Fields("Parameters Requirements TypeValue Value Values Bodies Body Tail Fields Methods Targets Arguments Atom Prefix Constructor Subject Bounds Binder Predicates Payload Pattern Guard Arms Conditions Future Selection Step Initializer Condition Updates Statements Terminators Operands") {
		g.plan.slots[k] = nil
	}

	for _, v := range g.SegmentalSystem.fileMeta.GetMetaData() {
		g.Meta.SetAdd(v.Key, v.Value, v.Type)
	}
	for k, v := range m {
		if k == "Imports" && form == "A" && g.SegmentalSystem.fileMeta.Get("Imports").Type == "PreparedStrings" {
			// File collection already includes this A and its sibling/child
			// glyphs. A local list must not hide their file-owned imports.
			continue
		}
		switch k {
		case "Keys", "Values", "Props":
			g.Meta.SetAdd(k, v, "Array")
		case "Func":
			g.Meta.SetAdd(k, v, "Func")
		default:
			g.setSourceMeta(k, v)
		}
	}
	// The copied wrappers consume the existing plural metadata interfaces,
	// including when callers supplied their singular shorthand. Keep the
	// canonical array framing rather than exposing native values as strings.
	for key, values := range map[string][]string{"Keys": inputKeys(m), "Values": inputValues(m)} {
		if values != nil {
			g.Meta.SetAdd(key, "&["+strings.Join(values, ", ")+"]&", "Array")
		}
	}
	g.setSourceMeta("PayloadRole", strconv.FormatBool(ctx.role == "payload"))
	g.setSourceMeta("BodyEnd", "")
	g.setSourceMeta("SelectionGuard", strconv.FormatBool(ctx.role == "selection"))
	if _, ok := m["Deferred"]; ok {
		return lerrors.Unsupported.Errorf("Rust %s: Deferred is unsupported; Rust has no deferred execution operation", form)
	}
	if v, ok := m["Mutable"]; ok {
		if v != "True" && v != "False" {
			return invalidGlyph(g, "Mutable must be True or False")
		}
		if !(form == "BA" || form == "BO" || strings.HasSuffix(form, "OB") || form == "OB") {
			return invalidGlyph(g, "Mutable only applies to newly declared simple bindings")
		}
	}
	if v, ok := m["Capture"]; ok {
		if v != "Infer" && v != "Move" {
			return invalidGlyph(g, "Capture must be Infer or Move")
		}
		if form != "CV" && form != "FW" && form != "FWC" {
			return invalidGlyph(g, "Capture requires an anonymous callable or future")
		}
	}
	g.setSourceMeta("CaptureKeyword", "")
	if m["Capture"] == "Move" {
		g.setSourceMeta("CaptureKeyword", "move ")
	}
	g.setSourceMeta("AsyncKeyword", "")
	if form == "FWC" {
		g.setSourceMeta("AsyncKeyword", "async ")
	}
	roles := map[int]string{}
	childctx := ctx
	childctx.parent = form
	one := func(typ, slot string, required bool) error {
		ids := g.childIndices(typ)
		if len(ids) > 1 || (required && len(ids) != 1) {
			return invalidGlyph(g, slot+" requires exactly one "+typ+" child")
		}
		if len(ids) == 1 {
			g.appendChildSlot(slot, ids[0])
			roles[ids[0]] = slot
		}
		return nil
	}
	use := func(i int, slot, role string) { g.appendChildSlot(slot, i); roles[i] = role }
	all := func(typ, slot, role string) {
		for _, i := range g.childIndices(typ) {
			use(i, slot, role)
		}
	}
	headers := func() error {
		for i := range g.Conjoinments {
			c := &g.Conjoinments[i]
			cm, err := readGlyphInputs(c.InputString)
			if err != nil {
				return err
			}
			if glyphForm(c) == "BA" && cm["Parameter"] == "List" {
				use(i, "Parameters", "parameters")
			} else if glyphForm(c) == "U" && cm["Relation"] == "Requirements" {
				use(i, "Requirements", "requirements")
			}
		}
		if len(g.plan.slots["Parameters"]) > 1 || len(g.plan.slots["Requirements"]) > 1 {
			return invalidGlyph(g, "at most one BA parameter list and one U requirements list")
		}
		return nil
	}
	switch form {
	case "A":
		if len(g.Conjoinments) > 0 {
			return invalidGlyph(g, "file header cannot have children")
		}
	case "EA":
		return invalidGlyph(g, "EA must be a standalone complete file")
	case "DA", "KA", "HA", "GA", "MA", "PA", "ME":
		typ := map[string]string{"DA": "Calculation", "KA": "Operation", "HA": "Description", "GA": "Evaluation", "MA": "Instantiation", "PA": "Iteration", "ME": "Instantiation"}[form]
		if e := one(typ, "Value", true); e != nil {
			return e
		}
		if form == "ME" {
			role := ctx.role
			if role != "type" && role != "argument" && role != "callee" {
				role = "expression"
			}
			roles[g.plan.slots["Value"][0]] = role
		}
	case "BA", "BO":
		if m["Parameter"] == "List" {
			if form != "BA" || len(g.Conjoinments) == 0 {
				return invalidGlyph(g, "parameter list requires one or more parameters")
			}
			seenDefault, seenOther := false, false
			for i := range g.Conjoinments {
				c := &g.Conjoinments[i]
				cm, _ := readGlyphInputs(c.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
				cf := glyphForm(c)
				if cf != "BA" && cf != "OB" && cf != "DOB" && cf != "MOB" {
					return invalidGlyph(g, "parameter list child requires BA declaration or explicit binding definitions")
				}
				kind := cm["Parameter"]
				if kind != "Type" && kind != "Lifetime" && kind != "Const" {
					return invalidGlyph(g, "invalid parameter kind")
				}
				if kind == "Lifetime" && seenOther {
					return invalidGlyph(g, "lifetime parameters precede type and const parameters")
				}
				if kind != "Lifetime" {
					seenOther = true
				}
				if cf != "BA" {
					seenDefault = true
				} else if seenDefault && kind != "Lifetime" {
					return invalidGlyph(g, "required parameter appears after default")
				}
				use(i, "Parameters", "parameter")
			}
		} else {
			keys := inputKeys(m)
			if len(keys) == 0 || m["Parameter"] != "" && len(keys) != 1 {
				return invalidGlyph(g, "declaration requires names; parameters exactly one name")
			}
			for _, k := range keys {
				if !isNativeIdentifier(k) && k != "_" && m["Parameter"] != "Lifetime" {
					return invalidGlyph(g, "declaration requires simple names")
				}
			}
			if len(inputValues(m)) > 0 {
				return invalidGlyph(g, "declaration cannot supply a value or default")
			}
			if m["Type"] == "" && m["Parameter"] != "Type" && m["Parameter"] != "Lifetime" && g.childCount("Instantiation") == 0 {
				return invalidGlyph(g, "declaration or const parameter requires a type")
			}
			g.setSourceMeta("Binding", keys[0])
			for _, k := range keys {
				if m["Mutable"] == "True" && k != "_" {
					k = "mut " + k
				}
				g.appendNativeAtoms("Bindings", k)
			}
			if e := one("Instantiation", "TypeValue", false); e != nil {
				return e
			}
			for _, i := range g.plan.slots["TypeValue"] {
				roles[i] = "type"
			}
			if m["Type"] != "" && len(g.plan.slots["TypeValue"]) > 0 {
				return invalidGlyph(g, "native and composed type are exclusive")
			}
			if ctx.role != "parameter" {
				g.plan.statement = false
			}
		}
	case "OG", "DOG":
		// Definition owns the pattern binding; X supplies repetition and scope.
		// These fragments never acquire ordinary statement punctuation.
		if ctx.parent != "X" || ctx.role != "condition" {
			return invalidGlyph(g, "conditional definition requires direct X condition context")
		}
		keys := inputKeys(m)
		if len(keys) != 1 || strings.TrimSpace(keys[0]) == "" {
			return invalidGlyph(g, "conditional definition requires exactly one nonempty pattern")
		}
		if form == "OG" {
			if len(g.Conjoinments) != 0 {
				return invalidGlyph(g, "OG cannot have children")
			}
			values := inputValues(m)
			if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
				return invalidGlyph(g, "OG requires exactly one explicit nonempty value")
			}
		} else {
			if len(g.Conjoinments) != 1 || g.Conjoinments[0].Vowels[0].getSegmentName() != "Calculation" {
				return invalidGlyph(g, "DOG requires exactly one Calculation child")
			}
			use(0, "Values", "expression")
		}
	case "PO":
		if ctx.parent != "W" || ctx.role != "statement" || !ctx.callable {
			return invalidGlyph(g, "label definition requires a callable W statement context")
		}
		keys := inputKeys(m)
		if len(keys) != 1 || strings.TrimSpace(keys[0]) == "" {
			return invalidGlyph(g, "label definition requires exactly one nonempty label")
		}
		if len(g.Conjoinments) != 1 || g.Conjoinments[0].Vowels[0].getSegmentName() != "Iteration" {
			return invalidGlyph(g, "PO requires exactly one Iteration child")
		}
		// Forward the loop itself: a surrounding block would change the label's
		// continue target. Native spelling remains the caller's responsibility.
		use(0, "Iteration", "iteration")
	case "OC", "DOC":
		// C qualifies the supplied value with an async block. This definition
		// is a branch-header fragment, not an ordinary assignment statement.
		if ctx.parent != "BI" || ctx.role != "selection" {
			return invalidGlyph(g, "asynchronous definition requires direct BI selection context")
		}
		keys := inputKeys(m)
		if len(keys) != 1 {
			return invalidGlyph(g, "asynchronous definition requires exactly one binding name")
		}
		if !isNativeIdentifier(keys[0]) && keys[0] != "_" {
			return invalidGlyph(g, "asynchronous definition requires a simple binding name or _")
		}
		if form == "OC" {
			if len(g.Conjoinments) != 0 {
				return invalidGlyph(g, "OC cannot have children")
			}
			values := inputValues(m)
			if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
				return invalidGlyph(g, "OC requires exactly one explicit nonempty value")
			}
		} else {
			if len(g.Conjoinments) != 1 || g.Conjoinments[0].Vowels[0].getSegmentName() != "Calculation" {
				return invalidGlyph(g, "DOC requires exactly one Calculation child")
			}
			use(0, "Values", "expression")
			childctx.async = true
		}
	case "CO", "COB":
		if !ctx.callable {
			return invalidGlyph(g, "execution binding requires callable context")
		}
		bindings := inputKeys(m)
		if len(bindings) != 1 || (!isNativeIdentifier(bindings[0]) && bindings[0] != "_") {
			return invalidGlyph(g, "execution binding requires exactly one simple name")
		}
		if len(g.Conjoinments) != 1 {
			return invalidGlyph(g, "execution binding requires exactly one CV or DV value")
		}
		switch glyphForm(&g.Conjoinments[0]) {
		case "CV", "DV":
			use(0, "Values", "expression")
		default:
			return invalidGlyph(g, "execution binding requires exactly one CV or DV value")
		}
		g.plan.statement = true
	case "O", "OB", "FO", "FOB", "KO", "KOB", "DO", "DOB", "GO", "GOB", "MO", "MOB":
		keys := inputKeys(m)
		if len(keys) == 0 {
			return invalidGlyph(g, "definition requires Keys")
		}
		binding := strings.HasSuffix(form, "B")
		if _, ok := m["Mutable"]; ok {
			for _, k := range keys {
				if !isNativeIdentifier(k) && k != "_" {
					return invalidGlyph(g, "Mutable requires simple names, not patterns")
				}
			}
		}
		bindings := append([]string(nil), keys...)
		if m["Mutable"] == "True" {
			for i := range bindings {
				if bindings[i] != "_" {
					bindings[i] = "mut " + bindings[i]
				}
			}
		}
		g.setSourceMeta("Binding", nativeTuple(bindings))
		g.setSourceMeta("DeclareKeyword", "")
		if binding {
			g.setSourceMeta("DeclareKeyword", "let ")
		}
		if form == "O" || form == "OB" {
			vals := inputValues(m)
			if len(vals) == 0 {
				return invalidGlyph(g, "definition requires an explicit value")
			}
			g.appendNativeAtoms("Values", nativeTuple(vals))
		} else {
			typ := map[byte]string{'F': "Declaration", 'K': "Operation", 'D': "Calculation", 'G': "Evaluation", 'M': "Instantiation"}[form[0]]
			if e := one(typ, "Values", true); e != nil {
				return e
			}
			if typ == "Instantiation" {
				role := "expression"
				if ctx.role == "argument" || ctx.role == "parameter" {
					role = "type"
				}
				roles[g.plan.slots["Values"][0]] = role
			}
		}
		if ctx.role == "argument" {
			if form != "MO" || len(keys) != 1 {
				return invalidGlyph(g, "associated equality requires MO with one name and type")
			}
			g.setSourceMeta("DeclareKeyword", "")
		}
		if m["Parameter"] != "" {
			if !binding || ctx.role != "parameter" {
				return invalidGlyph(g, "parameter default must be an explicit binding definition")
			}
			if m["Parameter"] == "Lifetime" {
				return invalidGlyph(g, "lifetime parameters cannot have defaults")
			}
			g.setSourceMeta("DeclareKeyword", "")
			if m["Parameter"] == "Const" {
				if form == "OB" || form == "DOB" {
					if err := one("Instantiation", "TypeValue", false); err != nil {
						return err
					}
					if len(g.plan.slots["TypeValue"]) > 0 {
						roles[g.plan.slots["TypeValue"][0]] = "type"
					}
				}
				if m["Type"] != "" && len(g.plan.slots["TypeValue"]) > 0 {
					return invalidGlyph(g, "const type sources are exclusive: native Type or composed DY")
				}
				if m["Type"] == "" && len(g.plan.slots["TypeValue"]) == 0 {
					return invalidGlyph(g, "const parameter requires a native or composed type")
				}
			}
		}
		g.plan.statement = true
	case "BOD", "BOK", "BOH", "BOG", "BOM", "BOP", "BOC":
		typ := map[byte]string{'D': "Calculation", 'K': "Operation", 'H': "Description", 'G': "Evaluation", 'M': "Instantiation", 'P': "Iteration", 'C': "Execution"}[form[2]]
		if e := one(typ, "Value", true); e != nil {
			return e
		}
		g.plan.statement = true
	case "HOB":
		if e := one("Description", "Pattern", true); e != nil {
			return e
		}
		if e := one("Calculation", "Value", true); e != nil {
			return e
		}
		roles[g.plan.slots["Pattern"][0]] = "pattern"
		g.plan.statement = true
	case "E":
		if _, ok := m["Statement"]; !ok {
			return invalidGlyph(g, "Calculation requires Statement")
		}
	case "GE":
		if len(g.Conjoinments) == 0 {
			return invalidGlyph(g, "grouping requires Calculation children")
		}
		all("Calculation", "Value", "expression")
	case "DE":
		raw := m["Ops"]
		if raw == "" {
			raw = m["Op"]
		}
		ops, err := readNativeOps(raw)
		if err != nil || len(ops) == 0 {
			return invalidGlyph(g, "calculation requires Ops")
		}
		op := ops[0]
		ids := g.childIndices("Calculation")
		if len(ids) > 0 && len(ops) != 1 {
			return invalidGlyph(g, "composed calculation requires one Op")
		}
		g.setSourceMeta("NativeOps", "false")
		if len(ids) == 0 {
			data, _ := json.Marshal(ops) //nolint:errcheck // Op contains only strings and string slices, which JSON can always encode.
			g.Meta.SetAdd("Ops", string(data), "PreparedOps")
			g.setSourceMeta("NativeOps", "true")
		}
		if len(ids) > 0 && len(op.Operands) > 0 {
			return invalidGlyph(g, "native and composed operands are exclusive")
		}
		if len(ids) == 0 {
			g.appendNativeAtoms("Operands", op.Operands...)
		} else {
			all("Calculation", "Operands", "expression")
		}
		for _, o := range ops {
			if len(ids) == 0 && len(o.Operands) == 2 && strings.TrimSpace(o.Operator) == "" {
				return invalidGlyph(g, "binary calculation requires an explicit operator")
			}
			if len(ids) == 0 && len(o.Operands) != 2 && !(len(o.Operands) == 1 && o.Type == "Calculate" && o.Operator == "") {
				return invalidGlyph(g, "native operation requires two operands or a grouped operand")
			}
		}
		count := len(ids) + len(op.Operands)
		if count == 2 && strings.TrimSpace(op.Operator) == "" {
			return invalidGlyph(g, "binary calculation requires an explicit operator")
		}
		if count != 2 && !(count == 1 && op.Type == "Calculate" && op.Operator == "") {
			return invalidGlyph(g, "calculation requires two operands or one grouped operand")
		}
		g.setSourceMeta("Operator", op.Operator)
		g.setSourceMeta("Prefix", op.Prefix)
		g.setSourceMeta("Postfix", op.Postfix)
		g.setSourceMeta("Group", strconv.FormatBool(op.Type == "Calculate"))
	case "CE":
		raw := m["Ops"]
		if raw == "" {
			raw = m["Op"]
		}
		ops, err := readNativeOps(raw)
		if err != nil || len(ops) != 1 {
			return invalidGlyph(g, "communication requires exactly one Op")
		}
		op := ops[0]
		if op.Operator != "<-" || op.Prefix != "" || op.Postfix != "" || len(op.Operands) < 1 || len(op.Operands) > 2 {
			return invalidGlyph(g, "communication requires <- and one receiver or sender/value operands; no prefix or postfix")
		}
		for _, v := range op.Operands {
			if strings.TrimSpace(v) == "" {
				return invalidGlyph(g, "empty communication operand")
			}
		}
		data, _ := json.Marshal(ops) //nolint:errcheck // Op contains only strings and string slices, which JSON can always encode.
		g.Meta.SetAdd("Ops", string(data), "PreparedOps")
		g.setSourceMeta("Endpoint", op.Operands[0])
		g.setSourceMeta("Send", "false")
		if len(op.Operands) == 2 {
			g.setSourceMeta("Send", "true")
			g.setSourceMeta("Value", op.Operands[1])
		}
	case "W":
		all("Definition", "Statements", "statement")
	case "FW", "BW", "FWC", "HW":
		// The copied FW/BW cases consume Func and Operations only. Keep
		// unsupported headers unconsumed rather than silently dropping them.
		if form != "FW" && form != "BW" {
			if e := headers(); e != nil {
				return e
			}
		}
		if form == "BW" && (!ctx.callable ||
			!((ctx.parent == "KY" && ctx.role == "callee") ||
				((ctx.parent == "KO" || ctx.parent == "KOB") && ctx.role == "Values"))) {
			return invalidGlyph(g, "BW requires KY callee or KO/KOB value within callable context")
		}
		f, err := readNativeFunc(m)
		if err != nil {
			return invalidGlyph(g, err.Error())
		}
		if form == "FW" && strings.TrimSpace(f.Key) == "" {
			return invalidGlyph(g, "FW requires a named Func; use BW for a closure")
		}
		if form == "BW" && (f.Key != "" || !f.Receiver.IsVoid()) {
			return invalidGlyph(g, "BW excludes a function name and receiver")
		}
		// Func remains a Func record; the primitive owns the signature syntax.
		// Reuse the existing wrapper framing and retain supplied native fields.
		if raw, supplied := m["Func"]; supplied {
			g.Meta.SetAdd("Func", raw, "Func")
		} else {
			var inputs []SegmentInput
			for _, key := range []string{"Key", "Inputs", "Outputs", "Receiver"} {
				if value, supplied := m[key]; supplied {
					inputs = append(inputs, SegmentInput{Key: key, Value: value})
				}
			}
			record := LoadFuncInput("Func", inputs)
			g.Meta.SetAdd("Func", record.Value, "Func")
		}
		g.setSourceMeta("Key", f.Key)
		if _, ok := m["Capture"]; ok && f.Key != "" {
			return invalidGlyph(g, "Capture is invalid on named functions; only anonymous callables")
		}
		if !f.Receiver.IsVoid() && ctx.parent != "FY" && ctx.parent != "MY" {
			return invalidGlyph(g, "method receiver requires FY or MY association")
		}
		var params []string
		if !f.Receiver.IsVoid() {
			receiver, err := formatReceiver(f.Receiver, ctx.receiverOwner)
			if err != nil {
				return invalidGlyph(g, err.Error())
			}
			params = append(params, receiver)
		}
		for _, p := range f.Inputs {
			if p.Key == "" || p.Type == "" {
				return invalidGlyph(g, "function parameters require name and type")
			}
			params = append(params, p.Key+": "+p.Type)
		}
		g.setSourceMeta("Inputs", strings.Join(params, ", "))
		g.setSourceMeta("Result", "")
		if len(f.Outputs) > 0 {
			g.setSourceMeta("Result", " -> "+nativeTuple(f.Outputs))
		}
		if form == "FW" && ctx.parent == "" && f.Key == "main" && g.SegmentalSystem.initializeConfig {
			// The support operation is a prepared sibling before the original
			// body. It neither parses nor changes the user's statements.
			statement := "if let Err(error) = config::initialize_configuration_settings() { ::std::eprintln!(\"configuration startup: {error}\"); ::std::process::exit(1); }"
			prologue := newLoweredGlyph(g.SegmentalSystem, "W", "", newLoweredGlyph(g.SegmentalSystem, "BOD", "", newLoweredGlyph(g.SegmentalSystem, "E", "Statement: ~"+statement+"~")))
			g.Conjoinments = append([]Glyph{prologue}, g.Conjoinments...)
		}
		all("Operation", "Bodies", "body")
		for _, i := range g.plan.slots["Bodies"] {
			end := ""
			if glyphForm(&g.Conjoinments[i]) == "DW" {
				end = "\n"
			}
			g.appendNativeAtoms("BodyEnds", end)
		}
		if form == "HW" && len(g.plan.slots["Bodies"]) > 0 {
			return invalidGlyph(g, "signature cannot have a body")
		}
		if f.Key == "" && len(g.plan.slots["Parameters"])+len(g.plan.slots["Requirements"]) > 0 {
			return invalidGlyph(g, "anonymous callable cannot have generic headers")
		}
		childctx.callable = true
		childctx.async = form == "FWC"
	case "DW":
		all("Operation", "Bodies", "body")
		if e := one("Definition", "Tail", true); e != nil {
			return e
		}
		roles[g.plan.slots["Tail"][0]] = "tail"
	case "CV":
		if !ctx.callable {
			return invalidGlyph(g, "execution is invalid at file scope; requires callable context")
		}
		if len(g.Conjoinments) != 1 || glyphForm(&g.Conjoinments[0]) != "W" {
			return invalidGlyph(g, "future construction requires exactly one W body")
		}
		use(0, "Bodies", "body")
		childctx.async, childctx.callable = true, true
	case "DV":
		if !ctx.callable {
			return invalidGlyph(g, "execution is invalid at file scope; requires callable context")
		}
		if !ctx.async {
			return invalidGlyph(g, "await requires async context")
		}
		// Local operand presence owns this projection; file metadata cannot
		// add an extra operand beside a composed child.
		g.setSourceMeta("FutureKey", m["Key"])
		if key, supplied := m["Key"]; supplied {
			if strings.TrimSpace(key) == "" {
				return invalidGlyph(g, "future Key must be nonempty when supplied")
			}
			if len(g.Conjoinments) != 0 {
				return invalidGlyph(g, "native and composed future operands are exclusive")
			}
			g.appendNativeAtoms("Value", key)
		} else {
			if len(g.Conjoinments) != 1 {
				return invalidGlyph(g, "await requires exactly one future operand")
			}
			switch glyphForm(&g.Conjoinments[0]) {
			case "E", "KY", "CV", "DV":
				use(0, "Value", "future")
			default:
				return invalidGlyph(g, "await requires a future expression, not a scheduling statement or item")
			}
		}
	case "V", "MV":
		if !ctx.callable {
			return invalidGlyph(g, "execution is invalid at file scope; requires callable context")
		}
		sources := 0
		if m["Key"] != "" {
			g.appendNativeAtoms("Value", m["Key"])
			sources++
		}
		for i := range g.Conjoinments {
			c := &g.Conjoinments[i]
			if child := glyphForm(c); child == "CV" || child == "DV" {
				return invalidGlyph(g, "scheduling requires its native Key or invocation; direct CV/DV values are unsupported")
			}
			if c.Vowels[0].Symbol == "E" || c.Vowels[0].Symbol == "Y" || c.Vowels[0].Symbol == "V" {
				use(i, "Value", "future")
				sources++
			}
		}
		if sources != 1 {
			return invalidGlyph(g, "execution requires exactly one native or composed future")
		}
	case "Y", "HY", "MY", "FY", "DY", "KY", "KYC":
		if e := prepareInstantiation(g, ctx, roles, headers); e != nil {
			return e
		}
		if form == "FY" {
			childctx.receiverOwner = m["Receiver"]
			if ids := g.plan.slots["Targets"]; len(ids) > 0 {
				target := &g.Conjoinments[ids[len(ids)-1]]
				targetInputs, _ := readGlyphInputs(target.InputString) //nolint:errcheck // Role preflight keeps its empty-metadata fallback; child preparation owns metadata diagnostics.
				childctx.receiverOwner = targetInputs["Key"]
			}
		} else if form == "MY" {
			childctx.receiverOwner = "Self"
		}
	case "U", "MU", "FU", "BU":
		if e := prepareDescription(g, ctx, roles); e != nil {
			return e
		}
	case "I", "GI", "MI", "HI", "CI", "BI":
		if e := prepareEvaluation(g, ctx, roles, names); e != nil {
			return e
		}
	case "X", "CX":
		if e := prepareIteration(g, ctx, roles); e != nil {
			return e
		}
	}
	for i := range g.Conjoinments {
		role, ok := roles[i]
		if !ok {
			return invalidGlyph(g, fmt.Sprintf("unexpected child %s: wrong category or role", glyphForm(&g.Conjoinments[i])))
		}
		cc := childctx
		cc.role = role
		if e := prepareGlyph(&g.Conjoinments[i], cc, names); e != nil {
			return &glyphChildError{form: form, index: i + 1, cause: e}
		}
	}
	if ctx.role == "tail" {
		switch form {
		case "BOD", "BOK", "BOM", "BOC", "BOG", "BOP":
		default:
			return invalidGlyph(g, "DW value tail requires expression-valued Definition adapter")
		}
	}
	if form == "HI" || form == "BI" {
		for _, i := range g.plan.slots["Body"] {
			if glyphForm(&g.Conjoinments[i]) == "DW" {
				g.setSourceMeta("BodyEnd", "\n")
			}
		}
	}
	if form == "U" && m["Shape"] == "Record" {
		for _, i := range g.plan.slots["Fields"] {
			c := &g.Conjoinments[i]
			end := ",\n"
			if c.plan.input["Shape"] == "Atom" {
				end = ""
			}
			g.appendNativeAtoms("FieldEnds", end)
		}
	}
	if form == "W" {
		for _, i := range g.plan.slots["Statements"] {
			c := &g.Conjoinments[i]
			if c.plan.statement {
				g.appendNativeAtoms("Terminators", ";\n")
			} else {
				g.appendNativeAtoms("Terminators", "")
			}
		}
	}
	for _, key := range []string{"Key", "Type", "Iterable"} {
		g.setSourceMeta("Has"+key, strconv.FormatBool(g.hasSourceMeta(key)))
	}
	if err := validatePreparedGlyph(g, ctx); err != nil {
		return err
	}
	return lowerPreparedGlyph(g, ctx, names)
}

// Checked expansion builds the same glyph graph developers compose. No emitted
// source is searched or rewritten; ordinary Define/Evaluate/Iterate own syntax.
func newLoweredGlyph(ss *SegmentalSystem, form, input string, children ...Glyph) Glyph {
	g := ss.InitializeGlyph([]byte(form), nil, nil)
	g.InputString = input
	g.Conjoinments = children
	return g
}
func lowerPreparedGlyph(g *Glyph, ctx glyphContext, names *generatedNames) error {
	ss := g.SegmentalSystem
	node := func(f, in string, c ...Glyph) Glyph { return newLoweredGlyph(ss, f, in, c...) }
	if glyphForm(g) == "CI" {
		if len(g.Conjoinments) == 0 {
			// The block remains a statement-compatible expression when W places
			// another definition after this indefinitely pending selection.
			expanded := node("I", "", node("E", "Statement: ~{ ::core::future::pending::<()>().await }~"))
			expanded.origin = g.origin
			expanded.Conjoinments[0].origin = g.origin
			if err := prepareGlyph(&expanded, ctx, names); err != nil {
				return err
			}
			g.plan.lowered = &expanded
		} else {
			// Rendering consumes the actual category children. Keep ordinary
			// branches in source order and put the sole explicit default last.
			children := make([]Glyph, 0, len(g.Conjoinments))
			for _, child := range g.Conjoinments {
				if child.plan.input["Default"] != "True" {
					children = append(children, child)
				}
			}
			for _, child := range g.Conjoinments {
				if child.plan.input["Default"] == "True" {
					children = append(children, child)
				}
			}
			g.Conjoinments = children
		}
	}
	return nil
}
