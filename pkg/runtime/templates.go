package runtime

import _ "embed"

// Embedded Ubuntu-based Lima configs for runtimes that need provisioning beyond
// Lima's built-in templates. docker uses the built-in template://docker-rootful
// (already Ubuntu-based and proven), so it has no embedded config.

//go:embed templates/podman.yaml
var podmanLimaConfig []byte

//go:embed templates/nerdctl.yaml
var nerdctlLimaConfig []byte
