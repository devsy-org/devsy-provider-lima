// Command provider generates provider.yaml.
// Usage: go run ./hack/provider/main.go <version> > ./dist/provider.yaml
package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	providerName = "lima"
	githubOwner  = "devsy-org"
	githubRepo   = "devsy-provider-lima"
	driverDocker = "docker"
)

type Provider struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Icon         string            `yaml:"icon,omitempty"`
	OptionGroups []OptionGroup     `yaml:"optionGroups"`
	Options      Options           `yaml:"options"`
	Agent        Agent             `yaml:"agent"`
	Binaries     Binaries          `yaml:"binaries"`
	Exec         map[string]string `yaml:"exec"`
}

type OptionGroup struct {
	Name           string   `yaml:"name"`
	DefaultVisible bool     `yaml:"defaultVisible"`
	Options        []string `yaml:"options"`
}

type Options map[string]Option

type Option struct {
	Description string   `yaml:"description,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Suggestions []string `yaml:"suggestions,omitempty"`
}

type Agent struct {
	Path                    string            `yaml:"path"`
	Driver                  string            `yaml:"driver"`
	Docker                  DockerAgent       `yaml:"docker"`
	InactivityTimeout       string            `yaml:"inactivityTimeout"`
	InjectGitCredentials    string            `yaml:"injectGitCredentials"`
	InjectDockerCredentials string            `yaml:"injectDockerCredentials"`
	Exec                    map[string]string `yaml:"exec"`
}

type DockerAgent struct {
	Path    string `yaml:"path"`
	Install string `yaml:"install"`
}

type Binaries struct {
	LimaProvider []Binary `yaml:"LIMA_PROVIDER"`
}

type Binary struct {
	OS       string `yaml:"os"`
	Arch     string `yaml:"arch"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum"`
}

type buildConfig struct {
	version     string
	projectRoot string
	isRelease   bool
	checksums   map[string]string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}

	cfg, err := newBuildConfig(os.Args[1])
	if err != nil {
		return err
	}

	output, err := yaml.Marshal(buildProvider(cfg))
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	_, err = os.Stdout.Write(output)
	return err
}

func newBuildConfig(version string) (*buildConfig, error) {
	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		owner := envOr("GITHUB_OWNER", githubOwner)
		projectRoot = fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s",
			owner, githubRepo, version,
		)
	}

	isRelease := strings.Contains(projectRoot, "github.com") &&
		strings.Contains(projectRoot, "/releases/")

	return &buildConfig{
		version:     version,
		projectRoot: projectRoot,
		isRelease:   isRelease,
		checksums:   checksums,
	}, nil
}

func buildProvider(cfg *buildConfig) Provider {
	return Provider{
		Name:    providerName,
		Version: cfg.version,
		Description: "Devsy on Lima virtual machines. Each workspace runs in a " +
			"dedicated Lima VM with a configurable container runtime.",
		OptionGroups: buildOptionGroups(),
		Options:      buildOptions(),
		Agent:        buildAgent(),
		Binaries:     Binaries{LimaProvider: buildBinaryList(cfg, allPlatforms())},
		Exec: map[string]string{
			"init":    "${LIMA_PROVIDER} init",
			"command": "${LIMA_PROVIDER} command",
			"create":  "${LIMA_PROVIDER} create",
			"delete":  "${LIMA_PROVIDER} delete",
			"start":   "${LIMA_PROVIDER} start",
			"stop":    "${LIMA_PROVIDER} stop",
			"status":  "${LIMA_PROVIDER} status",
		},
	}
}

func buildOptionGroups() []OptionGroup {
	return []OptionGroup{
		{
			Name:           "VM options",
			DefaultVisible: true,
			Options: []string{
				"LIMA_RUNTIME", "LIMA_CPUS", "LIMA_MEMORY", "LIMA_DISK",
				"LIMA_VM_TYPE", "LIMA_TEMPLATE",
			},
		},
		{
			Name:           "Agent options",
			DefaultVisible: false,
			Options: []string{
				"LIMA_PATH", "AGENT_PATH", "INACTIVITY_TIMEOUT",
				"INJECT_DOCKER_CREDENTIALS", "INJECT_GIT_CREDENTIALS",
			},
		},
	}
}

func buildOptions() Options {
	return Options{
		"LIMA_RUNTIME": {
			Description: "Container runtime provisioned in the VM: " +
				"docker (default), podman, or nerdctl.",
			Default:     driverDocker,
			Suggestions: []string{driverDocker, "podman", "nerdctl"},
		},
		"LIMA_CPUS":   {Description: "Number of CPUs to allocate to the VM.", Default: "4"},
		"LIMA_MEMORY": {Description: "Memory (in GiB) to allocate to the VM.", Default: "8"},
		"LIMA_DISK":   {Description: "Disk size (in GiB) for the VM.", Default: "60"},
		"LIMA_VM_TYPE": {
			Description: "Lima VM type. Leave empty to auto-detect (vz on Apple " +
				"Silicon, qemu elsewhere).",
			Default:     "",
			Suggestions: []string{"vz", "qemu"},
		},
		"LIMA_TEMPLATE": {
			Description: "Override the Lima template locator. Leave empty to derive " +
				"it from LIMA_RUNTIME.",
			Default: "",
		},
		"LIMA_PATH": {Description: "Path to the limactl binary.", Default: "limactl"},
		"AGENT_PATH": {
			Description: "The path where to inject the Devsy agent inside the VM.",
			Default:     "/tmp/devsy",
		},
		"INACTIVITY_TIMEOUT": {
			Description: "If defined, automatically stops the VM after the inactivity period.",
			Default:     "10m",
		},
		"INJECT_GIT_CREDENTIALS": {
			Description: "If Devsy should inject git credentials into the VM.",
			Default:     "true",
		},
		"INJECT_DOCKER_CREDENTIALS": {
			Description: "If Devsy should inject docker credentials into the VM.",
			Default:     "true",
		},
	}
}

func buildAgent() Agent {
	//nolint:gosec // G101: Inject*Credentials values are env-var placeholders, not secrets.
	return Agent{
		Path:   "${AGENT_PATH}",
		Driver: driverDocker,
		Docker: DockerAgent{
			// docker.runtime is deliberately unset: devsy does not substitute
			// ${OPTION} into it, so it auto-detects from docker.path instead.
			Path:    "${LIMA_RUNTIME}",
			Install: "false",
		},
		InactivityTimeout:       "${INACTIVITY_TIMEOUT}",
		InjectGitCredentials:    "${INJECT_GIT_CREDENTIALS}",
		InjectDockerCredentials: "${INJECT_DOCKER_CREDENTIALS}",
		// Halting the guest stops the Lima VM (how the inactivity timeout works).
		Exec: map[string]string{"shutdown": "sudo shutdown -h now || shutdown -h now"},
	}
}

func buildBinaryList(cfg *buildConfig, platforms []string) []Binary {
	result := make([]Binary, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, buildBinary(cfg, platform))
	}
	return result
}

func buildBinary(cfg *buildConfig, platform string) Binary {
	goos, arch, _ := strings.Cut(platform, "/")

	path := cfg.projectRoot
	if !cfg.isRelease {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			joined, _ := url.JoinPath(path, buildDir(platform))
			path = joined
		} else {
			absPath, _ := filepath.Abs(path)
			path = filepath.Join(absPath, buildDir(platform))
		}
	}

	filename := fmt.Sprintf("devsy-provider-%s-%s-%s", providerName, goos, arch)

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		path, _ = url.JoinPath(path, filename)
	} else {
		path = filepath.Join(path, filename)
	}

	return Binary{OS: goos, Arch: arch, Path: path, Checksum: cfg.checksums[filename]}
}

func buildDir(platform string) string {
	dirs := map[string]string{
		"linux/amd64":  "build_linux_amd64_v1",
		"linux/arm64":  "build_linux_arm64_v8.0",
		"darwin/amd64": "build_darwin_amd64_v1",
		"darwin/arm64": "build_darwin_arm64_v8.0",
	}
	return dirs[platform]
}

// allPlatforms is unix-only; Lima does not run on windows.
func allPlatforms() []string {
	return []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"}
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path) //nolint:gosec // build-time tool, fixed path.
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), " "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}
	return checksums, scanner.Err()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
