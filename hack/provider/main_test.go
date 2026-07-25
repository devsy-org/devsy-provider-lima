package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildBinaryReleaseURL(t *testing.T) {
	cfg := &buildConfig{
		version:     "v1.2.3",
		projectRoot: "https://github.com/devsy-org/devsy-provider-lima/releases/download/v1.2.3",
		isRelease:   true,
		checksums:   map[string]string{"devsy-provider-lima-linux-arm64": "deadbeef"},
	}
	got := buildBinary(cfg, "linux/arm64")

	want := "https://github.com/devsy-org/devsy-provider-lima/releases/download/v1.2.3/devsy-provider-lima-linux-arm64"
	if got.Path != want {
		t.Errorf("Path = %q, want %q", got.Path, want)
	}
	if got.Checksum != "deadbeef" {
		t.Errorf("Checksum = %q, want deadbeef", got.Checksum)
	}
	if got.OS != "linux" || got.Arch != "arm64" {
		t.Errorf("OS/Arch = %s/%s, want linux/arm64", got.OS, got.Arch)
	}
	// Release mode must not include the goreleaser build dir.
	if strings.Contains(got.Path, "build_linux_arm64") {
		t.Errorf("release path should not contain a build dir: %q", got.Path)
	}
}

func TestBuildBinaryLocalFilesystem(t *testing.T) {
	cfg := &buildConfig{projectRoot: "./dist", isRelease: false, checksums: map[string]string{}}
	got := buildBinary(cfg, "darwin/arm64")

	abs, _ := filepath.Abs("./dist")
	want := filepath.Join(abs, "build_darwin_arm64_v8.0", "devsy-provider-lima-darwin-arm64")
	if got.Path != want {
		t.Errorf("Path = %q, want %q", got.Path, want)
	}
	if got.Checksum != "" {
		t.Errorf("missing checksum should be empty, got %q", got.Checksum)
	}
}

func TestBuildBinaryLocalHTTP(t *testing.T) {
	cfg := &buildConfig{
		projectRoot: "http://localhost:8080",
		isRelease:   false,
		checksums:   map[string]string{},
	}
	got := buildBinary(cfg, "linux/amd64")

	want := "http://localhost:8080/build_linux_amd64_v1/devsy-provider-lima-linux-amd64"
	if got.Path != want {
		t.Errorf("Path = %q, want %q", got.Path, want)
	}
}

func TestParseChecksums(t *testing.T) {
	// goreleaser's checksums.txt uses the sha256sum format: hash, two spaces, name.
	content := "aaa111  devsy-provider-lima-linux-arm64\n" +
		"bbb222  devsy-provider-lima-darwin-arm64\n" +
		"\n" +
		"malformed-line-without-separator\n"
	path := filepath.Join(t.TempDir(), "checksums.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	sums, err := parseChecksums(path)
	if err != nil {
		t.Fatal(err)
	}
	if sums["devsy-provider-lima-linux-arm64"] != "aaa111" {
		t.Errorf("linux-arm64 = %q, want aaa111", sums["devsy-provider-lima-linux-arm64"])
	}
	if sums["devsy-provider-lima-darwin-arm64"] != "bbb222" {
		t.Errorf("darwin-arm64 = %q, want bbb222", sums["devsy-provider-lima-darwin-arm64"])
	}
}

func TestAllPlatformsExcludesWindows(t *testing.T) {
	for _, p := range allPlatforms() {
		if strings.HasPrefix(p, "windows") {
			t.Errorf("windows platform %q should be excluded (Lima is unix-only)", p)
		}
		if buildDir(p) == "" {
			t.Errorf("no build dir mapping for platform %q", p)
		}
	}
}
