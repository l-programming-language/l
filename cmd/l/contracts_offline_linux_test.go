package main

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

func contractsOfflineCommand(ctx context.Context, executable string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, executable, "-test.run=^TestContractsOfflineCapabilitiesHelper$", "-test.v"), nil
}

func contractsDenyProcessAndNetwork() error {
	const loadWord, jumpEqual, jumpSet, bitAnd, ret = 0x20, 0x15, 0x45, 0x54, 0x06
	const allow, errno = 0x7fff0000, 0x00050000
	denied := uint32(errno | syscall.EACCES)
	filters := []syscall.SockFilter{
		{Code: loadWord, K: 4}, // seccomp_data.arch
		{Code: jumpEqual, K: contractsAuditArch, Jt: 1},
		{Code: ret, K: denied},
		{Code: loadWord, K: 0}, // seccomp_data.nr
		// Reject the x32 ABI bit rather than letting alternate syscall numbers
		// evade the x86-64 deny list. It is not a supported ABI on either host.
		{Code: jumpSet, K: 0x40000000, Jf: 1},
		{Code: ret, K: denied},
	}
	for _, number := range contractsDeniedSyscalls {
		filters = append(filters, syscall.SockFilter{Code: jumpEqual, K: number, Jf: 1}, syscall.SockFilter{Code: ret, K: denied})
	}
	// clone3's indirect flags cannot be inspected by classic BPF. ENOSYS lets
	// pthread implementations fall back to clone for permitted OS threads.
	filters = append(filters,
		syscall.SockFilter{Code: jumpEqual, K: 435, Jf: 1},
		syscall.SockFilter{Code: ret, K: errno | uint32(syscall.ENOSYS)},
		syscall.SockFilter{Code: jumpEqual, K: uint32(syscall.SYS_CLONE), Jf: 4},
		syscall.SockFilter{Code: loadWord, K: 16}, // args[0], clone flags
		syscall.SockFilter{Code: bitAnd, K: uint32(syscall.CLONE_THREAD)},
		syscall.SockFilter{Code: jumpEqual, K: uint32(syscall.CLONE_THREAD), Jt: 1},
		syscall.SockFilter{Code: ret, K: denied},
		syscall.SockFilter{Code: ret, K: allow},
	)
	program := syscall.SockFprog{Len: uint16(len(filters)), Filter: &filters[0]}
	if _, _, err := syscall.RawSyscall6(syscall.SYS_PRCTL, 38, 1, 0, 0, 0, 0); err != 0 {
		return fmt.Errorf("set no-new-privileges: %w", err)
	}
	// TSYNC applies the filter to every existing Go runtime thread, not merely
	// the installing goroutine's current OS thread.
	result, _, err := syscall.RawSyscall(contractsSeccompSyscall, 1, 1, uintptr(unsafe.Pointer(&program)))
	if err != 0 {
		return fmt.Errorf("install synchronized seccomp filter: %w", err)
	}
	if result != 0 {
		return fmt.Errorf("seccomp synchronization rejected thread %d", result)
	}
	return nil
}
