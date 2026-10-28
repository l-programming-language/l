package Lyger

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Process implementations are selected by semantic operation, never by a
// competing lexical definition of Run or by inspecting native Rust text.
type builtinProcessOperation uint8

const (
	builtinProcessBuild builtinProcessOperation = iota
	builtinProcessRun
	builtinProcessStop
	builtinProcessAwait
	builtinProcessFinish
	builtinProcessCount
)

var builtinProcessSignatures = [builtinProcessCount]struct {
	name  string
	slots []string
}{
	builtinProcessBuild:  {"build", []string{"Receiver"}},
	builtinProcessRun:    {"run", []string{"Receiver"}},
	builtinProcessStop:   {"stop", []string{"Receiver"}},
	builtinProcessAwait:  {"await", []string{"Receiver", "Invocation", "Body"}},
	builtinProcessFinish: {"finish-continuation", []string{"Invocation"}},
}

// A catalog belongs to one Compilation and therefore one selected filesystem.
// Parsed prototypes are immutable; every emission instantiates fresh bytes.
type builtinProcessCatalog struct {
	definitions [builtinProcessCount]*DefinitionTemplate
}

func loadBuiltinProcesses(resources fs.FS) (*builtinProcessCatalog, error) {
	catalog := &builtinProcessCatalog{}
	for operation, signature := range builtinProcessSignatures {
		name := "builtins/operations/rust/process-" + signature.name + ".def"
		raw, err := fs.ReadFile(resources, name)
		if err != nil {
			return nil, lerrors.IO.Errorf("read builtin process implementation %s: %w", name, err)
		}
		if _, framed := splitDefinitionFrame(strings.TrimSpace(string(raw))); !framed {
			return nil, lerrors.InvalidConfiguration.Errorf("builtin process implementation %s requires paired %s markers", name, definitionMarker)
		}
		template, err := ParseDefinitionTemplate(raw, SourceSpan{Path: name})
		if err != nil {
			return nil, err
		}
		if err := validateBuiltinProcessTemplate(template, builtinProcessOperation(operation)); err != nil {
			return nil, err
		}
		catalog.definitions[operation] = template
	}
	return catalog, nil
}

func validateBuiltinProcessTemplate(template *DefinitionTemplate, operation builtinProcessOperation) error {
	signature := builtinProcessSignatures[operation].slots
	wanted := make(map[string]bool, len(signature))
	for _, name := range signature {
		wanted[name] = true
	}
	for _, slot := range template.Slots {
		if !wanted[slot.Name] {
			return template.failure(lerrors.InvalidConfiguration.Errorf("unknown builtin parameter %q", slot.Name))
		}
		if slot.Type != "" {
			return template.failure(lerrors.InvalidConfiguration.Errorf("builtin parameter %q has a private compiler type and cannot declare %q", slot.Name, slot.Type))
		}
		delete(wanted, slot.Name)
	}
	for _, name := range signature {
		if wanted[name] {
			return template.failure(lerrors.InvalidConfiguration.Errorf("missing builtin parameter %q", name))
		}
	}
	// Validate composition eagerly, including unused implementations in a
	// complete distribution. These inert holes are never emitted or evaluated.
	bindings := map[string]DefinitionArgument{
		"Receiver":   {Kind: "Syllable", Text: "receiver"},
		"Invocation": {Kind: "Syllable", Text: "invocation"},
		"Body":       {Kind: "Syllable", Text: "W"},
	}
	instance, err := template.Instantiate(bindings)
	if err != nil {
		return err
	}
	expressions, err := builtinProcessExpressions(template, instance)
	if err != nil {
		return err
	}
	want := 1
	if operation == builtinProcessAwait {
		want = 2
	}
	if len(expressions) != want {
		return template.failure(lerrors.InvalidConfiguration.Errorf("builtin %s requires %d ordered expressions", builtinProcessSignatures[operation].name, want))
	}
	if operation == builtinProcessAwait {
		return validateBuiltinAwaitSlots(template)
	}
	return nil
}

func validateBuiltinAwaitSlots(template *DefinitionTemplate) error {
	// The first expression owns the process join and invocation binding; the
	// second consumes that binding and the planned continuation. Validate these
	// private slot roles without parsing or prescribing the native Rust leaves.
	for index, expected := range [][]string{{"Receiver", "Invocation"}, {"Invocation", "Body"}} {
		counts := make(map[string]int)
		for _, name := range expected {
			counts[name] = 0
		}
		for _, part := range template.words[0][index] {
			if part.slot == nil {
				continue
			}
			name := part.slot.Name
			if _, allowed := counts[name]; !allowed {
				return template.failure(lerrors.InvalidConfiguration.Errorf("builtin await parameter %s is invalid in expression %d", name, index+1))
			}
			counts[name]++
		}
		for _, name := range expected {
			if counts[name] == 0 {
				return template.failure(lerrors.InvalidConfiguration.Errorf("builtin await expression %d requires parameter %s", index+1, name))
			}
		}
		if index == 1 && counts["Body"] != 1 {
			return template.failure(lerrors.InvalidConfiguration.Errorf("builtin await Body must occur exactly once in the guarded continuation"))
		}
	}
	return nil
}

// These operands are private compiler values. They cannot be supplied through
// DefinitionArgument, lexical source values, or a public Native argument kind.
type builtinProcessReceiver struct {
	value *orchestrationValue
	name  string
}

type builtinProcessInvocation struct{ index int }

func (invocation builtinProcessInvocation) name() string {
	return fmt.Sprintf("lInvocation%d", invocation.index)
}

type builtinProcessBody struct{ definition DefinitionInstance }

type builtinProcessOperands struct {
	receiver   *builtinProcessReceiver
	invocation *builtinProcessInvocation
	body       *builtinProcessBody
}

func (catalog *builtinProcessCatalog) instantiate(operation builtinProcessOperation, operands builtinProcessOperands) ([]string, error) {
	if catalog == nil || operation >= builtinProcessCount || catalog.definitions[operation] == nil {
		return nil, lerrors.Internal.Errorf("missing selected builtin process implementation")
	}
	template := catalog.definitions[operation]
	bindings := make(map[string]DefinitionArgument)
	for _, slot := range builtinProcessSignatures[operation].slots {
		var expression string
		switch slot {
		case "Receiver":
			receiver := operands.receiver
			if receiver == nil || receiver.value == nil || receiver.value.Type != valueApp || receiver.value.App == nil || !validDefinitionNativeName(receiver.name) {
				return nil, template.failure(lerrors.Internal.Errorf("Receiver requires a planned App handle"))
			}
			expression = receiver.name
		case "Invocation":
			if operands.invocation == nil || operands.invocation.index < 0 {
				return nil, template.failure(lerrors.Internal.Errorf("Invocation requires a compiler-owned invocation binding"))
			}
			expression = operands.invocation.name()
		case "Body":
			if operands.body == nil {
				return nil, template.failure(lerrors.Internal.Errorf("Body requires a composed continuation"))
			}
			bindings[slot] = DefinitionArgument{Kind: "Definition", Definition: &operands.body.definition}
			continue
		}
		// The generic renderer only receives checked compiler operands after
		// the private signature has enforced their distinct roles.
		bindings[slot] = DefinitionArgument{Kind: "Syllable", Text: expression}
	}
	instance, err := template.Instantiate(bindings)
	if err != nil {
		return nil, err
	}
	return builtinProcessExpressions(template, instance)
}

func builtinProcessExpressions(template *DefinitionTemplate, instance DefinitionInstance) ([]string, error) {
	if len(instance.before) > 0 {
		var expressions []string
		for _, step := range definitionSteps(instance) {
			selected, err := builtinProcessExpressions(template, step)
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, selected...)
		}
		return expressions, nil
	}
	if instance.sequence != nil {
		var expressions []string
		for _, child := range instance.sequence {
			selected, err := builtinProcessExpressions(template, child)
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, selected...)
		}
		return expressions, nil
	}
	if instance.Words == nil {
		return []string{instance.Expression}, nil
	}
	if len(instance.Words) != 1 {
		return nil, template.failure(lerrors.InvalidConfiguration.Errorf("builtin process composition requires one ordered word"))
	}
	return append([]string(nil), instance.Words[0]...), nil
}

func definitionSteps(instance DefinitionInstance) []DefinitionInstance {
	var steps []DefinitionInstance
	for _, prior := range instance.before {
		steps = append(steps, definitionSteps(prior)...)
	}
	instance.before = nil
	return append(steps, instance)
}
