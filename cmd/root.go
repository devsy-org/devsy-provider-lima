package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/logging"
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

func initLogger() error {
	cfg, invalid := logging.ResolveProvider(
		os.Getenv("DEVSY_LOG_LEVEL"), os.Getenv("DEVSY_LOG_FORMAT"),
	)
	if invalid.Level {
		return errors.New("invalid DEVSY_LOG_LEVEL: expected error, warn, info, debug, or trace")
	}
	if invalid.Format {
		return errors.New("invalid DEVSY_LOG_FORMAT: expected auto, json, text, plain, or logfmt")
	}
	log.Init(
		log.Config{
			Level:        logging.LevelName(cfg.Level),
			DefaultLevel: log.LevelInfoName,
			Format:       string(cfg.Format),
		},
	)
	return nil
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
	if err := initLogger(); err != nil {
		log.Init(log.Config{DefaultLevel: log.LevelInfoName, Format: "json"})
		log.Errorf("%v", err)
		_ = log.Sync()
		os.Exit(1)
	}
	if err := BuildRoot().Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if len(exitErr.Stderr) > 0 {
				log.Errorf("%s", string(exitErr.Stderr))
			}
			_ = log.Sync()
			os.Exit(exitErr.ExitCode())
		}
		log.Errorf("%v", err)
		_ = log.Sync()
		os.Exit(1)
	}
	_ = log.Sync()
}
