package native

import (
	"context"
	"time"
)

// CommandPhase classifies a supervised command at its call site, independently
// of its executable name or arguments.
type CommandPhase string

const (
	BuildPrerequisite CommandPhase = "build prerequisite; runtime not exercised"
	RuntimeExecution  CommandPhase = "runtime"
)

// Context preserves runtime watchdogs without imposing a short compilation
// deadline. Both phases retain parent cancellation and owned-process cleanup.
func (phase CommandPhase) Context(parent context.Context, runtimeTimeout time.Duration) (context.Context, context.CancelFunc) {
	switch phase {
	case BuildPrerequisite:
		return context.WithCancel(parent)
	case RuntimeExecution:
		if runtimeTimeout <= 0 {
			panic("runtime command requires a positive watchdog")
		}
		return context.WithTimeout(parent, runtimeTimeout)
	default:
		panic("unknown supervised command phase")
	}
}
