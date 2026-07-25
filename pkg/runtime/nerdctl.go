package runtime

// nerdctlRuntime uses an embedded config with system containerd. Lima's default
// template is rootless (wrong socket for devsy's sudo-based exec).
type nerdctlRuntime struct{}

func (nerdctlRuntime) Name() string         { return Nerdctl }
func (nerdctlRuntime) LimaTemplate() string { return "" }
func (nerdctlRuntime) LimaConfig() []byte   { return nerdctlLimaConfig }
