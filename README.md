# devsy-provider-lima

A [Devsy](https://github.com/devsy-org/devsy) machine provider that runs
workspaces inside [Lima](https://lima-vm.io) virtual machines.

Each workspace gets a dedicated Lima VM with a configurable container runtime.
The devcontainer is deployed by Devsy's existing Docker driver **inside** the VM,
so the VM is the isolation boundary and existing `devcontainer.json` configs work
unchanged.

## Design

Devsy has a two-tier model: a **machine provider** (Tier 1) owns the VM; a
**driver** (Tier 2) runs the container. This provider is Tier 1 only — it manages
the Lima VM lifecycle via the `limactl` CLI and lets the existing Docker driver
(which already speaks docker/podman/nerdctl) run the container in the VM. No new
Devsy driver is required.

The provider **shells out to `limactl`** rather than importing the Lima Go
packages: `limactl` is the stable, supported interface, and importing Lima as a
library would pull an enormous vz/qemu dependency tree.

### Container runtimes

Selected via the `LIMA_RUNTIME` option. All ride the Docker driver through
`agent.docker.runtime`:

| Runtime | Lima provisioning | Status |
|---------|-------------------|--------|
| `docker` (default) | built-in `template://docker-rootful` (Ubuntu) | ✅ validated E2E: up, connect, exec, stop, restart, reset |
| `podman` | embedded Ubuntu config, rootful podman | ✅ validated E2E: up, connect, exec, stop, restart, reset |
| `nerdctl` | embedded Ubuntu config, system containerd + nerdctl-full | ✅ validated E2E: up, connect, exec |

devsy runs the container runtime via `sudo` (rootful) and execs into the
container as the workspace user. The key lesson: this needs a **system**
daemon/socket, not Lima's rootless defaults.

- Lima's Fedora-based `template://podman` starts the container but breaks
  in-container exec (`/bin/bash: permission denied`); an **Ubuntu-based** rootful
  podman (same OS family as docker-rootful) works.
- Lima's `template://default` runs **rootless** containerd, but devsy targets the
  system socket `/run/containerd/containerd.sock`; the embedded config installs
  the `nerdctl-full` bundle and enables **system** containerd.

The `docker` and `podman` runtimes use built-in Lima templates and the embedded
Ubuntu config respectively; `nerdctl` uses an embedded config that downloads the
nerdctl-full release at VM provisioning time. Override any of it with the
`LIMA_TEMPLATE` option.

`docker` uses the **rootful** template deliberately: Devsy overrides
`DOCKER_CONFIG` when injecting its credential helper, which bypasses the rootless
template's per-user docker context and causes `permission denied`. The rootful
template owns `/var/run/docker.sock` as the VM user, avoiding that.

## Validated devcontainer features (docker runtime)

Tested end-to-end on the lima provider: prebuilt `image`, `build` from a
Dockerfile, `dockerComposeFile`, `forwardPorts` (host→container through devsy's
tunnel), lifecycle `stop`/restart/`--recreate`, and `LIMA_TEMPLATE` override.

## Usage

Requires [Lima](https://lima-vm.io) installed on the host (`limactl` on `PATH`).

```sh
devsy provider add devsy-org/devsy-provider-lima
devsy workspace up <repo> --provider lima
```

## Provider commands

The binary implements the Devsy provider verbs as subcommands, invoked by Devsy
via the generated `provider.yaml`:

| Verb | Action |
|------|--------|
| `init` | Validate options and that `limactl` is runnable |
| `create` | Create + start the Lima VM |
| `start` / `stop` | Start / stop the VM |
| `delete` | Force-delete the VM |
| `status` | Report `Running`/`Stopped`/`Busy`/`NotFound` |
| `command` | Run a command in the VM (exec channel; `limactl shell`) |

## Development

```sh
task build:cli            # cross-compile binaries into ./dist
task build:provider:dev   # generate ./dist/provider.yaml pointing at ./dist
task test
task lint
```

Add the locally built provider:

```sh
task build:provider:dev
devsy provider add ./dist/provider.yaml --name lima
```

## License

MPL-2.0
