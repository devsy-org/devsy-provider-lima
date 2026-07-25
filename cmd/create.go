package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-lima/pkg/lima"
	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a Lima VM",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreate(cmd.Context())
		},
	}
}

func runCreate(ctx context.Context) error {
	opts, err := options.FromEnv(true)
	if err != nil {
		return err
	}

	template, err := opts.ResolveTemplate(opts.MachineFolder)
	if err != nil {
		return err
	}

	client := lima.NewClient(opts.LimactlPath)
	log.Infof(
		"creating Lima VM %q (runtime=%s, template=%s, "+
			"cpus=%s, memory=%sGiB, disk=%sGiB, vmType=%s)",
		opts.MachineID, opts.Runtime.Name(), template,
		opts.CPUs, opts.MemoryGiB, opts.DiskGiB, opts.VMType,
	)

	return client.Create(ctx, lima.CreateParams{
		Name:     opts.MachineID,
		Template: template,
		CPUs:     opts.CPUs,
		Memory:   opts.MemoryGiB,
		Disk:     opts.DiskGiB,
		VMType:   opts.VMType,
	})
}
