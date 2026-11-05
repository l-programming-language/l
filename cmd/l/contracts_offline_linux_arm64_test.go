package main

import "syscall"

const contractsAuditArch uint32 = 0xc00000b7
const contractsSeccompSyscall uintptr = syscall.SYS_SECCOMP

var contractsDeniedSyscalls = []uint32{syscall.SYS_SOCKET, syscall.SYS_SOCKETPAIR, syscall.SYS_CONNECT, syscall.SYS_SENDTO, syscall.SYS_SENDMSG, syscall.SYS_EXECVE, syscall.SYS_EXECVEAT}
