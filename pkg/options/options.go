// Package options reads the provider configuration devsy passes as environment
// variables when it invokes the provider binary.
package options

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/devsy-org/devsy-provider-lima/pkg/runtime"
)

// Options holds resolved provider configuration for a single invocation.
type Options struct {
	// MachineID is the devsy machine identifier; used as the Lima instance name.
	MachineID string
	// MachineFolder is the local devsy folder for this machine.
	MachineFolder string

	// Runtime is the selected container runtime (docker|podman|nerdctl).
	Runtime runtime.Runtime

	LimactlPath string
	// Template optionally overrides the Lima template locator. When empty, the
	// template is derived from the selected runtime.
	Template string

	CPUs      string
	MemoryGiB string
	DiskGiB   string
	// VMType is the Lima VM type (vz|qemu). Empty auto-detects.
	VMType string
}

// FromEnv reads and validates options from the environment.
//
// requireMachineID controls whether MACHINE_ID must be present (true for
// lifecycle verbs, which always operate on a specific instance).
func FromEnv(requireMachineID bool) (*Options, error) {
	rt, err := runtime.FromName(os.Getenv("LIMA_RUNTIME"))
	if err != nil {
		return nil, err
	}

	o := &Options{
		MachineID:     os.Getenv("MACHINE_ID"),
		MachineFolder: os.Getenv("MACHINE_FOLDER"),
		Runtime:       rt,
		LimactlPath:   envOr("LIMA_PATH", "limactl"),
		Template:      os.Getenv("LIMA_TEMPLATE"),
		CPUs:          envOr("LIMA_CPUS", "4"),
		MemoryGiB:     envOr("LIMA_MEMORY", "8"),
		DiskGiB:       envOr("LIMA_DISK", "60"),
		VMType:        os.Getenv("LIMA_VM_TYPE"),
	}

	if requireMachineID && o.MachineID == "" {
		return nil, fmt.Errorf("MACHINE_ID environment variable is missing")
	}

	if o.VMType == "" {
		o.VMType = defaultVMType()
	}

	return o, nil
}

// ResolveTemplate returns the locator to pass to `limactl create`. Precedence:
//  1. an explicit user LIMA_TEMPLATE override (locator or file path);
//  2. the runtime's embedded Lima config, written to writeDir and returned as a
//     file path;
//  3. the runtime's built-in template locator.
func (o *Options) ResolveTemplate(writeDir string) (string, error) {
	if t := strings.TrimSpace(o.Template); t != "" {
		return t, nil
	}

	if cfg := o.Runtime.LimaConfig(); cfg != nil {
		path, err := o.writeConfig(writeDir, cfg)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	return o.Runtime.LimaTemplate(), nil
}

// writeConfig writes an embedded Lima config to a stable path so limactl can
// read it. It prefers the machine folder, falling back to a temp dir.
func (o *Options) writeConfig(writeDir string, cfg []byte) (string, error) {
	dir := writeDir
	if strings.TrimSpace(dir) == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create config dir %q: %w", dir, err)
	}
	path := filepath.Join(dir, "lima-"+o.Runtime.Name()+".yaml")
	if err := os.WriteFile(path, cfg, 0o600); err != nil {
		return "", fmt.Errorf("write lima config: %w", err)
	}
	return path, nil
}

// defaultVMType picks the fastest VM type for the host: vz on Apple Silicon
// macOS, qemu elsewhere.
func defaultVMType() string {
	if goruntime.GOOS == "darwin" && goruntime.GOARCH == "arm64" {
		return "vz"
	}
	return "qemu"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
