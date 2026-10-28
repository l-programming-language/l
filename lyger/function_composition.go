package Lyger

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

type orchestrationFunction struct {
	Source         *sourceFunction
	Body           *orchestrationBlock
	Inputs         []*orchestrationStorage
	Self           *orchestrationStorage
	Captures       []*orchestrationValue
	CaptureHandles []*orchestrationCallableReference
}

func (p *orchestrationPlanner) functionInstruction(current *instruction, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	result := &orchestrationInstruction{Source: current, Effects: orchestrationEffectAssign, ReachesNext: flow.live}
	if current.Kind == instructionReturn {
		for _, source := range current.Arguments {
			value, err := p.value(source, flow)
			if err != nil {
				return nil, flow, err
			}
			result.Arguments = append(result.Arguments, value)
		}
		flow.live = false
		result.ReachesNext = false
		return result, flow, nil
	}
	if current.Kind == instructionInvoke {
		var err error
		result.Target, err = p.value(current.Target, flow)
		if err != nil {
			return nil, flow, err
		}
		for _, source := range current.Arguments {
			value, err := p.value(source, flow)
			if err != nil {
				return nil, flow, err
			}
			result.Arguments = append(result.Arguments, value)
		}
		for _, source := range current.Results {
			rec, kind, err := p.record(source)
			if err != nil {
				return nil, flow, err
			}
			slot, _, err := p.storage(rec, kind)
			if err != nil {
				return nil, flow, err
			}
			result.Results = append(result.Results, &orchestrationValue{Source: source, Record: rec, Type: kind, RuntimeType: slot.RuntimeType, Storage: slot})
			flow.assigned[slot], flow.written[slot] = true, true
		}
		return result, flow, nil
	}
	source := current.Function
	function := &orchestrationFunction{Source: source, Body: p.newBlock(source.Scope)}
	result.Function = function
	for _, capture := range source.Captures {
		value, err := p.value(capture, flow)
		if err != nil {
			return nil, flow, err
		}
		value.Access = orchestrationAccessSnapshot
		function.Captures = append(function.Captures, value)
	}
	for _, binding := range source.CaptureCalls {
		handle, err := p.captureReference(binding, flow)
		if err != nil {
			return nil, flow, err
		}
		function.CaptureHandles = append(function.CaptureHandles, handle)
	}
	slot, _, err := p.storage(source.Record, valueFunction)
	if err != nil {
		return nil, flow, err
	}
	result.Target = &orchestrationValue{Source: current.Target, Record: source.Record, Type: valueFunction, RuntimeType: source.signature(), Storage: slot}
	flow.assigned[slot], flow.written[slot] = true, true
	p.registerRuntimeType(source.signature())
	frame := &orchestrationFrame{block: function.Body, slots: make(map[*record]*orchestrationStorage)}
	p.frames = append(p.frames, frame)
	context := orchestrationCaptureContext{values: make(map[*record]*orchestrationStorage), calls: make(map[*record]*orchestrationCallableReference)}
	for index, value := range function.Captures {
		context.values[value.Record] = &orchestrationStorage{Record: value.Record, Type: value.Type, RuntimeType: value.RuntimeType, Mutable: true, Capture: &orchestrationCaptureAccess{Name: "lEnvironment", Index: index}}
	}
	for index, binding := range source.CaptureCalls {
		context.calls[binding] = &orchestrationCallableReference{Layout: function.CaptureHandles[index].Layout, Capture: &orchestrationCaptureAccess{Name: "lEnvironment", Index: len(function.Captures) + index}}
	}
	p.captureContexts = append(p.captureContexts, context)
	defer func() {
		p.frames = p.frames[:len(p.frames)-1]
		p.captureContexts = p.captureContexts[:len(p.captureContexts)-1]
	}()
	inner := flow.clone()
	p.functionDepth++
	defer func() { p.functionDepth-- }()
	allocate := func(rec *record, typ runtimeTypeRef) *orchestrationStorage {
		s := &orchestrationStorage{ID: len(p.plan.Storage), Record: rec, Type: typ.Kind, RuntimeType: typ, Owner: function.Body, Parameter: true}
		p.plan.Storage = append(p.plan.Storage, s)
		function.Body.Storage = append(function.Body.Storage, s)
		frame.slots[rec] = s
		inner.assigned[s] = true
		return s
	}
	function.Self = allocate(source.Record, source.signature())
	context.values[source.Record] = function.Self
	for _, param := range source.Inputs {
		function.Inputs = append(function.Inputs, allocate(param.Record, param.Type))
	}
	function.Body.Body, inner, err = p.instructions(source.Body, inner)
	if err != nil {
		return nil, flow, err
	}
	if inner.live && len(source.Outputs) > 0 {
		return nil, flow, lerrors.InvalidInput.Errorf("Function %s does not return on every continuing path", source.Record.ID)
	}
	return result, flow, nil
}

// Native Function values retain an environment and a typed Rust function pointer.
// The signature is compiled once; invocation never dispatches boxed L arguments.
func nativeFunctionPointer(typ runtimeTypeRef) (string, error) {
	return nativeFunctionPointerAt(typ, 0)
}
func nativeFunctionPointerAt(typ runtimeTypeRef, depth int) (string, error) {
	lifetime := fmt.Sprintf("'call%d", depth)
	inputs := []string{"&" + lifetime + " (runtime::Runtime, LProgram)", "LEnvironment"}
	for _, input := range typ.Inputs {
		name, err := nativeRuntimeTypeAt(input, depth+1)
		if err != nil {
			return "", err
		}
		inputs = append(inputs, name)
	}
	output, err := nativeFunctionFuture(typ.Outputs, depth+1, lifetime)
	if err != nil {
		return "", err
	}
	return "for<" + lifetime + "> fn(" + strings.Join(inputs, ", ") + ") -> " + output, nil
}

func nativeFunctionFuture(outputs []runtimeTypeRef, depth int, lifetime string) (string, error) {
	result, err := nativeFunctionOutputAt(outputs, depth)
	return "std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<std::ops::ControlFlow<(), " + result + ">>> + " + lifetime + ">>", err
}
func nativeFunctionOutputAt(outputs []runtimeTypeRef, depth int) (string, error) {
	parts := make([]string, len(outputs))
	for i, output := range outputs {
		name, err := nativeRuntimeTypeAt(output, depth)
		if err != nil {
			return "", err
		}
		parts[i] = name
	}
	if len(parts) == 1 {
		return "(" + parts[0] + ",)", nil
	}
	return "(" + strings.Join(parts, ", ") + ")", nil
}

func nativeRuntimeType(typ runtimeTypeRef) (string, error) { return nativeRuntimeTypeAt(typ, 0) }
func nativeRuntimeTypeAt(typ runtimeTypeRef, depth int) (string, error) {
	if typ.Kind == valueFunction {
		pointer, err := nativeFunctionPointerAt(typ, depth)
		return "(LEnvironment, " + pointer + ")", err
	}
	name, ok := runtimeTypeName(typ.Kind)
	if !ok {
		return "", lerrors.Unsupported.Errorf("no native storage for %s", typ.Kind)
	}
	return name, nil
}

func (e *orchestrationEmitter) functionOperation(function *sourceFunction, member string, bindings map[string]DefinitionArgument) ([]string, error) {
	if function == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("Function requires its %s composition", member)
	}
	if member == "Yield" || member == "Exit" {
		name := "lSuccess"
		if member == "Exit" {
			name = "lExit"
		}
		if function.ReturnTemplate == nil {
			return nil, lerrors.InvalidConfiguration.Errorf("Function requires its Return composition")
		}
		returned, err := function.ReturnTemplate.Instantiate(map[string]DefinitionArgument{"Statement": definitionReferenceArgument(name)})
		if err != nil {
			return nil, err
		}
		bindings["Return"] = DefinitionArgument{Kind: "Syllable", Definition: &returned}
	}
	return e.word("Storage.Function."+member, bindings)
}

func (e *orchestrationEmitter) collectFunctionContracts(body []*orchestrationInstruction) {
	for _, current := range body {
		if current.Function != nil {
			function := current.Function
			e.functionContracts[function.Source.Definition] = function.Source
			e.collectFunctionContracts(function.Body.Body)
		}
		if current.Block != nil {
			e.collectFunctionContracts(current.Block.Body)
		}
		e.collectFunctionContracts(current.Actions)
		e.collectFunctionContracts(current.Then)
		e.collectFunctionContracts(current.Else)
		if current.Continuation != nil {
			e.collectFunctionContracts(current.Continuation.Body)
		}
	}
}

func (e *orchestrationEmitter) functionInstruction(current *orchestrationInstruction) ([]string, error) {
	var types []runtimeTypeRef
	if current.Source.Kind == instructionReturn {
		types = current.Source.Function.Outputs
	} else if current.Target != nil {
		types = current.Target.RuntimeType.Inputs
	}
	if len(types) != len(current.Arguments) && current.Source.Kind != instructionFunction {
		return nil, lerrors.Internal.Errorf("Function transfer does not match its signature")
	}
	var arguments []DefinitionArgument
	for i, value := range current.Arguments {
		argument, err := e.transferRuntimeValue(value, types[i])
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, definitionReferenceArgument(argument))
	}
	if current.Source.Kind == instructionReturn {
		value, err := e.results(arguments)
		if err != nil {
			return nil, err
		}
		return e.functionOperation(e.functionWords, "Yield", map[string]DefinitionArgument{"Values": definitionReferenceArgument(value)})
	}
	if current.Source.Kind == instructionInvoke {
		contract := e.functionContracts[current.Target.RuntimeType.Definition]
		if contract == nil {
			return nil, lerrors.Internal.Errorf("invocation has no bound Function composition")
		}
		target, err := e.transferRuntimeValue(current.Target, current.Target.RuntimeType)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("lCalled%d", e.functionCall)
		e.functionCall++
		returnContract := contract
		var stopped string
		if e.functionDepth > 0 {
			returnContract = e.functionWords
			stopped, err = e.valueBinding("Storage.Function.Stop", nil)
		} else {
			stopped, err = e.results(nil)
		}
		if err != nil {
			return nil, err
		}
		exit, err := e.functionOperation(returnContract, "Exit", map[string]DefinitionArgument{"Value": definitionReferenceArgument(stopped)})
		if err != nil {
			return nil, err
		}
		exitProcess, err := e.processDefinition(exit)
		if err != nil {
			return nil, err
		}
		if contract.Members["Invoke"] == nil {
			return nil, lerrors.InvalidConfiguration.Errorf("Function requires its Invoke composition")
		}
		invoked := append([]DefinitionArgument{definitionReferenceArgument(name + "Context"), definitionReferenceArgument(name + "Environment")}, arguments...)
		invocation, err := contract.Members["Invoke"].Instantiate(map[string]DefinitionArgument{
			"Target": definitionReferenceArgument(name + ".1"), "Arguments": definitionAtomListArgument(invoked...),
		})
		if err != nil {
			return nil, err
		}
		copier, err := e.functionReference("Storage.Environment.Copy")
		if err != nil {
			return nil, err
		}
		result, err := e.word("Storage.Invoke", map[string]DefinitionArgument{
			"Value": definitionReferenceArgument(target), "Name": definitionReferenceArgument(name), "Context": definitionReferenceArgument("lContext"),
			"Invocation": {Kind: "Definition", Definition: &invocation}, "Exit": {Kind: "Syllable", Definition: &exitProcess}, "Copy": copier,
		})
		if err != nil {
			return nil, err
		}
		for index, target := range current.Results {
			start := len(e.pending)
			stored, err := e.storeRuntimeValue(target.Storage, fmt.Sprintf("%sResult.%d", name, index))
			if err != nil {
				return nil, err
			}
			// Result transfers must follow the invocation, including any composed
			// boxing and shared-slot operations queued by their storage word.
			result = append(result, e.pending[start:]...)
			e.pending = e.pending[:start]
			result = append(result, stored...)
		}
		return result, nil
	}
	return e.constructFunction(current)
}

func (e *orchestrationEmitter) constructFunction(current *orchestrationInstruction) ([]string, error) {
	function := current.Function
	captureTypes, captures, err := e.captureEnvironment(function.Captures, function.CaptureHandles)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("lFunction%d", function.Body.ID)
	pointer, err := nativeFunctionPointer(function.Source.signature())
	if err != nil {
		return nil, err
	}
	inputs := []DefinitionProperty{{Key: "lContext", Type: "&(runtime::Runtime, LProgram)"}, {Key: "lEnvironment", Type: "LEnvironment"}}
	body, err := e.word("Storage.Context.Read", map[string]DefinitionArgument{"Context": definitionReferenceArgument("lContext")})
	if err != nil {
		return nil, err
	}
	copier, err := e.functionReference("Storage.Environment.Copy")
	if err != nil {
		return nil, err
	}
	self, err := e.word("Storage.Function.Retain", map[string]DefinitionArgument{"Name": definitionReferenceArgument(function.Self.name()), "Environment": definitionReferenceArgument("lEnvironment"), "Function": definitionReferenceArgument(name), "Type": definitionTypeArgument(pointer), "Copy": copier})
	if err != nil {
		return nil, err
	}
	body = append(body, self...)
	for index, app := range e.plan.Apps {
		bound, err := e.namedBinding(e.apps[app]+"Outcome", "Storage.Program.Select", map[string]DefinitionArgument{"Program": definitionReferenceArgument("lProgram"), "Index": definitionIndexArgument(uint64(index))})
		if err != nil {
			return nil, err
		}
		checked, err := e.describedWord("Storage.Binding", []string{"Storage.Checked"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument(e.apps[app]), "Value": definitionReferenceArgument(e.apps[app] + "Outcome")})
		bound = append(bound, checked...)
		if err != nil {
			return nil, err
		}
		body = append(body, bound...)
	}
	for _, parameter := range function.Inputs {
		typ, err := nativeRuntimeType(parameter.RuntimeType)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, DefinitionProperty{Key: parameter.name(), Type: typ})
		if parameter.Mutable {
			mutable, err := e.bindValue(parameter.name(), parameter.name(), true)
			if err != nil {
				return nil, err
			}
			body = append(body, mutable...)
		}
	}
	previousWords := e.functionWords
	e.functionWords, e.functionDepth = function.Source, e.functionDepth+1
	statements, err := e.block(function.Body, false)
	e.functionWords, e.functionDepth = previousWords, e.functionDepth-1
	if err != nil {
		return nil, err
	}
	body = append(body, statements...)
	if len(function.Source.Outputs) == 0 {
		emptyName := fmt.Sprintf("lFunction%dEmpty", function.Body.ID)
		empty, err := e.describedWord("Storage.Tuple", []string{"Storage.Empty"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument(emptyName)})
		if err != nil {
			return nil, err
		}
		tail, err := e.functionOperation(function.Source, "Yield", map[string]DefinitionArgument{"Values": definitionReferenceArgument(emptyName)})
		if err != nil {
			return nil, err
		}
		body = append(body, empty...)
		body = append(body, tail...)
	}
	bodyProcess, err := e.processDefinition(body)
	if err != nil {
		return nil, err
	}
	process, err := e.functionOperation(function.Source, "Suspend", map[string]DefinitionArgument{"Body": {Kind: "Syllable", Definition: &bodyProcess}})
	if err != nil {
		return nil, err
	}
	suspended, err := e.processDefinition(process)
	if err != nil {
		return nil, err
	}
	output, err := nativeFunctionFuture(function.Source.Outputs, 0, "'_")
	if err != nil {
		return nil, err
	}
	signature, err := function.Source.Template.signatureArguments(name, inputs, []string{output})
	if err != nil {
		return nil, err
	}
	for _, slot := range function.Source.Template.Slots {
		if _, present := signature[slot.Name]; present {
			continue
		}
		if slot.Type != "" {
			return nil, lerrors.Unsupported.Errorf("unsupported Function body slot %s", slot.Name)
		}
		signature[slot.Name] = DefinitionArgument{Kind: "Definition", Definition: &suspended}
	}
	definition, err := function.Source.Template.Instantiate(signature)
	if err != nil {
		return nil, err
	}
	if !e.functionEmitted[function.Body.ID] {
		e.functionItems = append(e.functionItems, definition.Expression)
		e.functionEmitted[function.Body.ID] = true
	}
	created := fmt.Sprintf("lCreatedFunction%d", function.Body.ID)
	environment, err := e.compilation.instantiateBuiltin("Storage.Environment.Create", map[string]DefinitionArgument{"Types": definitionReferenceArgument(captureTypes), "Values": definitionReferenceArgument(captures)})
	if err != nil {
		return nil, err
	}
	bindings, err := e.functionOperation(function.Source, "Bind", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(created), "Function": definitionReferenceArgument(name), "Type": definitionTypeArgument(pointer),
		"Environment": {Kind: "Definition", Definition: &environment},
	})
	if err != nil {
		return nil, err
	}
	// Publish the completed value only after the selected construction words.
	start := len(e.pending)
	stored, err := e.storeRuntimeValue(current.Target.Storage, created)
	bindings = append(bindings, e.pending[start:]...)
	e.pending = e.pending[:start]
	return append(bindings, stored...), err
}
