package Lyger

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// orchestrationEmitter binds a checked plan to AML and typed runtime operands.
// Function execution composes selected definitions; source bindings are resolved.
type orchestrationEmitter struct {
	definitionBindings []DefinitionBinding
	definitionActive   map[*record]bool
	definitionDynamic  map[*record]bool
	compilation        *Compilation
	pending            []string
	dependencies       []DefinitionInstance
	valueCount         int
	plan               *orchestrationPlan
	artifacts          map[*App]string
	apps               map[*App]string
	invocation         int
	processSelection   int
	functionDepth      int
	functionWords      *sourceFunction
	functionItems      []string
	functionContracts  map[*record]*sourceFunction
	functionEmitted    map[int]bool
	functionCall       int
	programStorage     map[*orchestrationStorage]int
	programActions     map[*record]int
	processes          *builtinProcessCatalog
}

func (r *CompilationResult) OrchestratorUnit(id string, companions []AppArtifact) (BuildUnit, error) {
	program, err := r.buildProgram()
	if err != nil {
		return BuildUnit{}, err
	}
	if err := validateAMLName(id); err != nil {
		return BuildUnit{}, err
	}
	byID, err := validateBuildGraph(program)
	if err != nil {
		return BuildUnit{}, err
	}
	if r.Compilation.Semantics == nil {
		return BuildUnit{}, lerrors.Internal.Errorf("orchestration requires semantic instructions")
	}
	emitter := &orchestrationEmitter{compilation: r.Compilation, artifacts: make(map[*App]string), apps: make(map[*App]string), processes: r.Compilation.builtinProcesses}
	paths := make(map[string]bool)
	for _, artifact := range companions {
		app := byID[artifact.ID]
		path := artifact.RelativeExecutable
		clean := filepath.Clean(path)
		if app == nil || emitter.artifacts[app] != "" {
			return BuildUnit{}, lerrors.InvalidInput.Errorf("unknown or duplicate companion application %q", artifact.ID)
		}
		if path == "" || strings.ContainsRune(path, 0) || filepath.IsAbs(path) || filepath.VolumeName(path) != "" || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || paths[clean] {
			return BuildUnit{}, lerrors.InvalidInput.Errorf("invalid companion executable path %q", path)
		}
		paths[clean] = true
		emitter.artifacts[app] = filepath.ToSlash(clean)
	}
	if len(emitter.artifacts) != len(byID) {
		return BuildUnit{}, lerrors.InvalidInput.Errorf("orchestration requires exactly one executable for every application")
	}
	emitter.plan, err = planOrchestration(r.Compilation.Semantics)
	if err != nil {
		return BuildUnit{}, err
	}
	emitter.functionContracts = make(map[*record]*sourceFunction)
	emitter.functionEmitted = make(map[int]bool)
	emitter.collectFunctionContracts(emitter.plan.Root.Body)
	for index, app := range emitter.plan.Apps {
		if emitter.artifacts[app] == "" {
			return BuildUnit{}, lerrors.Internal.Errorf("application %s has no companion executable", app.ID)
		}
		emitter.apps[app] = fmt.Sprintf("lApp%d", index)
	}
	var body []string
	for _, app := range emitter.plan.Apps {
		created, err := emitter.word("Storage.App.Create", map[string]DefinitionArgument{
			"Name": definitionReferenceArgument(emitter.apps[app]), "Path": {Kind: "Text", Text: emitter.artifacts[app]},
		})
		if err != nil {
			return BuildUnit{}, err
		}
		body = append(body, created...)
	}
	var actions []*record
	for _, slot := range emitter.plan.Root.Callables {
		actions = append(actions, slot.Record)
	}
	context, err := emitter.programContext(actions)
	if err != nil {
		return BuildUnit{}, err
	}
	body = append(body, emitter.pending...)
	emitter.pending = nil
	body = append(body, context...)
	statements, err := emitter.block(emitter.plan.Root, false)
	if err != nil {
		return BuildUnit{}, err
	}
	body = append(body, statements...)
	completed, err := emitter.word("Storage.Completion", map[string]DefinitionArgument{"Name": definitionReferenceArgument("lCompleted")})
	if err != nil {
		return BuildUnit{}, err
	}
	result, err := emitter.returned("lCompleted")
	if err != nil {
		return BuildUnit{}, err
	}
	body = append(body, completed...)
	body = append(body, result...)
	function, err := r.Compilation.futureEntry("l_main", "runtime::Result<()>", body)
	if err != nil {
		return BuildUnit{}, err
	}
	observedReturn, err := r.Compilation.instantiateBuiltin("Function.Return", map[string]DefinitionArgument{"Statement": definitionReferenceArgument("lObservedStatus")})
	if err != nil {
		return BuildUnit{}, err
	}
	observedBody, err := emitter.word("Storage.Entry.Observe", map[string]DefinitionArgument{"Return": {Kind: "Definition", Definition: &observedReturn}})
	if err != nil {
		return BuildUnit{}, err
	}
	observed, err := r.Compilation.futureEntry("l_entry_with_runtime", "u8", observedBody)
	if err != nil {
		return BuildUnit{}, err
	}
	observer, err := emitter.word("Storage.Test", map[string]DefinitionArgument{"Function": {Kind: "Definition", Definition: &observed}})
	if err != nil {
		return BuildUnit{}, err
	}
	syllables := []*s{{CompiledBytes: []byte("A")}}
	dependencies, err := definitionClosure(emitter.dependencies...)
	if err != nil {
		return BuildUnit{}, err
	}
	for _, dependency := range dependencies {
		syllables = append(syllables, &s{CompiledBytes: []byte(dependency.Expression)})
	}
	syllables = append(syllables, &s{CompiledBytes: []byte(function.Expression)})
	for _, item := range append(observer, emitter.functionItems...) {
		syllables = append(syllables, &s{CompiledBytes: []byte(item)})
	}
	mainObject, mainService := newRootObject("main"), newRootService("main")
	mainService.Objects = []*object{mainObject}
	app := &App{ID: id, Meta: appMetadata(id), Services: []*service{mainService}}
	mainBook := &book{Name: id, Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: []*w{{Syllables: syllables}}}}}}}
	_, backend, err := renderProgram(&Program{Books: []*book{mainBook}}, []*App{app})
	return BuildUnit{ID: id, BackendAML: backend}, err
}

func (e *orchestrationEmitter) declarations(block *orchestrationBlock) ([]string, error) {
	var declarations []string
	for _, slot := range block.Callables {
		if _, exists := e.programActions[slot.Record]; exists {
			continue
		}
		declared, err := e.describedWord("Storage.Binding", []string{"Storage.Empty"}, map[string]DefinitionArgument{
			"Name": definitionReferenceArgument(slot.name()), "Type": definitionTypeArgument("LEnvironment"),
		})
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, declared...)
	}
	for _, slot := range block.Storage {
		if _, shared := e.programStorage[slot]; shared {
			if slot.Initial != nil {
				value, err := e.literalValue(slot.Initial)
				if err != nil {
					return nil, err
				}
				stored, err := e.storeRuntimeValue(slot, value.text)
				if err != nil {
					return nil, err
				}
				declarations = append(declarations, e.pending...)
				e.pending = nil
				declarations = append(declarations, stored...)
			}
			continue
		}
		if slot.Parameter {
			continue
		}
		declared, err := e.storageDeclaration(slot)
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, e.pending...)
		e.pending = nil
		declarations = append(declarations, declared...)
	}
	return declarations, nil
}

func (e *orchestrationEmitter) storageDeclaration(slot *orchestrationStorage) ([]string, error) {
	typeName, err := nativeRuntimeType(slot.RuntimeType)
	if err != nil {
		return nil, err
	}
	arguments := map[string]DefinitionArgument{"Name": definitionReferenceArgument(slot.name())}
	if slot.optionalStorage() {
		// None records unassigned compiler storage, not an invented L value.
		arguments["Type"] = definitionTypeArgument(typeName)
		return e.describedWord("Storage.Binding", []string{"Storage.Empty"}, arguments)
	}
	if slot.Initial != nil {
		value, err := e.literalValue(slot.Initial)
		if err != nil {
			return nil, err
		}
		arguments["Value"] = definitionReferenceArgument(value.text)
		return e.bindValue(slot.name(), value.text, slot.Mutable)
	}
	arguments["Type"] = definitionTypeArgument(typeName)
	return e.word("Storage.Declaration", arguments)
}

func (e *orchestrationEmitter) block(block *orchestrationBlock, scoped bool) ([]string, error) {
	if block == nil {
		return nil, lerrors.Internal.Errorf("planned block is missing")
	}
	outer := e.pending
	e.pending = nil
	defer func() { e.pending = outer }()
	body, err := e.declarations(block)
	if err != nil {
		return nil, err
	}
	statements, err := e.instructions(block.Body)
	if err != nil {
		return nil, err
	}
	body = append(body, statements...)
	if scoped && len(body) != 0 {
		process, err := e.compilation.processDefinition(body)
		if err != nil {
			return nil, err
		}
		word := "Storage.Scope"
		if len(block.Body) != 0 && !block.Body[len(block.Body)-1].ReachesNext {
			word = "Storage.Loop"
		}
		return e.word(word, map[string]DefinitionArgument{"Body": {Kind: "Definition", Definition: &process}})
	}
	return body, nil
}

// emittedExpression retains how a Text is stored so read-only display and
// comparison borrow it, while language assignments explicitly own snapshots.
type emittedExpression struct {
	atom     *DefinitionArgument
	optional bool
	text     string
	kind     valueType
	arc      bool
	borrow   bool
}

func (e *orchestrationEmitter) value(value *orchestrationValue) (emittedExpression, error) {
	if value == nil {
		return emittedExpression{}, lerrors.Internal.Errorf("planned value is missing")
	}
	result := emittedExpression{kind: value.Type}
	if value.Source != nil && value.Source.Atom != nil {
		result.atom = value.Source.Atom
		var err error
		result.text, err = result.atom.metadataAtom()
		return result, err
	}
	if value.Source != nil && (value.Source.Kind == operandConstruct || value.Source.Kind == operandRead) {
		return e.runtimeOperand(value, result)
	}
	if value.Literal != nil {
		return e.literalValue(value.Literal)
	}
	if slot := value.Storage; slot != nil {
		var boxed string
		var err error
		if index, shared := e.programStorage[slot]; shared {
			boxed, err = e.borrowedResultBinding("Storage.Program.Values.Read", map[string]DefinitionArgument{"Program": definitionReferenceArgument("lProgram"), "Index": definitionIndexArgument(uint64(index))}, "Program")
		} else if capture := slot.Capture; capture != nil {
			boxed, err = e.borrowedResultBinding("Storage.Environment.Read", map[string]DefinitionArgument{"Environment": definitionReferenceArgument(capture.Name), "Index": definitionIndexArgument(uint64(capture.Index))}, "Environment")
		} else {
			result.text = slot.name()
			if slot.optionalStorage() {
				if slot.Record == nil {
					return result, lerrors.Internal.Errorf("unassigned storage has no source declaration")
				}
				result.text, err = e.borrowedResultBinding("Storage.Access", map[string]DefinitionArgument{"Value": definitionReferenceArgument(result.text), "Message": {Kind: "Text", Text: "symbol " + slot.Record.ID + " is not initialized"}}, "Value")
				result.borrow = true
			}
			return result, err
		}
		if err != nil {
			return result, err
		}
		result.text, err = e.extractValue(boxed, value.RuntimeType)
		return result, err
	}
	if value.App != nil {
		result.text = e.apps[value.App]
		if result.text == "" {
			return result, lerrors.Internal.Errorf("planned App has no allocated handle")
		}
		return result, nil
	}
	if value.Owner != nil {
		owner, err := e.value(value.Owner)
		if err != nil {
			return result, err
		}
		word := ""
		switch value.Source.Member {
		case memberOutput:
			word = "Storage.App.Output"
			result.arc = true
		case memberError:
			word = "Storage.App.Error"
			result.optional = true
		default:
			return result, lerrors.Internal.Errorf("unsupported planned member %s", value.Source.Member)
		}
		result.text, err = e.resultBinding(word, map[string]DefinitionArgument{"App": definitionReferenceArgument(owner.text)})
		return result, err
	}
	if value.Payload != nil {
		payload, err := e.transferRuntimeValue(value.Payload, value.Payload.RuntimeType)
		if err != nil {
			return result, err
		}
		result.text, err = e.valueBinding("Storage.Error.Create", map[string]DefinitionArgument{"Value": definitionReferenceArgument(payload)})
		return result, err
	}
	return result, lerrors.Internal.Errorf("planned value has no representation")
}

func (value emittedExpression) definitionAtom() DefinitionArgument {
	if value.atom != nil {
		return definitionAtomListArgument(*value.atom)
	}
	return definitionReferenceArgument(value.text)
}

func (e *orchestrationEmitter) condition(condition *orchestrationCondition) (string, error) {
	left, err := e.value(condition.Left)
	if err != nil {
		return "", err
	}
	op, _ := parseConditionOperator(condition.Source.Operator)
	result := left.text
	if op == conditionTruth && left.atom != nil {
		result, err = e.valueBinding("Storage.Binding", map[string]DefinitionArgument{"Value": left.definitionAtom()})
		if err != nil {
			return "", err
		}
	}
	if op == conditionTruth && left.borrow {
		result, err = e.copyValue(left)
		if err != nil {
			return "", err
		}
	}
	if op != conditionTruth {
		right, err := e.value(condition.Right)
		if err != nil {
			return "", err
		}
		if op == conditionHas {
			result, err = e.valueBinding("Storage.Presence", map[string]DefinitionArgument{"Value": right.definitionAtom()})
		} else {
			operator, ok := op.NativeOperator()
			if !ok {
				return "", lerrors.Internal.Errorf("invalid comparison operator")
			}
			if left.kind == valueText || left.borrow {
				left.text, err = e.copyValue(left)
				if err != nil {
					return "", err
				}
			}
			if right.kind == valueText || right.borrow {
				right.text, err = e.copyValue(right)
				if err != nil {
					return "", err
				}
			}
			// The selected definition supplies the calculation; the enum supplies only
			// its already-resolved comparison token and the two operand references.
			result, err = e.valueBinding("Storage.Compare", map[string]DefinitionArgument{"Operator": {Kind: "Syllable", Text: operator}, "Left": left.definitionAtom(), "Right": right.definitionAtom()})
		}
		if err != nil {
			return "", err
		}
	}
	if condition.Source.Negated {
		return e.valueBinding("Storage.Negation", map[string]DefinitionArgument{"Value": definitionReferenceArgument(result)})
	}
	return result, nil
}

func (e *orchestrationEmitter) attribution(current *orchestrationInstruction) ([]string, error) {
	if current.Source.Kind == instructionCall && len(current.Arguments) == 0 || current.Source.Kind == instructionBlock {
		return nil, nil
	}
	span := current.Source.Source
	if current.Source.CallSite != nil && current.Source.CallSite.source != nil {
		span = current.Source.CallSite.source.Span
	}
	seen := make(map[*App]bool)
	var statements []string
	var visit func(*orchestrationValue) error
	visit = func(value *orchestrationValue) error {
		if value == nil {
			return nil
		}
		if value.App != nil && !seen[value.App] {
			seen[value.App] = true
			source := current.Source.Kind.String() + " " + value.App.ID
			if span.Path != "" {
				source = fmt.Sprintf("%s:%d:%d: %s", span.Path, span.Start.Line, span.Start.Column, source)
			}
			words, err := e.word("Storage.App.Attribute", map[string]DefinitionArgument{"App": definitionReferenceArgument(e.apps[value.App]), "Source": {Kind: "Text", Text: source}})
			if err != nil {
				return err
			}
			statements = append(statements, words...)
		}
		for _, child := range append([]*orchestrationValue{value.Owner, value.Payload, value.Key}, value.Elements...) {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	values := append([]*orchestrationValue{current.Target, current.Value}, current.Arguments...)
	if current.Condition != nil {
		values = append(values, current.Condition.Left, current.Condition.Right)
	}
	for _, value := range values {
		if err := visit(value); err != nil {
			return nil, err
		}
	}
	return statements, nil
}

func (e *orchestrationEmitter) instructions(body []*orchestrationInstruction) ([]string, error) {
	outer := e.pending
	e.pending = nil
	defer func() { e.pending = outer }()
	var result []string
	for _, current := range body {
		attribution, err := e.attribution(current)
		if err != nil {
			return nil, current.Source.wrapError(err)
		}
		statements, err := e.instruction(current)
		if err != nil {
			return nil, current.Source.wrapError(err)
		}
		result = append(result, attribution...)
		result = append(result, e.pending...)
		result = append(result, statements...)
		e.pending = nil
	}
	return result, nil
}

func (e *orchestrationEmitter) returned(value string) ([]string, error) {
	return e.word("Function.Return", map[string]DefinitionArgument{"Statement": definitionReferenceArgument(value)})
}

func (e *orchestrationEmitter) instruction(current *orchestrationInstruction) ([]string, error) {
	if current.Source.Application != nil || current.Source.Constructor != nil {
		var instance DefinitionInstance
		var err error
		if current.Source.Constructor != nil {
			instance, _, err = e.compilation.instantiateStaticConstruction(current.Source.Constructor, nil, e.definitionActive, e.definitionDynamic)
		} else {
			instance, _, err = e.compilation.instantiateStaticAction(current.Source.Application, nil, e.definitionActive, e.definitionDynamic)
		}
		if err != nil {
			return nil, err
		}
		e.dependencies = append(e.dependencies, instance.Dependencies...)
		e.definitionBindings = append(e.definitionBindings, instance.Bindings...)
		return builtinProcessExpressions(nil, instance)
	}
	switch current.Source.Kind {
	case instructionFunction, instructionInvoke, instructionReturn:
		return e.functionInstruction(current)
	case instructionBlock:
		return e.block(current.Block, true)
	case instructionBindAction, instructionAliasAction:
		return e.bindCapture(current)
	case instructionCall:
		var bindings []string
		if current.CaptureSource != nil {
			source, err := e.captureReference(current.CaptureSource)
			if err != nil {
				return nil, err
			}
			bound, err := e.word("Storage.Binding", map[string]DefinitionArgument{"Name": definitionReferenceArgument(current.CaptureName), "Value": definitionReferenceArgument(source)})
			if err != nil {
				return nil, err
			}
			bindings = append(bindings, e.pending...)
			e.pending = nil
			bindings = append(bindings, bound...)
		}
		for index, argument := range current.Arguments {
			value, err := e.transferRuntimeValue(argument, argument.RuntimeType)
			if err != nil {
				return nil, err
			}
			slot := current.Results[index].Storage
			bound, err := e.bindValue(slot.name(), value, slot.Mutable)
			if err != nil {
				return nil, err
			}
			bindings = append(bindings, e.pending...)
			e.pending = nil
			bindings = append(bindings, bound...)
		}
		body, err := e.instructions(current.Actions)
		if err != nil {
			return nil, err
		}
		body = append(bindings, body...)
		if len(current.Arguments) != 0 {
			process, err := e.compilation.processDefinition(body)
			if err != nil {
				return nil, err
			}
			word := "Storage.Scope"
			if !current.ReachesNext {
				word = "Storage.Loop"
			}
			return e.word(word, map[string]DefinitionArgument{"Body": {Kind: "Definition", Definition: &process}})
		}
		return body, nil
	case instructionBuild, instructionRun, instructionStop, instructionAwait:
		return e.processInstruction(current)
	case instructionUpdate:
		return e.runtimeUpdate(current)
	case instructionAssign:
		value, err := e.value(current.Value)
		if err != nil {
			return nil, err
		}
		text, err := e.copyValue(value)
		if err != nil {
			return nil, err
		}
		return e.storeRuntimeValue(current.Target.Storage, text)
	case instructionShow:
		value, err := e.value(current.Value)
		if err != nil {
			return nil, err
		}
		boxed, err := e.boxValue(value, current.Value.RuntimeType)
		if err != nil {
			return nil, err
		}
		borrowed, err := e.valueBinding("Storage.Borrow", map[string]DefinitionArgument{"Value": definitionReferenceArgument(boxed)})
		if err != nil {
			return nil, err
		}
		formatted, err := e.valueBinding("Storage.Value.Format", map[string]DefinitionArgument{"Value": definitionReferenceArgument(borrowed)})
		if err != nil {
			return nil, err
		}
		bytes, err := e.checkedBinding(formatted)
		if err != nil {
			return nil, err
		}
		shown, err := e.valueBinding("Storage.Show", map[string]DefinitionArgument{"Value": definitionReferenceArgument(bytes)})
		if err != nil {
			return nil, err
		}
		_, err = e.describedValueBinding("Storage.Binding", []string{"Storage.Checked"}, map[string]DefinitionArgument{"Value": definitionReferenceArgument(shown)})
		return nil, err
	case instructionExit:
		if current.Value == nil {
			if e.functionDepth > 0 {
				value, err := e.valueBinding("Storage.Function.Stop", nil)
				if err != nil {
					return nil, err
				}
				return e.functionOperation(e.functionWords, "Exit", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
			}
			value, err := e.valueBinding("Storage.Completion", nil)
			if err != nil {
				return nil, err
			}
			return e.returned(value)
		}
		value, err := e.transferRuntimeValue(current.Value, current.Value.RuntimeType)
		if err != nil {
			return nil, err
		}
		present, err := e.valueBinding("Storage.Present", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
		if err != nil {
			return nil, err
		}
		failed, err := e.valueBinding("Storage.Program.Fail", map[string]DefinitionArgument{"Value": definitionReferenceArgument(present)})
		if err != nil {
			return nil, err
		}
		return e.returned(failed)
	case instructionIf:
		condition, err := e.condition(current.Condition)
		if err != nil {
			return nil, err
		}
		then, err := e.instructions(current.Then)
		if err != nil {
			return nil, err
		}
		otherwise, err := e.instructions(current.Else)
		if err != nil {
			return nil, err
		}
		return e.selectWords(condition, then, otherwise)
	default:
		return nil, lerrors.Internal.Errorf("unsupported planned instruction %s", current.Source.Kind)
	}
}

func (e *orchestrationEmitter) processReceiver(current *orchestrationInstruction) *builtinProcessReceiver {
	if current.App == nil || current.Target == nil || current.Target.Type != valueApp || current.Target.App != current.App {
		return nil
	}
	return &builtinProcessReceiver{value: current.Target, name: e.apps[current.App]}
}
