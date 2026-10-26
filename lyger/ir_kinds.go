package Lyger

import (
	"strconv"

	"github.com/l-programming-language/l/lerrors"
)

// These tags describe closed compiler operations, never configured vocabulary.
type instructionKind uint8

const (
	instructionInvalid instructionKind = iota
	instructionAssign
	instructionCall
	instructionBuild
	instructionRun
	instructionShow
	instructionExit
	instructionIf
	instructionBlock
	instructionAwait
	instructionStop
	instructionBindAction
	instructionAliasAction
	instructionFunction
	instructionInvoke
	instructionReturn
	instructionUpdate
	instructionKindCount
)

var instructionKindNames = [instructionKindCount]string{
	instructionAssign: "assign", instructionCall: "call", instructionBuild: "build",
	instructionRun: "run", instructionShow: "show", instructionExit: "exit", instructionIf: "if", instructionBlock: "block",
	instructionAwait: "await", instructionStop: "stop",
	instructionBindAction: "bind-action", instructionAliasAction: "alias-action",
	instructionUpdate:   "update",
	instructionFunction: "function", instructionInvoke: "invoke", instructionReturn: "return",
}

func (k instructionKind) Valid() bool { return k > instructionInvalid && k < instructionKindCount }

func (k instructionKind) String() string {
	if k >= instructionKindCount {
		return "instructionKind(" + strconv.Itoa(int(k)) + ")"
	}
	return instructionKindNames[k]
}

func (k instructionKind) MarshalText() ([]byte, error) {
	if k >= instructionKindCount {
		return nil, lerrors.Internal.Errorf("invalid instruction kind %d", uint8(k))
	}
	return []byte(instructionKindNames[k]), nil
}

type operandKind uint8

const (
	operandInvalid operandKind = iota
	operandSymbol
	operandLiteral
	operandMember
	operandError
	operandConstruct
	operandRead
	operandKindCount
)

var operandKindNames = [operandKindCount]string{
	operandSymbol: "symbol", operandLiteral: "literal", operandMember: "member", operandError: "error", operandConstruct: "construct", operandRead: "read",
}

func (k operandKind) String() string {
	if k >= operandKindCount {
		return "operandKind(" + strconv.Itoa(int(k)) + ")"
	}
	return operandKindNames[k]
}

func (k operandKind) MarshalText() ([]byte, error) {
	if k >= operandKindCount {
		return nil, lerrors.Internal.Errorf("invalid operand kind %d", uint8(k))
	}
	return []byte(operandKindNames[k]), nil
}

type nativeMember uint8

const (
	memberInvalid nativeMember = iota
	memberOutput
	memberError
	nativeMemberCount
)

var nativeMembers = [nativeMemberCount]struct {
	name   string
	result valueType
}{
	memberOutput: {"output", valueText},
	memberError:  {"error", valueError},
}

func (m nativeMember) Valid() bool { return m > memberInvalid && m < nativeMemberCount }

func (m nativeMember) String() string {
	if m >= nativeMemberCount {
		return "nativeMember(" + strconv.Itoa(int(m)) + ")"
	}
	return nativeMembers[m].name
}

func (m nativeMember) MarshalText() ([]byte, error) {
	if m >= nativeMemberCount {
		return nil, lerrors.Internal.Errorf("invalid native member %d", uint8(m))
	}
	return []byte(nativeMembers[m].name), nil
}

func (m nativeMember) ValueType() (valueType, bool) {
	if !m.Valid() {
		return valueInvalid, false
	}
	return nativeMembers[m].result, true
}

type nativeRole uint8

const (
	roleInvalid nativeRole = iota
	roleGenerate
	roleExecute
	roleObserve
	roleTerminate
	roleDefine
	roleOutput
	roleError
	roleSuspend
	roleInterrupt
	roleDeclare
	rolePut
	rolePlace
	nativeRoleCount
)

var nativeRoles = [nativeRoleCount]struct {
	name        string
	instruction instructionKind
	member      nativeMember
}{
	roleGenerate:  {"action.generative", instructionBuild, memberInvalid},
	roleExecute:   {"action.executive", instructionRun, memberInvalid},
	roleObserve:   {"action.observative", instructionShow, memberInvalid},
	roleTerminate: {"action.terminative", instructionExit, memberInvalid},
	roleDefine:    {"action.definitive", instructionAssign, memberInvalid},
	roleOutput:    {"functional.deflective", instructionInvalid, memberOutput},
	roleError:     {"functional.error", instructionInvalid, memberError},
	roleSuspend:   {"action.suspensive", instructionAwait, memberInvalid},
	roleInterrupt: {"action.interruptive", instructionStop, memberInvalid},
	roleDeclare:   {"action.declarative", instructionAssign, memberInvalid},
	rolePut:       {"action.expansive", instructionUpdate, memberInvalid}, rolePlace: {"action.locative", instructionUpdate, memberInvalid},
}

func (r nativeRole) String() string {
	if r >= nativeRoleCount {
		return "nativeRole(" + strconv.Itoa(int(r)) + ")"
	}
	return nativeRoles[r].name
}

func parseNativeRole(raw string) (nativeRole, bool) {
	for r := roleInvalid + 1; r < nativeRoleCount; r++ {
		if nativeRoles[r].name == raw {
			return r, true
		}
	}
	return roleInvalid, false
}

func (r nativeRole) InstructionKind() (instructionKind, bool) {
	if r == roleInvalid || r >= nativeRoleCount {
		return instructionInvalid, false
	}
	kind := nativeRoles[r].instruction
	return kind, kind.Valid()
}

func (r nativeRole) Member() (nativeMember, bool) {
	if r == roleInvalid || r >= nativeRoleCount {
		return memberInvalid, false
	}
	member := nativeRoles[r].member
	return member, member.Valid()
}
