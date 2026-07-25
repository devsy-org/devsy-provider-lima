package runtime

// dockerRuntime uses the rootful template: devsy overrides DOCKER_CONFIG, which
// breaks the rootless template's per-user docker context. Rootful owns
// /var/run/docker.sock as the VM user instead.
type dockerRuntime struct{}

func (dockerRuntime) Name() string         { return Docker }
func (dockerRuntime) LimaTemplate() string { return "template://docker-rootful" }
func (dockerRuntime) LimaConfig() []byte   { return nil }
