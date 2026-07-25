package cmd

import (
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-lima/pkg/lima"
	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get the status of a Lima VM",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}

			status, err := lima.NewClient(opts.LimactlPath).Status(cmd.Context(), opts.MachineID)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(os.Stdout, status)
			return err
		},
	}
}
