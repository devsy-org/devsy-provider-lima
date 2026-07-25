// Package lima wraps the limactl CLI.
package lima

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devsy-org/devsy/pkg/log"
)

const (
	StatusRunning  = "Running"
	StatusStopped  = "Stopped"
	StatusBusy     = "Busy"
	StatusNotFound = "NotFound"
)

const (
	flagNoTTY  = "--tty=false"
	verbCreate = "create"
)

type Client struct {
	path string
}

func NewClient(limactlPath string) *Client {
	return &Client{path: limactlPath}
}

type CreateParams struct {
	Name     string
	Template string
	CPUs     string
	Memory   string
	Disk     string
	VMType   string
}

func createArgs(p CreateParams) []string {
	args := []string{
		verbCreate, flagNoTTY, "--name=" + p.Name,
		"--cpus=" + p.CPUs,
		"--memory=" + p.Memory,
		"--disk=" + p.Disk,
	}
	if p.VMType != "" {
		args = append(args, "--vm-type="+p.VMType)
	}
	return append(args, p.Template)
}

// Create starts the instance, reusing an existing one as-is.
func (c *Client) Create(ctx context.Context, p CreateParams) error {
	exists, err := c.exists(ctx, p.Name)
	if err != nil {
		return err
	}

	if exists {
		log.Infof("instance %q already exists; reusing it "+
			"(requested cpus/memory/disk are not applied to an existing VM)", p.Name)
	} else if err := c.run(ctx, createArgs(p)...); err != nil {
		return fmt.Errorf("create instance %q: %w", p.Name, err)
	}

	return c.Start(ctx, p.Name)
}

func (c *Client) Start(ctx context.Context, name string) error {
	if err := c.run(ctx, "start", flagNoTTY, name); err != nil {
		return fmt.Errorf("start instance %q: %w", name, err)
	}
	return nil
}

func (c *Client) Stop(ctx context.Context, name string) error {
	if err := c.run(ctx, "stop", name); err != nil {
		return fmt.Errorf("stop instance %q: %w", name, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, name string) error {
	if err := c.run(ctx, "delete", "--force", name); err != nil {
		return fmt.Errorf("delete instance %q: %w", name, err)
	}
	return nil
}

func (c *Client) Status(ctx context.Context, name string) (string, error) {
	raw, found, err := c.lookup(ctx, name)
	if err != nil {
		return "", err
	}
	if !found {
		return StatusNotFound, nil
	}
	status, err := mapStatus(raw)
	if err != nil {
		return "", fmt.Errorf("instance %q: %w", name, err)
	}
	return status, nil
}

func mapStatus(limaStatus string) (string, error) {
	switch strings.TrimSpace(limaStatus) {
	case StatusRunning:
		return StatusRunning, nil
	case StatusStopped:
		return StatusStopped, nil
	case "Broken":
		return "", errors.New("instance is in a broken state")
	default:
		return StatusBusy, nil
	}
}

// Shell wires stdio through so the caller can propagate the exit code.
func (c *Client) Shell(ctx context.Context, name, command string) error {
	//nolint:gosec // fixed binary; command comes from devsy.
	cmd := exec.CommandContext(ctx, c.path, "shell", "--workdir", "/", name, "sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) exists(ctx context.Context, name string) (bool, error) {
	_, found, err := c.lookup(ctx, name)
	return found, err
}

// lookup returns found=false for a missing instance, but an error for a limactl
// failure, so a broken limactl is not mistaken for a missing VM.
func (c *Client) lookup(ctx context.Context, name string) (status string, found bool, err error) {
	out, err := c.output(ctx, "list", "--format", "{{.Name}}\t{{.Status}}")
	if err != nil {
		return "", false, fmt.Errorf("list instances: %w", withStderr(err))
	}
	status, found = parseInstanceStatus(out, name)
	return status, found, nil
}

func parseInstanceStatus(listOutput, name string) (status string, found bool) {
	for line := range strings.SplitSeq(listOutput, "\n") {
		n, s, ok := strings.Cut(line, "\t")
		if ok && strings.TrimSpace(n) == name {
			return strings.TrimSpace(s), true
		}
	}
	return "", false
}

func (c *Client) run(ctx context.Context, args ...string) error {
	log.Debugf("running: %s %s", c.path, strings.Join(args, " "))
	//nolint:gosec // fixed binary; internal args.
	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Stdout = os.Stderr // keep limactl progress off command stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) output(ctx context.Context, args ...string) (string, error) {
	//nolint:gosec // fixed binary; internal args.
	cmd := exec.CommandContext(ctx, c.path, args...)
	out, err := cmd.Output()
	return string(out), err
}

func withStderr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}
