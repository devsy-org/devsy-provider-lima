// Package runtime maps a container runtime choice (docker, podman, nerdctl) onto
// the Lima template or config that provisions it.
//
// All three runtimes ride the existing devsy Docker driver, which already
// abstracts docker/podman/nerdctl (see devsy pkg/docker/runtime.go). The VM is a
// Tier-1 machine target; no new devsy driver is required.
package runtime

import "fmt"

// Names of the supported runtimes. Each value is also the container CLI binary
// name inside the VM (the generated manifest sets docker.path to it).
const (
	Docker  = "docker"
	Podman  = "podman"
	Nerdctl = "nerdctl"
)

// Runtime describes how a container runtime is provisioned in the Lima VM and
// how devsy should drive containers with it.
type Runtime interface {
	// Name is the runtime identifier (docker|podman|nerdctl).
	Name() string

	// LimaTemplate is the Lima template locator that provisions this runtime,
	// e.g. "template://docker-rootful". Used when LimaConfig returns nil.
	LimaTemplate() string

	// LimaConfig returns an embedded Lima YAML config to provision this runtime,
	// or nil to use LimaTemplate instead. Runtimes that need provisioning beyond
	// Lima's built-in templates (podman, nerdctl) ship an embedded Ubuntu-based
	// config here; the provider writes it to disk and passes it to limactl.
	LimaConfig() []byte
}

// FromName returns the Runtime for the given name.
func FromName(name string) (Runtime, error) {
	switch name {
	case Docker, "":
		return dockerRuntime{}, nil
	case Podman:
		return podmanRuntime{}, nil
	case Nerdctl:
		return nerdctlRuntime{}, nil
	default:
		return nil, fmt.Errorf(
			"unknown runtime %q, expected one of: %s, %s, %s",
			name, Docker, Podman, Nerdctl,
		)
	}
}
