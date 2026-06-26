# KubeVirt AIE

**KubeVirt Accelerated Infrastructure Enablement (AIE)** is a release-branch fork of [KubeVirt](https://github.com/kubevirt/kubevirt) that produces an alternative `virt-launcher` container image based on [CentOS Stream 10](https://centos.org/) with NVIDIA-optimised (el10nv) RPMs for libvirt and QEMU. The image includes IOMMU-FD support and backported device patches required for GPU passthrough on NVIDIA ARM64 platforms such as GraceHopper, GraceBlackwell, and Vera Rubin.

This repository tracks the upstream KubeVirt `release-1.8` branch with a minimal set of additional commits. Only the `virt-launcher` image is produced here; all other KubeVirt components (virt-operator, virt-api, virt-controller, virt-handler) are consumed from the standard KubeVirt release.

## Why a separate fork?

GPU passthrough on ARM64 depends on a newer userspace stack (IOMMU-FD, backported QEMU device patches, libvirt changes) that is only available in the CentOS Stream 10 el10nv RPM variants. Since the standard KubeVirt `virt-launcher` is based on CentOS Stream 9, the alternative image must be built and delivered separately.

The [WG AIE](https://github.com/kubevirt/community/tree/main/wg-aie) requires that changes carried in this fork are upstreamed back into [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt) as soon as possible. Retiring this fork requires upstreaming at two levels: the el10nv RPM patches must be integrated into the main CentOS Stream 10 repositories by the AIE SIG, and the broader [CentOS Stream 10 transition (VEP #210)](https://github.com/kubevirt/enhancements/issues/210) must land in KubeVirt (planned across v1.9-v1.11). Only once both are complete can the standard KubeVirt `virt-launcher` build include the necessary hardware enablement and this fork be retired.

## KubeVirt WG AIE

The [KubeVirt Accelerated Infrastructure Enablement Working Group (WG AIE)](https://github.com/kubevirt/community/tree/main/wg-aie) aims to track and where possible extend KubeVirt in line with the CentOS Stream Accelerated Infrastructure Enablement SIG. The working group is a stakeholder of both sig-compute and sig-network within the KubeVirt community.

**Chairs:**

- [Lee Yarwood](https://github.com/lyarwood), Red Hat
- [Alay Patel](https://github.com/alaypatel07), NVIDIA

**Meetings:**

- Fortnightly on Thursdays at 15:00 UTC
- [Zoom link](https://zoom.us/j/92261532235)
- [Meeting notes](https://docs.google.com/document/d/1Tz7GubzbSOliRv2kFXamVhV66v1Wo9HH9hwwRQVYrDk)

## CentOS Stream AIE SIG

This work is aligned with the [CentOS Accelerated Infrastructure SIG](https://sigs.centos.org/aie/), which produces the el10nv RPM variants of libvirt and QEMU in CentOS Stream 10. The AIE SIG focuses on enabling accelerated and heterogeneous computing infrastructure, with packages optimised for NVIDIA hardware available from [CentOS Stream 10 koji](https://kojihub.stream.centos.org). WG AIE coordinates the integration of these packages into the KubeVirt ecosystem. For more information on the broader CentOS AIE SIG initiative, see the [CentOS Accelerated Infrastructure SIG documentation](https://docs.centos.org/centos-accelerated-infrastructure-sig/).

## Architecture

The kubevirt-aie `virt-launcher` image is one component of a larger architecture for NVIDIA hardware enablement in KubeVirt. The full system consists of four components:

| Component | Repository | Summary |
| :-------- | :--------- | :------ |
| **kubevirt-aie virt-launcher** | This repo | CentOS Stream 10-based `virt-launcher` with el10nv RPMs for libvirt and QEMU |
| **AIE webhook** | [kubevirt-aie-webhook](https://github.com/kubevirt/kubevirt-aie-webhook) | Mutating admission webhook that replaces the launcher image, injects IOMMUFD resource limits, and optionally adds node affinity |
| **IOMMUFD device plugin** | [iommufd-device-plugin](https://github.com/kubevirt/iommufd-device-plugin) | Kubernetes device plugin that opens and configures `/dev/iommu` on the host and passes the file descriptor to virt-launcher via SCM_RIGHTS |
| **HCO integration** | [hyperconverged-cluster-operator](https://github.com/kubevirt/hyperconverged-cluster-operator) | HCO operand handlers that deploy and reconcile the webhook and device plugin on OpenShift |

For detailed design documents covering each component, see the [kubevirt-aie-veps](https://github.com/kubevirt/kubevirt-aie-veps) repository.

## Rebasing on upstream releases

Manual rebases onto new v1.8.z releases from the upstream [kubevirt/kubevirt `release-1.8`](https://github.com/kubevirt/kubevirt/tree/release-1.8) branch will be performed on this branch. The patch delta is kept intentionally small to minimise rebase friction.

## What this fork changes

The delta against the upstream `release-1.8` branch is intentionally small:

- **NV RPM sync infrastructure** (`hack/sync-nv-rpms.sh`) -- Automated discovery and download of el10nv RPMs from CentOS Stream 10 koji
- **WORKSPACE and rpm/BUILD.bazel updates** -- Bazel build definitions for the el10nv RPM packages
- **CentOS Stream 10 builder** (`hack/builder/Dockerfile.cs10`) -- Build toolchain for multi-architecture (x86_64, aarch64) image builds
- **WIP image push script** (`hack/push-virt-launcher-pr.sh`) -- Convenience script for building and pushing images from feature branches

## Quick start

### Syncing NV RPMs

```shell
# Auto-discover latest el10nv versions from koji
hack/sync-nv-rpms.sh

# Override specific versions
LIBVIRT_NV_VERSION=11.10.0 QEMU_NV_VERSION=10.1.0 hack/sync-nv-rpms.sh
```

### Building and pushing a WIP virt-launcher image

```shell
# Push from current branch with auto-generated tag
hack/push-virt-launcher-pr.sh

# Override registry and tag
DOCKER_PREFIX=quay.io/myorg DOCKER_TAG=test-1 hack/push-virt-launcher-pr.sh
```

### Building release images

```shell
DOCKER_PREFIX=quay.io/kubevirt/kubevirt-aie \
DOCKER_TAG=v1.8.0-aie-nv \
BUILD_ARCH=amd64,crossbuild-aarch64 \
make bazel-push-images
```

## Upstream KubeVirt

This fork is based on [KubeVirt](https://github.com/kubevirt/kubevirt), a virtual machine management add-on for Kubernetes. For general KubeVirt documentation, see:

- [KubeVirt user guide](https://kubevirt.io/user-guide)
- [KubeVirt architecture](https://github.com/kubevirt/kubevirt/blob/main/docs/architecture.md)
- [KubeVirt API reference](https://kubevirt.io/api-reference/)

## Community

- [CentOS Accelerated Infrastructure SIG](https://docs.centos.org/centos-accelerated-infrastructure-sig/) -- Upstream SIG producing the el10nv RPM variants
- [KubeVirt Slack](https://kubernetes.slack.com/?redir=%2Farchives%2FC8ED7RKFE) -- `#virtualization` on kubernetes.slack.com
- [kubevirt-dev Google Group](https://groups.google.com/forum/#!forum/kubevirt-dev)

## License

KubeVirt is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright The KubeVirt Authors.

[//]: # (Reference links)
   [k8s]: https://kubernetes.io
   [crd]: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
   [libvirt]: https://www.libvirt.org
