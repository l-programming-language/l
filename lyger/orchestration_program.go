package Lyger

import (
	"fmt"
	"strings"
)

// Functions borrow root state for published declarations and continuations.
// Explicit Function captures remain independent snapshots of these values.
func (e *orchestrationEmitter) programContext(actions []*record) ([]string, error) {
	e.programStorage = make(map[*orchestrationStorage]int)
	e.programActions = make(map[*record]int)
	var types, apps []string
	if e.plan.FunctionProgramState {
		for index, slot := range e.plan.Root.Storage {
			e.programStorage[slot] = index
			typ, err := runtimeTypeExpression(slot.RuntimeType, e.plan.Identities)
			if err != nil {
				return nil, err
			}
			types = append(types, typ)
		}
		for index, record := range actions {
			e.programActions[record] = index
		}
	}
	for _, app := range e.plan.Apps {
		apps = append(apps, e.apps[app]+".clone()")
	}
	return []string{orchestrationStatement(fmt.Sprintf("let lProgram = &LProgram::new(vec![%s], vec![%s], %d);", strings.Join(apps, ", "), strings.Join(types, ", "), len(e.programActions)))}, nil
}
