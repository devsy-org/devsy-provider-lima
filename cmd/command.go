package cmd

import (
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-lima/pkg/lima"
	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/spf13/cobra"
)

// NewCommandCmd runs a command inside the Lima VM. This is the exec channel
// devsy uses to inject its agent and tunnel all further communication; stdio is
// wired straight through.
func NewCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "command",
		Short: "Run a command inside a Lima VM",
		RunE: func(cmd *cobra.Command, _ []string) error {
			command := os.Getenv("COMMAND")
			if command == "" {
				return fmt.Errorf("COMMAND environment variable is missing")
			}

			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}

			return lima.NewClient(opts.LimactlPath).Shell(cmd.Context(), opts.MachineID, command)
		},
	}
}
