package native

import "time"

// BuildPrerequisiteTimeout bounds compilation performed by acceptance helpers.
// It does not apply to application execution, readiness, or cleanup watchdogs.
const BuildPrerequisiteTimeout = 40 * time.Second

// CommandPhase classifies a supervised command at its call site, independently
// of its executable name or arguments.
type CommandPhase string

const (
	BuildPrerequisite CommandPhase = "build prerequisite; runtime not exercised"
	RuntimeExecution  CommandPhase = "runtime"
)

// Timeout preserves the caller's runtime contract while giving compilation its
// separate budget. An unknown phase cannot silently receive a longer deadline.
func (phase CommandPhase) Timeout(runtimeTimeout time.Duration) time.Duration {
	switch phase {
	case BuildPrerequisite:
		return BuildPrerequisiteTimeout
	case RuntimeExecution:
		if runtimeTimeout <= 0 {
			panic("runtime command requires a positive watchdog")
		}
		return runtimeTimeout
	default:
		panic("unknown supervised command phase")
	}
}
