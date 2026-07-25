package runtime

// dockerRuntime provisions rootful Docker via Lima's docker-rootful template.
//
// Rootful is required: devsy overrides DOCKER_CONFIG when injecting its
// credential helper, which bypasses the per-user docker context that the
// rootless docker template relies on, so docker would fall back to
// /var/run/docker.sock and fail with "permission denied". The docker-rootful
// template sets docker.socket SocketUser to the VM user, so the default socket
// is owned by that user and works regardless of DOCKER_CONFIG. The VM remains
// the isolation boundary, so rootful-in-VM does not weaken host isolation.
type dockerRuntime struct{}

func (dockerRuntime) Name() string         { return Docker }
func (dockerRuntime) LimaTemplate() string { return "template://docker-rootful" }
func (dockerRuntime) LimaConfig() []byte   { return nil }
