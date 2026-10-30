package Lyger

import "github.com/l-programming-language/l/lerrors"

// Captures use resolved definition-site bindings, before actual arguments are
// substituted. Code identity, a source name binding and an environment instance
// remain separate throughout lowering.
func (b *instructionBuilder) prepareCaptures() error {
	b.captures = make(map[*record]*instruction)
	for _, ordered := range b.scopes.Ordered {
		if unit := ordered.Statement.source.Unit; unit != nil && unit.Kind != SourceProgram {
			continue
		}
		for _, rec := range ordered.Statement.records {
			if rec.Value == nil || rec.Value.Type != storageLexeme || b.semantics.Aliases[rec] != nil {
				continue
			}
			body := directiveClauses(ordered.Statement)
			if len(body) == 0 {
				continue
			}
			event := &instruction{Kind: instructionBindAction, Callable: rec, Target: &instructionValue{Kind: operandSymbol, Type: valueLexeme, Record: rec}, Definition: ordered.Statement}
			seen := make(map[*record]bool)
			seenCalls := make(map[*record]bool)
			parameters := make(map[*token]bool)
			for _, parameter := range ordered.Arguments {
				parameters[parameter.Token] = true
			}
			for _, clause := range body {
				for _, tok := range clause.tokens {
					if parameters[tok] || (tok.binding == nil && !tok.isPronoun()) {
						continue
					}
					if pronoun, ok := b.scopes.Pronouns[tok]; ok && pronoun.Kind == scopeActionArgument {
						continue
					}
					found, err := b.reference(tok)
					if err != nil {
						return err
					}
					if found == nil || found == rec || seen[found] {
						continue
					}
					kind, err := instructionType(found)
					if err != nil {
						return err
					}
					if kind == valueLexeme {
						binding := found
						if tok.binding != nil {
							binding = tok.binding.Record
						}
						if _, native := tok.nativeRole(); native {
							continue
						}
						if definition := b.semantics.Declarations[found]; definition != nil && len(directiveClauses(definition)) > 0 && !seenCalls[binding] {
							event.CaptureCalls = append(event.CaptureCalls, binding)
							seenCalls[binding] = true
						}
					} else if (kind != valueApp || found.Value.Runtime != nil) && kind != valueClass && kind != valueObject && kind != valueDictionary && kind != valueService {
						event.Captures = append(event.Captures, &instructionValue{Kind: operandSymbol, Type: kind, Record: found, ID: found.ID, Token: tok, Runtime: found.Value.Runtime != nil})
						seen[found] = true
					}
				}
			}
			b.captures[rec] = event
		}
	}
	return nil
}

// declarationInstruction keeps Function and action bindings in ordinary control flow.
func (b *instructionBuilder) declarationInstruction(site registrationSite, ctx instructionContext) (*instruction, error) {
	target, err := site.target()
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, nil
	}
	rec := b.semantics.Subjects[target]
	if rec == nil {
		return nil, nil
	}
	if function := b.semantics.Functions[rec]; function != nil && !function.Native && function.Header == ctx.statement {
		current := ctx.newInstruction(instructionFunction)
		current.Function = function
		current.Target = &instructionValue{Kind: operandSymbol, Type: valueFunction, RuntimeType: function.signature(), Record: rec, ID: rec.ID, Runtime: true}
		return current, nil
	}
	if site.isAlias() {
		from, _, err := site.aliasOperands()
		if err != nil {
			return nil, err
		}
		if from.binding == nil {
			return nil, lerrors.Internal.Errorf("alias source has no binding")
		}
		canonical, err := canonicalRecord(from.binding.Record, b.semantics.Aliases)
		if err != nil {
			return nil, err
		}
		if b.captures[canonical] == nil {
			return nil, nil
		}
		event := ctx.newInstruction(instructionAliasAction)
		event.Callable = rec
		event.SourceToken = target
		event.Target = &instructionValue{Kind: operandSymbol, Type: valueLexeme, Record: from.binding.Record, Token: from}
		return event, nil
	}
	template := b.captures[rec]
	if template == nil {
		return nil, nil
	}
	event := ctx.newInstruction(instructionBindAction)
	event.Callable = rec
	event.Target = template.Target
	event.Captures = template.Captures
	event.CaptureCalls = template.CaptureCalls
	event.Definition = template.Definition
	event.SourceToken = target
	return event, nil
}

// Immutable compile-time captures keep their existing expansion and introduce
// no executable definition event. A dynamic callee is itself a captured handle.
func (b *instructionBuilder) activateCaptures(body []*instruction) ([]*instruction, error) {
	temporary := &semanticProgram{Instructions: body, Aliases: b.semantics.Aliases}
	dynamic, err := temporary.runtimeRecords()
	if err != nil {
		return nil, err
	}
	enabled := make(map[*record]bool)
	for rec, event := range b.captures {
		for _, value := range event.Captures {
			if dynamic[value.Record] || value.Record.Value.Runtime != nil {
				enabled[rec] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for rec, event := range b.captures {
			for _, binding := range event.CaptureCalls {
				canonical, err := canonicalRecord(binding, b.semantics.Aliases)
				if err != nil {
					return nil, err
				}
				if enabled[canonical] && !enabled[rec] {
					enabled[rec] = true
					changed = true
				}
			}
		}
	}
	var filter func([]*instruction) ([]*instruction, error)
	filter = func(input []*instruction) ([]*instruction, error) {
		var output []*instruction
		for _, current := range input {
			if current.Kind == instructionBindAction && !enabled[current.Callable] {
				continue
			}
			if current.Kind == instructionAliasAction {
				canonical, err := canonicalRecord(current.Target.Record, b.semantics.Aliases)
				if err != nil {
					return nil, err
				}
				if !enabled[canonical] {
					continue
				}
			}
			if current.Kind == instructionCall {
				canonical, err := canonicalRecord(current.Callable, b.semantics.Aliases)
				if err != nil {
					return nil, err
				}
				if !enabled[canonical] {
					current.Callable = nil
				}
			}
			if current.Kind == instructionBindAction {
				var calls []*record
				for _, binding := range current.CaptureCalls {
					canonical, err := canonicalRecord(binding, b.semantics.Aliases)
					if err != nil {
						return nil, err
					}
					if enabled[canonical] {
						calls = append(calls, binding)
					}
				}
				current.CaptureCalls = calls
			}
			for _, nested := range []*[]*instruction{&current.Actions, &current.Then, &current.Else} {
				*nested, err = filter(*nested)
				if err != nil {
					return nil, err
				}
			}
			if current.Block != nil {
				current.Block.Body, err = filter(current.Block.Body)
				if err != nil {
					return nil, err
				}
			}
			if current.Kind == instructionFunction && current.Function != nil {
				current.Function.Body, err = filter(current.Function.Body)
				if err != nil {
					return nil, err
				}
			}
			output = append(output, current)
		}
		return output, nil
	}
	return filter(body)
}
