package runtime

// podmanRuntime provisions rootful Podman via an embedded Ubuntu-based Lima
// config (pkg/runtime/templates/podman.yaml).
//
// It deliberately does NOT use Lima's Fedora-based template://podman: devsy runs
// the runtime via sudo (rootful) and execs into the container as the workspace
// user, and on the Fedora template that in-container exec fails with
// "/bin/bash: permission denied". The Ubuntu base (same OS family as the
// docker-rootful template) avoids this.
type podmanRuntime struct{}

func (podmanRuntime) Name() string         { return Podman }
func (podmanRuntime) LimaTemplate() string { return "" }
func (podmanRuntime) LimaConfig() []byte   { return podmanLimaConfig }
