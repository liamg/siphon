//go:build arm

package main

import (
	"syscall"
)

func grabSyscallNo(regs *syscall.PtraceRegs) uint64 {
	return uint64(regs.Uregs[7])
}

func grabArgsFromRegs(regs *syscall.PtraceRegs) (fd, ptr, lng uint64) {
	return uint64(regs.Uregs[0]), uint64(regs.Uregs[1]), uint64(regs.Uregs[2])
}
