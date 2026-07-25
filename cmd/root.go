package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy-provider-lima/pkg/log"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the provider's root command.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "devsy-provider-lima",
		Short:         "Lima Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}

// BuildRoot assembles the root command with all provider verbs.
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewStatusCmd())
	return rootCmd
}

// Execute runs the root command. It propagates the exit code of an inner
// process (e.g. the command run inside the VM) so devsy sees it verbatim.
func Execute() {
	if err := BuildRoot().Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if len(exitErr.Stderr) > 0 {
				log.Errorf("%s", string(exitErr.Stderr))
			}
			os.Exit(exitErr.ExitCode())
		}
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
