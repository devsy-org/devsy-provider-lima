package runtime

// nerdctlRuntime provisions system (rootful) containerd + nerdctl via an
// embedded Ubuntu-based Lima config (pkg/runtime/templates/nerdctl.yaml).
//
// It deliberately does NOT use Lima's template://default, which runs rootless
// containerd (socket at /run/user/<uid>/containerd/containerd.sock). devsy runs
// nerdctl via sudo and targets the system socket /run/containerd/containerd.sock,
// so a system containerd is required. The embedded config installs the
// nerdctl-full bundle (containerd, runc, CNI, buildkit) and enables system
// containerd.
type nerdctlRuntime struct{}

func (nerdctlRuntime) Name() string         { return Nerdctl }
func (nerdctlRuntime) LimaTemplate() string { return "" }
func (nerdctlRuntime) LimaConfig() []byte   { return nerdctlLimaConfig }
