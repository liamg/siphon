//go:build amd64

package main

import (
	"syscall"
)

func grabSyscallNo(regs *syscall.PtraceRegs) uint64 {
	return regs.Orig_rax
}

func grabArgsFromRegs(regs *syscall.PtraceRegs) (fd, ptr, lng uint64) {
	return regs.Rdi, regs.Rsi, regs.Rdx
}
