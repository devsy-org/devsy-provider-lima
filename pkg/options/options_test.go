package options

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devsy-org/devsy-provider-lima/pkg/runtime"
)

func TestResolveTemplateOverrideWins(t *testing.T) {
	podman, _ := runtime.FromName(runtime.Podman)
	o := &Options{Runtime: podman, Template: "template://custom"}
	got, err := o.ResolveTemplate(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "template://custom" {
		t.Errorf("got %q, want the override", got)
	}
}

func TestResolveTemplateEmbeddedConfig(t *testing.T) {
	podman, _ := runtime.FromName(runtime.Podman)
	dir := t.TempDir()
	got, err := (&Options{Runtime: podman}).ResolveTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != dir {
		t.Errorf("config written to %q, want under %q", got, dir)
	}
	b, err := os.ReadFile(got) //nolint:gosec // reads a file the test just wrote
	if err != nil || len(b) == 0 {
		t.Errorf("embedded config not written: err=%v len=%d", err, len(b))
	}
}

func TestResolveTemplateEmptyWriteDirFallsBackToTemp(t *testing.T) {
	podman, _ := runtime.FromName(runtime.Podman)
	got, err := (&Options{Runtime: podman}).ResolveTemplate("")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != filepath.Clean(os.TempDir()) {
		t.Errorf("config written to %q, want under temp dir %q", got, os.TempDir())
	}
	if base := filepath.Base(got); base != "lima-podman.yaml" {
		t.Errorf("config file = %q, want lima-podman.yaml", base)
	}
	_ = os.Remove(got)
}

func TestResolveTemplateBuiltinLocator(t *testing.T) {
	docker, _ := runtime.FromName(runtime.Docker)
	got, err := (&Options{Runtime: docker}).ResolveTemplate(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "template://docker-rootful" {
		t.Errorf("got %q, want the docker built-in locator", got)
	}
}

func clearLimaEnv(t *testing.T) {
	for _, k := range []string{
		"MACHINE_ID", "MACHINE_FOLDER", "LIMA_RUNTIME", "LIMA_PATH",
		"LIMA_TEMPLATE", "LIMA_CPUS", "LIMA_MEMORY", "LIMA_DISK", "LIMA_VM_TYPE",
	} {
		t.Setenv(k, "")
	}
}

func TestFromEnvDefaults(t *testing.T) {
	clearLimaEnv(t)
	t.Setenv("MACHINE_ID", "ws")
	o, err := FromEnv(true)
	if err != nil {
		t.Fatal(err)
	}
	if o.CPUs != "4" || o.MemoryGiB != "8" || o.DiskGiB != "60" {
		t.Errorf("unexpected size defaults: %+v", o)
	}
	if o.LimactlPath != "limactl" || o.Runtime.Name() != runtime.Docker {
		t.Errorf("unexpected path/runtime defaults: %+v", o)
	}
	if o.VMType == "" {
		t.Error("VMType should be auto-detected, got empty")
	}
}

func TestFromEnvMachineIDRequired(t *testing.T) {
	clearLimaEnv(t)
	if _, err := FromEnv(true); err == nil {
		t.Fatal("expected error when MACHINE_ID is required and missing")
	}
	if _, err := FromEnv(false); err != nil {
		t.Fatalf("unexpected error when MACHINE_ID not required: %v", err)
	}
}

func TestFromEnvInvalidRuntime(t *testing.T) {
	clearLimaEnv(t)
	t.Setenv("MACHINE_ID", "ws")
	t.Setenv("LIMA_RUNTIME", "bogus")
	if _, err := FromEnv(true); err == nil {
		t.Fatal("expected error for invalid LIMA_RUNTIME")
	}
}

func TestFromEnvOverrides(t *testing.T) {
	clearLimaEnv(t)
	t.Setenv("MACHINE_ID", "ws")
	t.Setenv("LIMA_CPUS", "8")
	t.Setenv("LIMA_VM_TYPE", "qemu")
	o, err := FromEnv(true)
	if err != nil {
		t.Fatal(err)
	}
	if o.CPUs != "8" || o.VMType != "qemu" {
		t.Errorf("overrides not applied: cpus=%q vmType=%q", o.CPUs, o.VMType)
	}
}
