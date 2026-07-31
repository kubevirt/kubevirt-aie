# KubeVirt AIE — NVIDIA Platform Branch

**KubeVirt Accelerated Infrastructure Enablement (AIE)** is a release-branch fork of [KubeVirt](https://github.com/kubevirt/kubevirt) that produces an alternative `virt-launcher` container image. This **`release-1.9-nvidia`** branch is maintained by NVIDIA for developing and validating platform-specific features using NVIDIA's own kernel, libvirt, and QEMU packages before upstreaming to [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt).

## Purpose

This branch exists to enable rapid development and validation of NVIDIA-specific GPU virtualisation features on Grace Blackwell and Vera Rubin platforms. Unlike the [`release-1.9-aie-nv`](https://github.com/kubevirt/kubevirt-aie/tree/release-1.9-aie-nv) branch (which uses [CentOS Stream 10 AIE SIG](https://sigs.centos.org/aie/) el10nv RPMs), this branch builds against NVIDIA's own upstream kernel, libvirt, and QEMU trees:

| Component | NVIDIA upstream | CS10 AIE SIG (release-1.9-aie-nv) |
| :-------- | :-------------- | :--------------------------------- |
| **libvirt** | [nvidia_unstable-12.4](https://github.com/NVIDIA/libvirt/commits/nvidia_unstable-12.4/) | 11.10.0-10.7.el10nv |
| **QEMU** | [nvidia_unstable-11.0](https://github.com/NVIDIA/QEMU/commits/nvidia_unstable-11.0/) | 10.1.0-19.el10nv.1 |
| **kernel** | [NV-Kernels](https://github.com/NVIDIA/NV-Kernels) (6.18+) | 6.12.0-239.13.el10nv |

This allows NVIDIA to develop features that depend on newer libvirt/QEMU/kernel capabilities not yet available in the CentOS Stream AIE SIG packages.

## Target features

Features being developed and validated on this branch for upstream inclusion:

- **vEGM** (Virtual Extended GPU Memory) — QEMU/libvirt patches landed in CS10 AIE SIG
- **vCMDQ** (Virtual Command Queue) — QEMU/libvirt patches landed in CS10 AIE SIG
- **CXL Type-2 passthrough** — upstream kernel work ([LWN](https://lwn.net/Articles/1062595/))
- **ARM-CCA** (Confidential Compute Architecture) — TEE support for KubeVirt ([VEP-338](https://github.com/kubevirt/enhancements/issues/338))
- **Secure Boot on ARM** — ([VEP-227](https://github.com/kubevirt/enhancements/issues/227), [PR #17135](https://github.com/kubevirt/kubevirt/pull/17135))

## Relationship to other branches

| Branch | Maintainer | RPM source | Purpose |
| :----- | :--------- | :--------- | :------ |
| `release-1.9-nvidia` (this branch) | NVIDIA | NVIDIA upstream trees | Platform-specific feature development |
| `release-1.9-aie-nv` | Red Hat / WG-AIE | CS10 AIE SIG el10nv | Production virt-launcher for CNV |
| `main` (kubevirt/kubevirt) | KubeVirt community | CentOS Stream 10 | Upstream KubeVirt |

Changes validated on this branch should be:
1. **Upstreamed to [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)** via VEPs and PRs
2. **RPM patches contributed to the [CentOS Stream AIE SIG](https://sigs.centos.org/aie/)** for inclusion in el10nv packages
3. **Cherry-picked to `release-1.9-aie-nv`** where appropriate for production use

## KubeVirt WG AIE

The [KubeVirt Accelerated Infrastructure Enablement Working Group (WG AIE)](https://github.com/kubevirt/community/tree/main/wg-aie) coordinates upstream integration of accelerated infrastructure features across Red Hat and NVIDIA.

**Chairs:**

- [Lee Yarwood](https://github.com/lyarwood), Red Hat
- [Fan Zhang](https://github.com/fanzhangio), NVIDIA

**Meetings:**

- Fortnightly on Thursdays at 15:00 UTC
- [Zoom link](https://zoom.us/j/92261532235)
- [Meeting notes](https://docs.google.com/document/d/1Tz7GubzbSOliRv2kFXamVhV66v1Wo9HH9hwwRQVYrDk)
- [CNCF Slack: #kubevirt-wg-aie](https://cloud-native.slack.com/archives/C0ASUJQEWAE)

## Architecture

The kubevirt-aie `virt-launcher` image is one component of a larger architecture for NVIDIA hardware enablement in KubeVirt. The full system consists of four components:

| Component | Repository | Summary |
| :-------- | :--------- | :------ |
| **kubevirt-aie virt-launcher** | This repo | CentOS Stream 10-based `virt-launcher` with NVIDIA-optimised RPMs for libvirt and QEMU |
| **AIE webhook** | [kubevirt-aie-webhook](https://github.com/kubevirt/kubevirt-aie-webhook) | Mutating admission webhook that replaces the launcher image, injects IOMMUFD resource limits, and optionally adds node affinity |
| **IOMMUFD device plugin** | [iommufd-device-plugin](https://github.com/kubevirt/iommufd-device-plugin) | Kubernetes device plugin that opens and configures `/dev/iommu` on the host and passes the file descriptor to virt-launcher via SCM_RIGHTS |
| **HCO integration** | [hyperconverged-cluster-operator](https://github.com/kubevirt/hyperconverged-cluster-operator) | HCO operand handlers that deploy and reconcile the webhook and device plugin on OpenShift |

## Building with NVIDIA RPMs

To build a `virt-launcher` image using NVIDIA's custom libvirt and QEMU RPMs, the existing RPM customisation support in `hack/sync-nv-rpms.sh` can be extended with an overlay approach. See the [`release-1.9-aie-nv`](https://github.com/kubevirt/kubevirt-aie/tree/release-1.9-aie-nv) branch for the baseline sync infrastructure.

## Upstream KubeVirt

This fork is based on [KubeVirt](https://github.com/kubevirt/kubevirt) v1.9, a virtual machine management add-on for Kubernetes. For general KubeVirt documentation, see:

- [KubeVirt user guide](https://kubevirt.io/user-guide)
- [KubeVirt architecture](https://github.com/kubevirt/kubevirt/blob/main/docs/architecture.md)
- [KubeVirt API reference](https://kubevirt.io/api-reference/)

## Community

- [CentOS Accelerated Infrastructure SIG](https://docs.centos.org/centos-accelerated-infrastructure-sig/) — Upstream SIG producing the el10nv RPM variants
- [KubeVirt Slack](https://kubernetes.slack.com/?redir=%2Farchives%2FC8ED7RKFE) — `#virtualization` on kubernetes.slack.com
- [kubevirt-dev Google Group](https://groups.google.com/forum/#!forum/kubevirt-dev)
