//go:build 386

package main

import (
	"syscall"
)

func grabSyscallNo(regs *syscall.PtraceRegs) uint64 {
	return uint64(regs.Orig_eax)
}

func grabArgsFromRegs(regs *syscall.PtraceRegs) (fd, ptr, lng uint64) {
	return uint64(regs.Ebx), uint64(regs.Ecx), uint64(regs.Edx)
}
