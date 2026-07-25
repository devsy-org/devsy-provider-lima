// Package runtime maps a container runtime to the Lima config that provisions it.
package runtime

import "fmt"

const (
	Docker  = "docker"
	Podman  = "podman"
	Nerdctl = "nerdctl"
)

type Runtime interface {
	Name() string
	LimaTemplate() string
	// LimaConfig returns an embedded config, or nil to use LimaTemplate.
	LimaConfig() []byte
}

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
