//go:build arm64

package main

import (
	"syscall"
)

func grabSyscallNo(regs *syscall.PtraceRegs) uint64 {
	return regs.Regs[8]
}

func grabArgsFromRegs(regs *syscall.PtraceRegs) (fd, ptr, lng uint64) {
	return regs.Regs[0], regs.Regs[1], regs.Regs[2]
}
