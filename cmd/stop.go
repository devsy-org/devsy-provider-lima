package cmd

import (
	"github.com/devsy-org/devsy-provider-lima/pkg/lima"
	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/spf13/cobra"
)

// NewStopCmd stops a running Lima VM.
func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a Lima VM",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return lima.NewClient(opts.LimactlPath).Stop(cmd.Context(), opts.MachineID)
		},
	}
}
