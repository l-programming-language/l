package main

import "syscall"

const contractsAuditArch uint32 = 0xc000003e
const contractsSeccompSyscall uintptr = 317

// execveat and seccomp are absent from Go's frozen linux/amd64 syscall table.
const contractsExecveat = 322

var contractsDeniedSyscalls = []uint32{syscall.SYS_SOCKET, syscall.SYS_CONNECT, syscall.SYS_SENDTO, syscall.SYS_SENDMSG, syscall.SYS_SOCKETPAIR, syscall.SYS_FORK, syscall.SYS_VFORK, syscall.SYS_EXECVE, contractsExecveat}
