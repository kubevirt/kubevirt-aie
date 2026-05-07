# NVIDIA DSX Virtualization

This branch contains the open-sourced KubeVirt enablement used by the NVIDIA DSX Virtualization Reference Architecture for Grace-Blackwell systems. It is based on upstream KubeVirt `release-1.7` and carries experimental ARM64 hardware enablement for large GPU passthrough virtual machines on NVIDIA Grace platforms. The goal of the effort is to let KubeVirt turn a Grace-Blackwell compute tray into a dedicated VM isolation boundary for AI workloads on Kubernetes.

## Scope

The branch is intended for tech preview, development, and reference integration. It is not a generic upstream KubeVirt release. It assumes an ARM64 Grace-Blackwell host prepared for PCI passthrough, SMMUv3, IOMMUFD, optional vCMDQ acceleration, and optional EGM-backed guest memory. The Grace host kernel, Grace-VFIO-driver, container runtime and Kubernetes configuration are out of scope.

The implementation focuses on compute tray-size VMs, where a VM consumes the full set of GPUs and EGM devices exposed by the host/device plugin policy. EGM-backed VMs require the guest memory and assigned devices to match the EGM topology selected by the platform integration. Smaller partial-passthrough configurations are possible for non-EGM VMs.

## Relationship to KubeVirt AIE

The DSX KubeVirt work is open sourced as part of the KubeVirt Accelerated Infrastructure Enablement effort. The `release-1.8-aie-nv` branch follows a narrower WG AIE model focused on an alternative `virt-launcher` image approach, which is different from the model used by NVIDIA DSX. This `release-1.7-aie-nv` branch carries a broader experimental KubeVirt delta: alpha user-facing annotations, admission, controller, launcher, topology, and build changes required to validate the DSX Grace-Blackwell KubeVirt enablement stack for the GB200 compute-tray VM model.
The long-term goal is to collaborate under the KubeVirt AIE working group, converge on a consensus upstream design, and consolidate implementation toward the KubeVirt main branch and future releases.

## Architecture

| Area | Purpose |
| :--- | :------ |
| KubeVirt API and admission | Adds alpha opt-in annotations and validates Grace-specific combinations. |
| virt-controller and virt-handler | Preserve normal KubeVirt scheduling and device-plugin resource flow while carrying the extra topology data needed by virt-launcher. |
| virt-launcher conversion | Builds libvirt domain XML with Grace-aware PCIe, NUMA, SMMUv3, IOMMUFD, vCMDQ, and EGM wiring. |
| QEMU and libvirt RPMs | Uses aarch64 RPMs with NVIDIA Grace hardware-enablement patches while leaving other architectures on upstream defaults. |
| GPU device plugin or DRA integration | Advertises GPU resources and mounts the host device nodes required by the launcher pod. |
| Host platform | Provides firmware, kernel, VFIO/IOMMUFD, EGM, and sysfs topology used to construct the guest. |

## Grace Features

- `alpha.kubevirt.io/graceVirtualization` opt-in annotation for baseline Grace host-device wiring plus optional `smmuv3`, `vcmdq`, and `egm` controls.
- `alpha.kubevirt.io/pciHole64Size` annotation to request a larger 64-bit PCI MMIO aperture for large GPU BARs.
- PCIe NUMA topology planning for passthrough devices.
- Guest NUMA distance matrix synthesis for the qualified Grace topology.
- Root-port PCIe link information derived from host topology where available.
- vEGM-backed guest memory wiring for EGM-enabled VMs.
- aarch64 QEMU/libvirt version overrides through `make rpm-deps` without committing private RPM repository URLs.
- Multi-architecture image build and push support for ARM64 development flows.

## Runtime Opt-In

Enable the Grace feature gates in the KubeVirt CR before using the annotation:

```yaml
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - HostDevices
      - GraceIOVirtualization
      - PCINUMAAwareTopology
```

Use the annotation only for ARM64 VMIs that consume passthrough devices:

```yaml
metadata:
  annotations:
    alpha.kubevirt.io/pciHole64Size: "4294967296" # 4 TiB, in KiB
    alpha.kubevirt.io/graceVirtualization: '{"smmuv3":true,"vcmdq":true,"egm":true}'
```

Supported `graceVirtualization` keys are:

| Key | Meaning |
| :-- | :------ |
| `smmuv3` | Requests SMMUv3/IOMMU wiring. |
| `vcmdq` | Requests vCMDQ acceleration. Requires `smmuv3=true`; for non-EGM VMs, hugepages are also required. |
| `egm` | Requests EGM-backed guest memory. Requires `smmuv3=true`, explicit guest memory, dedicated CPU placement, and no hugepages. |

Omit the annotation for ordinary KubeVirt behavior. Use `{}` only when you want baseline Grace host-device wiring without enabling `smmuv3`, `vcmdq`, or `egm`. The annotation does not request GPUs by itself; the VMI must still use normal KubeVirt `spec.domain.devices.hostDevices` or GPU/DRA configuration, and the resource names must be allowed in `permittedHostDevices`.

## EGM Guest Memory

For EGM-backed VMs, set `spec.domain.memory.guest` to the total EGM memory selected for passthrough. For example, two EGM devices reporting `0x6fc0000000` bytes each provide `915456Mi` total:

```shell
for dev in /sys/class/egm/egm*/egm_size; do
    bytes=$((16#$(cat "$dev" | sed 's/^0x//')))
    echo "$dev: $((bytes / 1024 / 1024)) MiB"
done
```

Then set:

```yaml
domain:
  memory:
    guest: 915456Mi
```

Do not set a small pod memory limit such as `10Gi` when the guest memory is hundreds of GiB. For EGM flows, use scheduling requests for launcher overhead and allow the Grace EGM wiring to back the guest memory.

## RPM Dependency Flow

Private RPM repositories must not be committed. External users should build their own NVIDIA-patched QEMU/libvirt RPMs, serve them from an internal or local HTTP repository, and point KubeVirt at that repository with a local `rpm/nvidia-repo.yaml`.

Example:

```shell
cp rpm/nvidia-repo.yaml.example rpm/nvidia-repo.yaml
cp rpm/arch-overrides-nvidia-grace.sh.example rpm/arch-overrides-nvidia-grace.sh
# update the content in yaml files

make CUSTOM_REPO=rpm/nvidia-repo.yaml \
     ARCH_OVERRIDES=rpm/arch-overrides-nvidia-grace.sh \
     SINGLE_ARCH=aarch64 \
     rpm-deps
```

You can also pass the aarch64 NEVRAs directly:

```shell
make CUSTOM_REPO=rpm/nvidia-repo.yaml \
     QEMU_VERSION_AARCH64=17:10.1.0+nvidia5-1.el9 \
     LIBVIRT_VERSION_AARCH64=0:11.9.0+nvidia4-1.el9 \
     SINGLE_ARCH=aarch64 \
     rpm-deps
```

See the NVIDIA docs in this repository for open-source build references:

- [DSX Virtualization Development Environment](docs/NVIDIA/Grace-Environment-Setup.md)
- [Build NVIDIA QEMU RPMs](docs/NVIDIA/DSX-Virtualization-QEMU-RPM-Build.md)
- [Build NVIDIA libvirt RPMs](docs/NVIDIA/DSX-Virtualization-Libvirt-RPM-Build.md)

## Build Images

After `rpm-deps` updates the Bazel RPM definitions, build and push the KubeVirt images for the target registry:

```shell
DOCKER_PREFIX=quay.io/example/kubevirt-aie \
DOCKER_TAG=release-1.7-aie-nv-dev \
BUILD_ARCH=amd64,crossbuild-aarch64 \
make bazel-push-images
```

Use a registry that is reachable by the Kubernetes nodes where KubeVirt will be deployed.

## Examples

- [Grace-Blackwell EGM VMI](examples/vmi-dsx-grace-blackwell-egm.yaml)

The example is intentionally parameterized. Update the GPU resource name, PVC, network attachment definition, node selector, guest memory, and PCI hole size to match the target system.

## Related Public Work

- [KubeVirt WG AIE](https://github.com/kubevirt/community/tree/main/wg-aie)
- [CentOS Accelerated Infrastructure SIG](https://sigs.centos.org/aie/)
- [VEP-199 NVIDIA Grace-Blackwell Support in KubeVirt](https://github.com/kubevirt/enhancements/pull/270)
- [KubeVirt Summit talk, EU 2026](https://www.youtube.com/watch?v=jtnRFgu4tdI)

## NVIDIA Contact

- [Fan Zhang](https://github.com/fanzhangio), NVIDIA

## Upstream KubeVirt

KubeVirt is a virtual machine management add-on for Kubernetes. For general KubeVirt documentation, see:

- [KubeVirt user guide](https://kubevirt.io/user-guide)
- [KubeVirt architecture](https://github.com/kubevirt/kubevirt/blob/main/docs/architecture.md)
- [KubeVirt API reference](https://kubevirt.io/api-reference/)

## License

KubeVirt is distributed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright The KubeVirt Authors.
