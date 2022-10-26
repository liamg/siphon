package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:  "siphon [pid]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		pid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid pid: %w", err)
		}
		return watchProcess(int(pid), flagStdOut, flagStdErr, flagStdIn)
	},
}

var (
	flagStdOut = true
	flagStdErr = true
	flagStdIn  = false
)

func main() {

	cmd.Flags().BoolVarP(&flagStdOut, "stdout", "o", flagStdOut, "Show stdout")
	cmd.Flags().BoolVarP(&flagStdErr, "stderr", "e", flagStdErr, "Show stderr")
	cmd.Flags().BoolVarP(&flagStdIn, "stdin", "i", flagStdIn, "Show stdin")

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
	}
}
