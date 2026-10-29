package Logiceme

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func glyphInputSource(g *Glyph) string {
	s := g.InputString
	for i := range g.Conjoinments {
		s += " " + glyphInputSource(&g.Conjoinments[i])
	}
	return s
}
func (g *Glyph) compiledChildGroups() []CompiledConjoinment {
	out := make([]CompiledConjoinment, 0, 18)
	add := func(k string, contents [][]byte, plural bool) {
		out = append(out, CompiledConjoinment{Key: k, Type: k, Contents: contents, Plural: plural, Index: -1})
	}
	for _, typ := range []string{"Declaration", "Calculation", "Evaluation", "Definition", "Description", "Operation", "Iteration", "Instantiation", "Execution"} {
		var values [][]byte
		for i := range g.Conjoinments {
			c := &g.Conjoinments[i]
			if c.Type == typ {
				values = append(values, c.Content)
			}
		}
		add(typ+"s", values, true)
		add(typ, values, false)
	}
	return out
}

// A primitive sees only the categories forwarded by its selected vowel. Build
// those records directly instead of allocating every singular/plural category
// before discarding the unforwarded ones.
func (g *Glyph) compiledChildGroup(name string) (CompiledConjoinment, bool) {
	typ := strings.TrimSuffix(name, "s")
	switch typ {
	case "Declaration", "Calculation", "Evaluation", "Definition", "Description", "Operation", "Iteration", "Instantiation", "Execution":
	default:
		return CompiledConjoinment{}, false
	}
	group := CompiledConjoinment{Key: name, Type: name, Plural: name != typ, Index: -1}
	for i := range g.Conjoinments {
		child := &g.Conjoinments[i]
		if child.Type == typ {
			group.Contents = append(group.Contents, child.Content)
		}
	}
	return group, true
}

// functionalArgumentData projects the wrapper's explicit argument boundary onto
// already-prepared values. Metadata retains lexical/private fallback, as on the
// generic path, while child collections are visible only under forwarded names.
// Argument sources are read from the caller snapshot, so alias order cannot
// mutate a later argument or leak back into the glyph's prepared state.
func (g *Glyph) functionalArgumentData(params []Parameter) (DataCompilation, error) {
	var meta MetaDict
	for _, value := range g.Meta.Metas {
		meta.SetAdd(value.Key, value.Value, value.Type)
	}
	var children []CompiledConjoinment
	seen := map[string]bool{}
	for _, param := range params {
		identity := param.Type + ":" + param.Key
		if seen[identity] {
			return DataCompilation{}, lerrors.InvalidConfiguration.Errorf("duplicate Rust functional %s argument %s", param.Type, param.Key)
		}
		seen[identity] = true
		switch param.Type {
		case "Meta":
			value := *g.Meta.Get(param.SubType)
			value.Key = param.Key
			replaced := false
			for _, existing := range meta.GetMetas() {
				if existing.Key == param.Key {
					// Equal bytes can still change the prepared record type.
					// Replace our private record instead of MetaDict.SetAdd's
					// equal-value no-op; the caller remains untouched.
					*existing = value
					replaced = true
					break
				}
			}
			if !replaced {
				meta.Add(value.Key, value.Value, value.Type)
			}
		case "Conjoinments":
			child, found := g.compiledChildGroup(param.SubType)
			if !found {
				return DataCompilation{}, lerrors.InvalidConfiguration.Errorf("Rust functional argument %s has unknown conjoinment source %s", param.Key, param.SubType)
			}
			// Keep the source category; only Key is exposed to the callee.
			child.Key = param.Key
			children = append(children, child)
			if param.Key != param.SubType {
				kind := "Conjoinment"
				if child.Plural {
					kind = "Array-Conjoinment"
				}
				meta.SetAdd(param.Key, "Conjoinments."+param.SubType, kind)
			}
		default:
			return DataCompilation{}, lerrors.InvalidConfiguration.Errorf("Rust functional argument %s has unsupported source %s", param.Key, param.Value)
		}
	}
	return DataCompilation{Conjoinments: children, Meta: meta, SegmentalSystem: g.SegmentalSystem}, nil
}

// Validate the existing fallback carrier after composition, since a wrapper
// can produce Void too. Comparison uses Lucid's indexed scalar semantics; no
// child source is rewritten or promoted to a new template interface.
func (g *Glyph) validateCompiledKeyedEvaluation() error {
	if glyphForm(g) != "GI" || !isKeyedEvaluation(g) {
		return nil
	}
	li := LucidInterpreter{Meta: g.Meta, Conjoinments: g.compiledChildGroups(), SegmentalSystem: g.SegmentalSystem}
	patterns := g.childIndices("Calculation")
	var fallbacks []int
	invalid := func(entry int, reason string) error {
		return &glyphChildError{form: "GI", index: patterns[entry] + 1, cause: invalidGlyph(g, reason)}
	}
	for entry, child := range patterns {
		if strings.TrimSpace(string(g.Conjoinments[child].Content)) == "" {
			return invalid(entry, fmt.Sprintf("keyed selection requires a nonempty pattern at entry %d", entry+1))
		}
		condition := fmt.Sprintf("if Conjoinments.Calculations[%d] == Void", entry)
		if li.interpretCondition(condition) == "true" {
			fallbacks = append(fallbacks, entry)
		}
	}
	if len(fallbacks) > 1 {
		return invalid(fallbacks[1], "keyed selection has multiple fallback patterns")
	}
	if len(fallbacks) == 1 && fallbacks[0] != len(patterns)-1 {
		return invalid(fallbacks[0], "keyed selection fallback must be terminal")
	}
	return nil
}

func (g *Glyph) compilePreparedGlyph() error {
	if g.plan.lowered != nil {
		if e := g.plan.lowered.compilePreparedGlyph(); e != nil {
			return e
		}
		g.Content = g.plan.lowered.Content
		return nil
	}
	for i := range g.Conjoinments {
		if e := g.Conjoinments[i].compilePreparedGlyph(); e != nil {
			return e
		}
	}
	if e := g.validateCompiledKeyedEvaluation(); e != nil {
		return e
	}
	vowel := g.Vowels[0]
	vowel.Tags = parseSegmentTags(vowel.selectedTemplateTags(g.SegmentalSystem, g.segmentSelection(), true), g.Meta, g.segmentSelection(), true, true)
	var cases []*LucidTag
	for _, candidate := range vowel.Tags {
		switch candidate.Expressions[1].ExpressionString {
		case "Headers", "Inputs":
			// Initialization declarations are never executable source bodies.
		default:
			cases = append(cases, candidate)
		}
	}
	if len(cases) != 1 {
		return lerrors.InvalidConfiguration.Errorf("Rust vowel %s missing case for %s", vowel.Name, glyphForm(g))
	}
	tag := cases[0]
	body := tag.Expressions[2]
	data := DataCompilation{Meta: g.Meta, SegmentalSystem: g.SegmentalSystem}
	var prefix, suffix string
	if body.ExpressionOutput != "Variable" && body.ExpressionOutput != "SubTags" {
		call := strings.TrimSpace(body.ExpressionString)
		if !strings.HasPrefix(call, "[|") || !strings.HasSuffix(call, "|]") {
			return lerrors.InvalidConfiguration.Errorf("Rust %s has an invalid primitive call", glyphForm(g))
		}
		start := strings.Index(body.ExpressionString, call)
		prefix, suffix = body.ExpressionString[:start], body.ExpressionString[start+len(call):]
		_, params, selector := g.StartFunctional(call)
		var err error
		data, err = g.functionalArgumentData(params)
		if err != nil {
			return err
		}
		var primitive Segment
		switch body.ExpressionOutput {
		case "Declare":
			primitive = g.SegmentalSystem.Declare
		case "Define":
			primitive = g.SegmentalSystem.Define
		case "Calculate":
			primitive = g.SegmentalSystem.Calculate
		case "Describe":
			primitive = g.SegmentalSystem.Describe
		case "Evaluate":
			primitive = g.SegmentalSystem.Evaluate
		case "Operate":
			primitive = g.SegmentalSystem.Operate
		case "Instantiate":
			primitive = g.SegmentalSystem.Instantiate
		case "Iterate":
			primitive = g.SegmentalSystem.Iterate
		case "Execute":
			primitive = g.SegmentalSystem.Execute
		default:
			return lerrors.InvalidConfiguration.Errorf("unknown Rust primitive %s", body.ExpressionOutput)
		}
		callee := *g
		callee.Meta = data.Meta
		tags := parseSegmentTags(primitive.selectedTemplateTags(callee.SegmentalSystem, selector, false), callee.Meta, selector, false, true)
		if len(tags) == 0 {
			return lerrors.InvalidConfiguration.Errorf("Rust primitive %s missing case %s", primitive.Name, selector)
		}
		tag = tags[0]
	} else {
		data.Conjoinments = g.compiledChildGroups()
	}
	// Source topology owns visibility at this already identified item boundary.
	// Native leaves and nested/local items are never inspected or rewritten.
	g.Content = append([]byte(g.plan.visibility+prefix), tag.Interpret(data)...)
	g.Content = append(g.Content, suffix...)
	return nil
}
