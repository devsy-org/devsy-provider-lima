## E2E tests

The suite has two label groups:

- `integration` builds the provider and exercises the no-VM paths (manifest
  generation, provider install, option validation, error handling). It needs
  `limactl` on `PATH` (provider install runs the provider's init) but does not
  provision a VM.
- `vm` provisions a real Lima VM and runs a full workspace lifecycle. It
  requires `limactl` and host virtualization.

### Run

Build the provider binaries first, then run a label group from this directory:

```sh
task build:cli
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=integration
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=vm
```
