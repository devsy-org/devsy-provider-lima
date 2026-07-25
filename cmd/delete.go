package cmd

import (
	"github.com/devsy-org/devsy-provider-lima/pkg/lima"
	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/spf13/cobra"
)

// NewDeleteCmd deletes the Lima VM for a workspace machine.
func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a Lima VM",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return lima.NewClient(opts.LimactlPath).Delete(cmd.Context(), opts.MachineID)
		},
	}
}
