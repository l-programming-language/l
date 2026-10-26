package Lyger

import (
	"fmt"
	"github.com/l-programming-language/l/lerrors"
	"strings"
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
	p.plan.RuntimeValues = true
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

func instantiateFunctionWord(function *sourceFunction, member string, values map[string]string) ([]string, error) {
	if function == nil || function.Members[member] == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("Function requires its %s composition", member)
	}
	bindings := make(map[string]DefinitionArgument, len(values))
	for key, value := range values {
		bindings[key] = DefinitionArgument{Kind: "Syllable", Text: value}
	}
	if member == "Yield" || member == "Exit" {
		name := "lSuccess"
		if member == "Exit" {
			name = "lExit"
		}
		if function.ReturnTemplate == nil {
			return nil, lerrors.InvalidConfiguration.Errorf("Function requires its Return composition")
		}
		returned, err := function.ReturnTemplate.Instantiate(map[string]DefinitionArgument{"Statement": {Kind: "Syllable", Text: name}})
		if err != nil {
			return nil, err
		}
		bindings["Return"] = DefinitionArgument{Kind: "Syllable", Text: returned.Expression}
	}
	template := function.Members[member]
	instance, err := template.Instantiate(bindings)
	if err != nil {
		return nil, err
	}
	return builtinProcessExpressions(template, instance)
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
	var arguments []string
	for _, value := range current.Arguments {
		expression, err := e.value(value)
		if err != nil {
			return nil, err
		}
		argument, err := runtimeCopyExpression(expression, orchestrationAccessSnapshot)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
	}
	if current.Source.Kind == instructionReturn {
		value := "(" + strings.Join(arguments, ", ") + ")"
		if len(arguments) == 1 {
			value = "(" + arguments[0] + ",)"
		}
		return instantiateFunctionWord(e.functionWords, "Yield", map[string]string{"Values": value})
	}
	if current.Source.Kind == instructionInvoke {
		contract := e.functionContracts[current.Target.RuntimeType.Definition]
		if contract == nil {
			return nil, lerrors.Internal.Errorf("invocation has no bound Function composition")
		}
		target, err := e.value(current.Target)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("lCalled%d", e.functionCall)
		e.functionCall++
		returnContract, returned := contract, "()"
		if e.functionDepth > 0 {
			returnContract, returned = e.functionWords, "std::ops::ControlFlow::Break(())"
		}
		exit, err := instantiateFunctionWord(returnContract, "Exit", map[string]string{"Value": returned})
		if err != nil {
			return nil, err
		}
		result, err := instantiateFunctionWord(contract, "Invoke", map[string]string{
			"Target": target.text, "Name": name, "Arguments": strings.Join(append([]string{"&lContext", name + ".0.clone()"}, arguments...), ", "), "Exit": orchestrationOperation(exit),
		})
		if err != nil {
			return nil, err
		}
		for index, target := range current.Results {
			stored, err := e.storeRuntimeValue(target.Storage, fmt.Sprintf("%sResult.%d", name, index))
			if err != nil {
				return nil, err
			}
			result = append(result, stored...)
		}
		return result, nil
	}
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
	body := []string{orchestrationBinding("runtime", "&lContext.0"), orchestrationBinding("lProgram", "&lContext.1"), orchestrationBinding(function.Self.name(), "(lEnvironment.clone(), "+name+" as "+pointer+")")}
	for index, app := range e.plan.Apps {
		body = append(body, orchestrationBinding(e.apps[app], fmt.Sprintf("lProgram.apps[%d].clone()", index)))
	}
	for _, parameter := range function.Inputs {
		typ, err := nativeRuntimeType(parameter.RuntimeType)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, DefinitionProperty{Key: parameter.name(), Type: typ})
		if parameter.Mutable {
			body = append(body, orchestrationBinding("mut "+parameter.name(), parameter.name()))
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
		tail, err := instantiateFunctionWord(function.Source, "Yield", map[string]string{"Values": "()"})
		if err != nil {
			return nil, err
		}
		body = append(body, tail...)
	}
	process, err := instantiateFunctionWord(function.Source, "Suspend", map[string]string{"Body": orchestrationOperation(body)})
	if err != nil {
		return nil, err
	}
	output, err := nativeFunctionFuture(function.Source.Outputs, 0, "'_")
	if err != nil {
		return nil, err
	}
	definition, err := function.Source.Template.Instantiate(map[string]DefinitionArgument{
		"Name": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props", Props: inputs},
		"Outputs": {Kind: "Texts", Texts: []string{output}},
		"Process": {Kind: "Syllable", Text: strings.Join(process, "*")},
	})
	if err != nil {
		return nil, err
	}
	if !e.functionEmitted[function.Body.ID] {
		e.functionItems = append(e.functionItems, definition.Expression)
		e.functionEmitted[function.Body.ID] = true
	}
	created := fmt.Sprintf("lCreatedFunction%d", function.Body.ID)
	bindings, err := instantiateFunctionWord(function.Source, "Bind", map[string]string{
		"Name": created, "Function": name, "Type": pointer, "Types": captureTypes, "Captures": captures,
	})
	if err != nil {
		return nil, err
	}
	stored, err := e.storeRuntimeValue(current.Target.Storage, created)
	return append(bindings, stored...), err
}
