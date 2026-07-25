// Package lima is a thin wrapper around the limactl CLI. The provider shells out
// to limactl rather than importing Lima as a library, which would pull an
// enormous vz/qemu dependency tree.
package lima

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devsy-org/devsy-provider-lima/pkg/log"
)

// Status values devsy expects on stdout from the status command.
const (
	StatusRunning  = "Running"
	StatusStopped  = "Stopped"
	StatusBusy     = "Busy"
	StatusNotFound = "NotFound"
)

// flagNoTTY disables limactl's interactive prompts for automation.
const flagNoTTY = "--tty=false"

const verbCreate = "create"

// Client drives a Lima instance through the limactl binary.
type Client struct {
	path string
}

// NewClient returns a Client using the given limactl binary path.
func NewClient(limactlPath string) *Client {
	return &Client{path: limactlPath}
}

// CreateParams configure a new Lima instance.
type CreateParams struct {
	Name     string
	Template string
	CPUs     string
	Memory   string // GiB
	Disk     string // GiB
	VMType   string
}

// createArgs builds the `limactl create` argument list; the template is last.
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

// Create provisions and starts an instance. An existing instance is reused
// as-is: its size/template are not reconciled with p.
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

// Status returns the devsy status for an instance.
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

// mapStatus maps a Lima instance status to a devsy status.
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

// Shell runs a command inside the instance with stdio wired through, returning
// limactl's exit error verbatim so the caller can propagate the exit code.
func (c *Client) Shell(ctx context.Context, name, command string) error {
	//nolint:gosec // fixed limactl binary; command is provided by devsy.
	cmd := exec.CommandContext(ctx, c.path, "shell", "--workdir", "/", name, "sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// exists reports whether the instance is present.
func (c *Client) exists(ctx context.Context, name string) (bool, error) {
	_, found, err := c.lookup(ctx, name)
	return found, err
}

// lookup finds an instance in `limactl list` output. A missing instance is
// found=false with no error; a limactl failure is returned as an error so devsy
// does not mistake a broken limactl for a missing VM.
func (c *Client) lookup(ctx context.Context, name string) (status string, found bool, err error) {
	out, err := c.output(ctx, "list", "--format", "{{.Name}}\t{{.Status}}")
	if err != nil {
		return "", false, fmt.Errorf("list instances: %w", withStderr(err))
	}
	status, found = parseInstanceStatus(out, name)
	return status, found, nil
}

// parseInstanceStatus finds name in tab-separated "name\tstatus" list output.
func parseInstanceStatus(listOutput, name string) (status string, found bool) {
	for _, line := range strings.Split(listOutput, "\n") {
		n, s, ok := strings.Cut(line, "\t")
		if ok && strings.TrimSpace(n) == name {
			return strings.TrimSpace(s), true
		}
	}
	return "", false
}

func (c *Client) run(ctx context.Context, args ...string) error {
	log.Debugf("running: %s %s", c.path, strings.Join(args, " "))
	//nolint:gosec // fixed limactl binary; internal args.
	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Stdout = os.Stderr // limactl progress must not corrupt command stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) output(ctx context.Context, args ...string) (string, error) {
	//nolint:gosec // fixed limactl binary; internal args.
	cmd := exec.CommandContext(ctx, c.path, args...)
	out, err := cmd.Output()
	return string(out), err
}

// withStderr enriches an *exec.ExitError with the captured stderr.
func withStderr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}
