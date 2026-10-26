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
	plan              *orchestrationPlan
	artifacts         map[*App]string
	apps              map[*App]string
	invocation        int
	processSelection  int
	functionDepth     int
	functionWords     *sourceFunction
	functionItems     []string
	functionContracts map[*record]*sourceFunction
	functionEmitted   map[int]bool
	functionCall      int
	programStorage    map[*orchestrationStorage]int
	programActions    map[*record]int
	processes         *builtinProcessCatalog
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
	emitter := &orchestrationEmitter{artifacts: make(map[*App]string), apps: make(map[*App]string), processes: r.Compilation.builtinProcesses}
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
		body = append(body, orchestrationBinding(emitter.apps[app], "runtime.app("+quoteOrchestrationString(emitter.artifacts[app])+")"))
	}
	if emitter.plan.RuntimeValues || len(emitter.plan.Callables) > 0 {
		var actions []*record
		for _, slot := range emitter.plan.Root.Callables {
			actions = append(actions, slot.Record)
		}
		context, err := emitter.programContext(actions)
		if err != nil {
			return BuildUnit{}, err
		}
		body = append(body, context...)
	}
	statements, err := emitter.block(emitter.plan.Root, false)
	if err != nil {
		return BuildUnit{}, err
	}
	body = append(body, statements...)
	body = append(body, "BOD^E{|Statement: ~Ok(())~|}")
	body = []string{"BOC^(CV{|Capture: ~Move~|}^(" + orchestrationOperation(body) + "))"}
	function := "FW{|Key: ~l_main~, Inputs: ~&[Prop{Key: runtime, Type: &runtime::Runtime}Prop]&~, Outputs: ~&[impl std::future::Future<Output = runtime::Result<()>> + '_]&~|}^" + orchestrationOperation(body)
	// This private observer seam invokes the same generated source; it is not
	// an alternate entry implementation or a global runtime interception hook.
	observer := "#[cfg(test)]\npub(crate) async fn l_entry_with_runtime(runtime: &runtime::Runtime) -> u8 {\n    runtime.run(l_main(runtime)).await\n}"
	if emitter.plan.RuntimeValues || len(emitter.plan.Callables) > 0 {
		observer = runtimeValueSupport + programRuntimeSupport() + observer
	}
	syllables := []*s{{CompiledBytes: []byte("A")}, {CompiledBytes: []byte(function)}, {CompiledBytes: []byte("DA^" + orchestrationEvaluation(observer))}}
	for _, item := range emitter.functionItems {
		syllables = append(syllables, &s{CompiledBytes: []byte(item)})
	}
	mainObject, mainService := newRootObject("main"), newRootService("main")
	mainService.Objects = []*object{mainObject}
	app := &App{ID: id, Meta: appMetadata(id), Services: []*service{mainService}}
	mainBook := &book{Name: id, Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: []*w{{Syllables: syllables}}}}}}}
	_, backend, err := renderProgram(&Program{Books: []*book{mainBook}}, []*App{app})
	return BuildUnit{ID: id, BackendAML: backend}, err
}

func orchestrationBinding(pattern, value string) string {
	return "OB{|Key: ~" + pattern + "~, Value: ~" + value + "~|}"
}

func orchestrationAssignment(target, value string) string {
	return "O{|Key: ~" + target + "~, Value: ~" + value + "~|}"
}

func (e *orchestrationEmitter) declarations(block *orchestrationBlock) ([]string, error) {
	var declarations []string
	for _, slot := range block.Callables {
		if _, exists := e.programActions[slot.Record]; exists {
			continue
		}
		declarations = append(declarations, orchestrationBinding("mut "+slot.name()+": Option<LEnvironment>", "None"))
	}
	for _, slot := range block.Storage {
		if _, shared := e.programStorage[slot]; shared {
			if slot.Initial != nil {
				value, _, err := nativePrimitiveExpression(slot.Initial)
				if err != nil {
					return nil, err
				}
				stored, err := e.storeRuntimeValue(slot, value)
				if err != nil {
					return nil, err
				}
				declarations = append(declarations, stored...)
			}
			continue
		}
		if slot.Parameter {
			continue
		}
		glyph, _, err := orchestrationStorageDeclaration(slot)
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, glyph)
	}
	return declarations, nil
}

func orchestrationStorageDeclaration(slot *orchestrationStorage) (glyph, native string, err error) {
	typeName, err := nativeRuntimeType(slot.RuntimeType)
	if err != nil {
		return "", "", err
	}
	name := slot.name()
	if slot.Published {
		// None is compiler state, never an initial L value. The plan rejects
		// unassigned reads; checked access also guards cached joins at runtime.
		pattern := "mut " + name + ": Option<" + typeName + ">"
		return orchestrationBinding(pattern, "None"), "let " + pattern + " = None;", nil
	} else if slot.Initial != nil {
		initial, _, err := nativePrimitiveExpression(slot.Initial)
		if err != nil {
			return "", "", err
		}
		pattern := name + ": " + typeName
		if slot.Mutable {
			pattern = "mut " + pattern
		}
		return orchestrationBinding(pattern, initial), "let " + pattern + " = " + initial + ";", nil
	} else if slot.Mutable {
		// Frozen BO does not consume the Mutable projection. A typed
		// declaration atom expresses inferred mutability without a new input
		// contract, initializer or template rewrite.
		native := "let mut " + name + ": " + typeName + ";"
		return orchestrationStatement(native), native, nil
	} else {
		return "BO{|Key: ~" + name + "~, Type: ~" + typeName + "~|}", "let " + name + ": " + typeName + ";", nil
	}
}

func (e *orchestrationEmitter) block(block *orchestrationBlock, scoped bool) ([]string, error) {
	if block == nil {
		return nil, lerrors.Internal.Errorf("planned block is missing")
	}
	body, err := e.declarations(block)
	if err != nil {
		return nil, err
	}
	statements, err := e.instructions(block.Body)
	if err != nil {
		return nil, err
	}
	body = append(body, statements...)
	if scoped {
		body = append([]string{orchestrationStatement("{")}, body...)
		body = append(body, orchestrationStatement("}"))
	}
	return body, nil
}

// emittedExpression retains how a Text is stored so read-only display and
// comparison borrow it, while language assignments explicitly own snapshots.
type emittedExpression struct {
	literal        bool
	optional       bool
	text           string
	kind           valueType
	typeExpression string
	arc            bool
	borrow         bool
}

func (value emittedExpression) textReference() string {
	if value.arc {
		return value.text + ".as_ref()"
	}
	if value.borrow {
		return value.text
	}
	return "&" + value.text
}

func (value emittedExpression) textSlice() string {
	if value.arc {
		return value.text + ".as_ref()"
	}
	return "(" + value.text + ").as_slice()"
}

func (e *orchestrationEmitter) value(value *orchestrationValue) (emittedExpression, error) {
	if value == nil {
		return emittedExpression{}, lerrors.Internal.Errorf("planned value is missing")
	}
	result := emittedExpression{kind: value.Type}
	var err error
	if value.Type == valueFunction {
		result.typeExpression, err = runtimeTypeExpression(value.RuntimeType, e.plan.Identities)
		if err != nil {
			return result, err
		}
	}
	if value.Source != nil && (value.Source.Kind == operandConstruct || value.Source.Kind == operandRead) {
		return e.runtimeOperand(value, result)
	}
	if value.Literal != nil {
		text, _, err := nativePrimitiveExpression(value.Literal)
		if err != nil {
			return result, err
		}
		result.text, result.literal = text, true
		return result, nil
	}
	if value.Storage != nil {
		if index, shared := e.programStorage[value.Storage]; shared {
			text, err := unboxTypedRuntimeValue(fmt.Sprintf("lProgram.get(%d)?", index), value.RuntimeType)
			result.text = text
			return result, err
		}
		if capture := value.Storage.Capture; capture != nil {
			result.text, err = unboxTypedRuntimeValue(fmt.Sprintf("%s.get(%d)?", capture.Name, capture.Index), value.RuntimeType)
			return result, err
		}
		result.text = value.Storage.name()
		if value.Storage.Published {
			result.text += ".as_ref().ok_or_else(|| runtime::Error::new(" + quoteOrchestrationString("symbol "+value.Record.ID+" is not initialized") + "))?"
			if value.Type != valueNumber && value.Type != valueTruth && value.Type != valueNull {
				result.borrow = true
			} else {
				result.text = "(*" + result.text + ")"
			}
		}
		return result, nil
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
		switch value.Source.Member {
		case memberOutput:
			result.text, result.arc = owner.text+".output()?", true
		case memberError:
			result.text, result.optional = owner.text+".error()?", true
		default:
			return result, lerrors.Internal.Errorf("unsupported planned member %s", value.Source.Member)
		}
		return result, nil
	}
	if value.Payload != nil {
		payload, err := e.value(value.Payload)
		if err != nil {
			return result, err
		}
		text := payload.text
		if payload.arc || value.Payload.Access == orchestrationAccessBorrow || value.Payload.Access == orchestrationAccessSnapshot {
			text = "(" + text + ").to_vec()"
		}
		result.text = "runtime::Error::from_bytes(" + text + ")"
		return result, nil
	}
	return result, lerrors.Internal.Errorf("planned value has no representation")
}

func (e *orchestrationEmitter) condition(condition *orchestrationCondition) (string, error) {
	left, err := e.value(condition.Left)
	if err != nil {
		return "", err
	}
	op, _ := parseConditionOperator(condition.Source.Operator)
	text := left.text
	if op != conditionTruth {
		right, err := e.value(condition.Right)
		if err != nil {
			return "", err
		}
		if op == conditionHas {
			text = right.text + ".is_some()"
		} else {
			operator, _ := op.NativeOperator()
			if left.kind == valueText {
				text = left.textSlice() + " " + operator + " " + right.textSlice()
			} else {
				text = left.text + " " + operator + " " + right.text
			}
		}
	}
	if condition.Source.Negated {
		text = "!(" + text + ")"
	}
	return text, nil
}

func (e *orchestrationEmitter) attribution(current *orchestrationInstruction) []string {
	if current.Source.Kind == instructionCall && len(current.Arguments) == 0 || current.Source.Kind == instructionBlock {
		return nil
	}
	span := current.Source.Source
	if current.Source.CallSite != nil && current.Source.CallSite.source != nil {
		span = current.Source.CallSite.source.Span
	}
	seen := make(map[*App]bool)
	var statements []string
	var visit func(*orchestrationValue)
	visit = func(value *orchestrationValue) {
		if value == nil {
			return
		}
		if value.App != nil && !seen[value.App] {
			seen[value.App] = true
			source := current.Source.Kind.String() + " " + value.App.ID
			if span.Path != "" {
				source = fmt.Sprintf("%s:%d:%d: %s", span.Path, span.Start.Line, span.Start.Column, source)
			}
			statements = append(statements, orchestrationStatement(e.apps[value.App]+".set_source("+quoteOrchestrationString(source)+");"))
		}
		visit(value.Owner)
		visit(value.Payload)
		visit(value.Key)
		for _, element := range value.Elements {
			visit(element)
		}
	}
	visit(current.Target)
	visit(current.Value)
	for _, argument := range current.Arguments {
		visit(argument)
	}
	if current.Condition != nil {
		visit(current.Condition.Left)
		visit(current.Condition.Right)
	}
	return statements
}

func (e *orchestrationEmitter) instructions(body []*orchestrationInstruction) ([]string, error) {
	var result []string
	for _, current := range body {
		statements, err := e.instruction(current)
		if err != nil {
			return nil, current.Source.wrapError(err)
		}
		result = append(result, e.attribution(current)...)
		result = append(result, statements...)
	}
	return result, nil
}

func (e *orchestrationEmitter) instruction(current *orchestrationInstruction) ([]string, error) {
	switch current.Source.Kind {
	case instructionFunction, instructionInvoke, instructionReturn:
		return e.functionInstruction(current)
	case instructionBlock:
		return e.block(current.Block, true)
	case instructionBindAction, instructionAliasAction:
		return e.bindCapture(current)
	case instructionCall:
		body, err := e.instructions(current.Actions)
		if err != nil {
			return nil, err
		}
		if current.CaptureSource != nil {
			source, err := e.captureReference(current.CaptureSource)
			if err != nil {
				return nil, err
			}
			body = append([]string{orchestrationBinding(current.CaptureName, source)}, body...)
		}
		if len(current.Arguments) != 0 {
			bindings := []string{orchestrationStatement("{")}
			for index, argument := range current.Arguments {
				text, err := e.transferRuntimeValue(argument, argument.RuntimeType)
				if err != nil {
					return nil, err
				}
				slot := current.Results[index].Storage
				typ, err := nativeRuntimeType(slot.RuntimeType)
				if err != nil {
					return nil, err
				}
				pattern := slot.name() + ": " + typ
				if slot.Mutable {
					pattern = "mut " + pattern
				}
				bindings = append(bindings, orchestrationBinding(pattern, text))
			}
			body = append(append(bindings, body...), orchestrationStatement("}"))
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
		text, err := runtimeCopyExpression(value, current.Value.Access)
		if err != nil {
			return nil, err
		}
		return e.storeRuntimeValue(current.Target.Storage, text)
	case instructionShow:
		value, err := e.value(current.Value)
		if err != nil {
			return nil, err
		}
		text := value.textReference()
		if value.kind == valueNumber {
			text = "&runtime::format_number(" + value.text + ")"
		} else if value.kind == valueTruth {
			text = "&" + value.text + ".to_string().into_bytes()"
		} else if value.kind == valueBytes {
			text = value.textReference()
		} else if value.kind != valueText {
			boxed, err := boxRuntimeValue(value)
			if err != nil {
				return nil, err
			}
			text = "&(" + boxed + ").display_bytes()?"
		}
		return []string{orchestrationStatement("runtime.show(" + text + ").await?;")}, nil
	case instructionExit:
		if current.Value == nil {
			if e.functionDepth > 0 {
				return instantiateFunctionWord(e.functionWords, "Exit", map[string]string{"Value": "std::ops::ControlFlow::Break(())"})
			}
			return []string{orchestrationStatement("return Ok(());")}, nil
		}
		value, err := e.value(current.Value)
		if err != nil {
			return nil, err
		}
		return []string{orchestrationStatement("let lExit = runtime.fail(" + value.text + ");"), orchestrationStatement("return Err(lExit);")}, nil
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
		parts := []string{orchestrationEvaluation(condition), orchestrationOperation(then)}
		if len(current.Source.Else) > 0 {
			parts = append(parts, orchestrationEvaluation("Void"), orchestrationOperation(otherwise))
		}
		return []string{"BOG^GI^(" + strings.Join(parts, "*") + ")"}, nil
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
