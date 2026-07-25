package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "devsy-provider-lima",
		Short:         "Lima Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}

func initLogger() {
	cfg := log.Config{Verbosity: 1}
	if os.Getenv("DEVSY_DEBUG") == "true" {
		cfg.Debug = true
	}
	log.Init(cfg)
}

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

// Execute propagates the inner process's exit code so devsy sees it verbatim.
func Execute() {
	initLogger()
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
