package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func watchProcess(pid int, stdout, stderr, stdin bool) error {
	// ensure tracing all comes from same thread
	runtime.LockOSThread()

	if _, err := os.FindProcess(pid); err != nil {
		return fmt.Errorf("could not find process with pid %d: %w", pid, err)
	}

	if err := syscall.PtraceAttach(pid); err == syscall.EPERM {
		return fmt.Errorf("could not attach to process with pid %d: %w - check your permissions", pid, err)
	} else if err != nil {
		return err
	}

	status := syscall.WaitStatus(0)
	if _, err := syscall.Wait4(pid, &status, 0, nil); err != nil {
		return err
	}

	defer func() {
		_ = syscall.PtraceDetach(pid)
		_, _ = syscall.Wait4(pid, &status, 0, nil)
	}()

	// deliver SIGTRAP|0x80
	if err := syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD); err != nil {
		return err
	}

	for {
		fd, data, err := interceptReadsAndWrites(pid)
		if err != nil {
			return err
		}

		if stdout && fd == uint64(syscall.Stdout) || stderr && fd == uint64(syscall.Stderr) || stdin && fd == uint64(syscall.Stdin) {
			if fd == uint64(syscall.Stdin) {
				fd = uint64(os.Stdin.Fd())
			}
			_, _ = fmt.Fprintf(os.NewFile(uintptr(fd), "pipe"), "%s", string(data))
		}
	}
}

func interceptReadsAndWrites(pid int) (fd uint64, data []byte, err error) {

	// intercept syscall
	err = syscall.PtraceSyscall(pid, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("could not intercept syscall: %w", err)
	}

	// wait for a syscall
	status := syscall.WaitStatus(0)
	_, err = syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		return 0, nil, err
	}

	// if interrupted, stop tracing
	if status.StopSignal().String() == "interrupt" {
		_ = syscall.PtraceSyscall(pid, int(status.StopSignal()))
		return 0, nil, fmt.Errorf("process interrupted")
	}

	defer func() {

		// continue the syscall we intercepted
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			err = fmt.Errorf("could not continue process: %w", err)
			return
		}

		// and wait for it to finish
		status := syscall.WaitStatus(0)
		_, err = syscall.Wait4(pid, &status, 0, nil)
		if err != nil {
			err = fmt.Errorf("could not wait for process: %w", err)
			return
		}

		// process exited
		if status.Exited() {
			err = fmt.Errorf("process exited")
			return
		}

		// if interrupted, stop tracing
		if status.StopSignal().String() == "interrupt" {
			_ = syscall.PtraceSyscall(pid, int(status.StopSignal()))
			err = fmt.Errorf("process interrupted")
			return
		}
	}()

	// if we have a syscall, examine it...
	if status.TrapCause()&int(syscall.SIGTRAP|0x80) > 0 {

		// read registers
		regs := &syscall.PtraceRegs{}
		if err := syscall.PtraceGetRegs(pid, regs); err != nil {
			return 0, nil, err
		}

		// find the syscall number for the host architecture
		syscallNo := grabSyscallNo(regs)

		// if it's a read/write syscall, grab the args
		switch syscallNo {
		case syscall.SYS_READ, syscall.SYS_WRITE:

			// grab the args to WRITE for the host architecture
			// fd == file descriptor (generally 1 for stdout, 2 for stderr)
			// ptr == pointer to the buffer
			// lng == length of the buffer
			fd, ptr, lng := grabArgsFromRegs(regs)

			// if we want to see this output, read it from memory
			if lng > 0 {
				data := make([]byte, lng)
				if _, err := syscall.PtracePeekData(pid, uintptr(ptr), data); err != nil {
					return 0, nil, err
				}
				return fd, data, nil
			}
		}

	}

	return 0, nil, err
}
