package runtime

// podmanRuntime uses an embedded Ubuntu-based rootful config. Lima's Fedora
// podman template breaks in-container exec ("/bin/bash: permission denied").
type podmanRuntime struct{}

func (podmanRuntime) Name() string         { return Podman }
func (podmanRuntime) LimaTemplate() string { return "" }
func (podmanRuntime) LimaConfig() []byte   { return podmanLimaConfig }
