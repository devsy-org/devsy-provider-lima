package runtime

import _ "embed"

//go:embed templates/podman.yaml
var podmanLimaConfig []byte

//go:embed templates/nerdctl.yaml
var nerdctlLimaConfig []byte
