# Lima Provider for Devsy

Runs Devsy workspaces inside [Lima](https://lima-vm.io) virtual machines. Each
workspace gets a dedicated Lima VM; the devcontainer is deployed by Devsy's
Docker driver inside that VM, so the VM is the isolation boundary and existing
`devcontainer.json` configurations work unchanged.

## Getting started

Install [Lima](https://lima-vm.io) first (`limactl` must be on `PATH`):

```sh
brew install lima
```

Then add the provider:

```sh
devsy provider add devsy-org/devsy-provider-lima
devsy provider use lima
```

Lima is available for macOS and Linux; it does not run on Windows.

### Creating your first devsy env

```sh
devsy workspace up .
```

The first run downloads the VM image and provisions the container runtime, so it
takes a few minutes.

## Container runtimes

The `LIMA_RUNTIME` option selects the container runtime provisioned in the VM.
All three are driven by Devsy's Docker driver.

| Runtime | Provisioning |
| ------- | ------------ |
| `docker` (default) | Lima's `docker-rootful` template |
| `podman` | embedded Ubuntu config, rootful Podman |
| `nerdctl` | embedded Ubuntu config, system containerd |

Override the template for any runtime with `LIMA_TEMPLATE`, for example to use a
different base image or a custom Lima configuration.

## Multiple workspaces on one VM

By default each workspace gets its own VM. To run multiple devcontainers on a
single shared VM, enable single-machine mode:

```sh
devsy provider init lima --single-machine
```

Devsy then reuses one VM for all workspaces on this provider and deletes it only
when the last workspace is removed.

## Options

| NAME | REQUIRED | DESCRIPTION | DEFAULT |
| ---- | -------- | ----------- | ------- |
| LIMA_RUNTIME | false | Container runtime in the VM: docker, podman, or nerdctl. | docker |
| LIMA_CPUS | false | Number of CPUs to allocate to the VM. | 4 |
| LIMA_MEMORY | false | Memory in GiB to allocate to the VM. | 8 |
| LIMA_DISK | false | Disk size in GiB for the VM. | 60 |
| LIMA_VM_TYPE | false | Lima VM type (vz or qemu). Empty auto-detects (vz on Apple Silicon, qemu otherwise). | |
| LIMA_TEMPLATE | false | Override the runtime-derived Lima template locator. | |
| LIMA_PATH | false | Path to the limactl binary. | limactl |
| AGENT_PATH | false | Path where the Devsy agent is injected in the VM. | /tmp/devsy |
| INACTIVITY_TIMEOUT | false | Stops the VM after the given idle period, e.g. 10m or 1h. | 10m |
| INJECT_GIT_CREDENTIALS | false | Inject git credentials into the VM. | true |
| INJECT_DOCKER_CREDENTIALS | false | Inject docker credentials into the VM. | true |

Set options at add time or later:

```sh
devsy provider add devsy-org/devsy-provider-lima -o LIMA_CPUS=8 -o LIMA_MEMORY=16
devsy provider set lima --option LIMA_RUNTIME=podman
```

## SELinux hosts

On SELinux-enforcing VM images (Fedora, RHEL, CentOS) the container runtime
labels containers, which blocks the Devsy agent from exec-ing into them. Devsy
disables the container SELinux label on enforcing hosts, so those images work
without extra configuration.

## Development

```sh
task build:cli            # cross-compile binaries into ./dist
task build:provider:dev   # generate ./dist/provider.yaml pointing at ./dist
task test
task lint
```

## License

MPL-2.0
