package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/devsy-org/devsy-provider-lima/pkg/options"
	"github.com/spf13/cobra"
)

// NewInitCmd validates that the provider is usable: options parse and limactl
// is installed and runnable.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Validate the Lima provider is ready",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(false)
			if err != nil {
				return err
			}
			return checkLimactl(cmd.Context(), opts.LimactlPath)
		},
	}
}

func checkLimactl(ctx context.Context, path string) error {
	//nolint:gosec // fixed limactl binary from provider options.
	out, err := exec.CommandContext(ctx, path, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"limactl not found or not runnable at %q: %w: %s (install Lima: https://lima-vm.io)",
			path, err, strings.TrimSpace(string(out)),
		)
	}
	return nil
}
