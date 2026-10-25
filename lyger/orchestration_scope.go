package Lyger

import "github.com/l-programming-language/l/lerrors"

// Each Rust activation owns distinct scalar slots while retaining the original
// lexical scope, caller facts and published state. No scope creates a task.
func (p *orchestrationPlanner) scoped(block *scopedInstructions, flow orchestrationFlow) (*orchestrationBlock, orchestrationFlow, error) {
	if block == nil || block.Scope == nil {
		return nil, flow, lerrors.Internal.Errorf("instruction block requires lexical scope")
	}
	result := p.newBlock(block.Scope)
	p.frames = append(p.frames, &orchestrationFrame{block: result, slots: make(map[*record]*orchestrationStorage)})
	defer func() { p.frames = p.frames[:len(p.frames)-1] }()
	// New private slots have no facts even when another activation shares the
	// same lexical scope. The suspended caller's slots remain untouched.
	body, next, err := p.instructions(block.Body, flow.clone())
	if err != nil {
		return nil, flow, err
	}
	result.Body = body
	for _, slot := range result.Storage {
		delete(next.assigned, slot)
		delete(next.written, slot)
	}
	return result, next, nil
}
