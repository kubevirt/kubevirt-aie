# DSX Virtualization Grace Development Environment

Prepared by: [Fan Zhang](https://github.com/fanzhangio), NVIDIA

This document describes an open-source development setup for the NVIDIA DSX Virtualization Reference Architecture on Grace-Blackwell systems. It is a reference checklist for reproducing the KubeVirt AIE `release-1.7-aie-nv` development flow. It is not a product support matrix.

## Assumptions

- ARM64 Grace-Blackwell host with firmware support for GPU passthrough and EGM.
- Linux kernel with KVM, VFIO, SMMUv3, IOMMUFD, and NVIDIA Grace EGM support.
- Kubernetes worker node configured for CPU pinning and device-plugin based passthrough.
- KubeVirt images built from this repository and pushed to a registry reachable by the cluster.
- NVIDIA-patched aarch64 QEMU/libvirt RPMs available from a local or internal HTTP RPM repository.

## Host Validation

Check the host before deploying KubeVirt:

```shell
uname -a
lscpu
ls -l /dev/kvm /dev/vfio/vfio
ls -l /dev/iommu || true
ls -l /dev/egm* || true
lsmod | grep -E 'kvm|vfio|iommufd|nvgrace'
lspci -nn | grep -i nvidia
```

Record host component versions in test reports and bug reports:

```shell
uname -r
containerd --version
kubelet --version
kubectl version
```

KubeVirt uses QEMU and libvirt from its container images, not from host packages. Verify the image RPMs separately before deployment. For example, on a host that can run the target architecture image:

```shell
podman run --rm --arch arm64 --entrypoint rpm <virt-launcher-image> \
  -q qemu-kvm-core libvirt-client libvirt-daemon-driver-qemu json-c

podman run --rm --arch arm64 --entrypoint rpm <virt-handler-image> \
  -q qemu-img json-c
```

After a VMI starts, you can also inspect the launcher pod:

```shell
kubectl -n <vmi-namespace> exec <virt-launcher-pod> -c compute -- \
  rpm -q qemu-kvm-core libvirt-client libvirt-daemon-driver-qemu json-c
```

For each passthrough GPU, inspect the BAR layout:

```shell
lspci -s <gpu-pci-address> -vv
```

Large Grace-Blackwell GPUs can expose a 256 GiB physical resizable BAR and a large virtual resizable BAR. Set `alpha.kubevirt.io/pciHole64Size` large enough for the selected devices. A 4 TiB hole is a practical starting point for the GB200 four-GPU example. GB300 and other platforms must be sized from the assigned device BARs and validated against the generated libvirt domain XML. Increase the hole when QEMU/libvirt reports PCI MMIO allocation failures.

## Prepare the GB200/GB300 Host

Use a distribution kernel that has been validated for Grace-Blackwell virtualization. The kernel must provide KVM on ARM64, VFIO, SMMUv3, IOMMUFD, and the NVIDIA Grace EGM interfaces used by the platform. Install or upgrade the kernel through the distribution package manager, then reboot into the selected kernel:

```shell
sudo dnf upgrade kernel kernel-modules
sudo reboot
uname -r
```

On Debian or Ubuntu based hosts, use the equivalent `apt` packages and reboot flow. The exact package names are distribution-specific; keep the command in deployment docs as a placeholder unless the repository publishes a tested host image or package channel.

Enable EGM in SBIOS before starting EGM-backed VMs. The exact menu path depends on the platform firmware. After reboot, verify that the EGM devices and sysfs attributes are present:

```shell
ls -l /dev/egm*
find /sys/class/egm -maxdepth 2 -type f -name egm_size -print -exec cat {} \;
find /sys/class/egm -maxdepth 2 -type f -name gpu_devices -print -exec cat {} \;
```

For VFIO-based GPU passthrough, the host GPU compute driver must not own the GPUs that will be assigned to VMs. Remove or disable host NVIDIA driver packages that bind those devices, then bind the passthrough GPUs to the platform VFIO driver used by the Grace-Blackwell kernel stack:

```shell
lspci -nnk -d 10de:
modprobe vfio-pci
# Use the platform-specific VFIO driver and binding policy validated for the host.
```

Install the NVIDIA GPU Operator, GPU device plugin, or DRA driver according to the cluster integration model. For passthrough VMs, configure it to advertise the KubeVirt resource names and mount the required device nodes without taking ownership of the GPUs with a host compute driver. For EGM-backed compute tray-size VMs, the device integration must allocate the full intended GPU set and expose the selected `/dev/egm*` devices to the virt-launcher pod.

## EGM Sizing

EGM Virtualization (vEGM) on Grace-Blackwell is only supported at the compute-tray boundary, which means all the GPUs of a host (compute tray) must be passthrough to one single VM.
When EGM is enabled, the guest memory must equal the total EGM memory selected for passthrough. Read the size from sysfs:

```shell
total_mib=0
for dev in /sys/class/egm/egm*/egm_size; do
    raw=$(cat "$dev")
    bytes=$((16#${raw#0x}))
    mib=$((bytes / 1024 / 1024))
    total_mib=$((total_mib + mib))
    echo "$dev: ${mib}Mi"
done
echo "total: ${total_mib}Mi"
```

Example output:
```
/sys/class/egm/egm4: gpu_devices=0008:01:00.0
0009:01:00.0, egm_size=0x6fc0000000 (457728 MB)
/sys/class/egm/egm5: gpu_devices=0018:01:00.0
0019:01:00.0, egm_size=0x6fc0000000 (457728 MB)
```

Use the total as `spec.domain.memory.guest` in the VMI. For example, two EGM devices of `457728Mi` each require:

```yaml
domain:
  memory:
    guest: 915456Mi # total of all egm_size
```

Do not count the guest memory into pod memory request and limit for an EGM-backed VM. The launcher pod still needs ordinary Kubernetes memory for process overhead, but the guest RAM is backed by the EGM devices on the host.

## Kubernetes Node Configuration

Use static CPU management so KubeVirt can pin vCPUs.
For large Grace systems, keep TopologyManager enabled and use a Kubernetes build that contains the device-manager NUMA hint scalability fix. The public issue is [kubevirt/kubevirt#16289](https://github.com/kubevirt/kubevirt/issues/16289), which points to [kubernetes/kubernetes#135541](https://github.com/kubernetes/kubernetes/issues/135541).The Kubernetes fix was merged by [kubernetes/kubernetes#138244](https://github.com/kubernetes/kubernetes/pull/138244). Use Kubernetes 1.36 or a downstream build that contains that change.

```yaml
kubeletConfig:
  cpuManagerPolicy: static
  topologyManagerPolicy: best-effort
  topologyManagerScope: pod
  cpuManagerPolicyOptions:
    distribute-cpus-across-numa: "true"
  topologyManagerPolicyOptions:
    prefer-closest-numa-nodes: "true"
    max-allowable-numa-nodes: "36" # GB200 exposes 2 cpu NUMA nodes and additonal 34 memory-less NUMA nodes. Use the actual value by `lscpu`
```

After changing kubelet CPU manager policy, drain the node and remove stale CPUManager state according to Kubernetes operational guidance before restarting kubelet.

Large Grace systems can expose many NUMA nodes. Use the actual NUMA output by `lscpu` and `numactl -H`

## Configure containerd and kubelet for virt-launcher

Grace-Blackwell passthrough VMs can require locked memory and elevated resource limits inside the virt-launcher pod. The runtime configuration used by virt-launcher must allow `CAP_IPC_LOCK`, `CAP_SYS_RESOURCE`, and a large enough `RLIMIT_MEMLOCK` for the selected QEMU/libvirt flow.

Scope this to the runtime handler or runtime configuration used by KubeVirt virt-launcher. Do not apply it globally to unrelated pods. If the runtime configuration replaces a base OCI spec or capability list, preserve the distribution/runtime defaults and add the Grace requirements rather than reducing the container to only the capabilities shown here.

The exact file paths depend on the Kubernetes distribution. The fragment below shows only the OCI fields relevant to the Grace requirement:

```json
{
  "process": {
    "capabilities": {
      "bounding": ["CAP_IPC_LOCK", "CAP_SYS_RESOURCE"],
      "effective": ["CAP_IPC_LOCK", "CAP_SYS_RESOURCE"],
      "permitted": ["CAP_IPC_LOCK", "CAP_SYS_RESOURCE"]
    },
    "rlimits": [
      {
        "type": "RLIMIT_MEMLOCK",
        "hard": 18446744073709551615,
        "soft": 18446744073709551615
      }
    ]
  }
}
```

Some runtimes represent unlimited `RLIMIT_MEMLOCK` differently; use the value accepted by the target runtime and verify it from the launched pod.

Treat this as a host integration requirement, not as VMI API configuration. Validate the effective pod sandbox settings on the node before debugging QEMU or libvirt failures.

## GPU Device Plugin

Deploy a GPU device plugin or DRA driver that:

- Advertises the GPU resource name used by VMIs, for example `nvidia.com/GB100_HGX_GB200`.
- Allocates the intended GPUs to the virt-launcher pod.
- Mounts the required device nodes into virt-launcher, including the VFIO devices and, for EGM VMs, the selected `/dev/egm*` devices.
- Enforces the platform policy for compute tray-size VMs when `egm=true`.

The KubeVirt code in this branch validates Grace annotation semantics and builds the libvirt domain XML. It does not replace the device plugin or DRA layer that decides which host devices are allocated to a pod.

## KubeVirt Configuration

Configure the KubeVirt CR with the experimental feature gates used by this branch:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - HostDevices
      - GraceIOVirtualization
      - PCINUMAAwareTopology
    permittedHostDevices: 
      pciHostDevices: # Optional to gate permitted host devices, see details in KubeVirt GPU Device Plugin
      - pciVendorSelector: "10DE:*"
        resourceName: nvidia.com/GB100_HGX_GB200
```

Use a tighter `pciVendorSelector` in production once the exact device IDs are known.

## VMI Example

Start from:

- [examples/vmi-dsx-grace-blackwell-egm.yaml](../../examples/vmi-dsx-grace-blackwell-egm.yaml)

Update these fields for the target cluster:

- `metadata.name` and `metadata.namespace`
- `spec.domain.memory.guest`
- `spec.domain.devices.hostDevices[].deviceName`
- `spec.networks[].multus.networkName`
- `spec.volumes[].persistentVolumeClaim.claimName`
- `spec.nodeSelector`
- `alpha.kubevirt.io/pciHole64Size`

## Deployment Checklist

1. Build or obtain NVIDIA-patched aarch64 QEMU/libvirt RPMs.
2. Serve those RPMs from a local or internal HTTP repository.
3. Create a local `rpm/nvidia-repo.yaml` from `rpm/nvidia-repo.yaml.example`.
4. Run `make rpm-deps` with aarch64 version overrides.
5. Build and push KubeVirt images to a cluster-visible registry.
6. Deploy or update KubeVirt with the Grace feature gates and permitted host devices.
7. Deploy the GPU device plugin or DRA driver.
8. Validate node allocatable resources and mounted device nodes.
9. Create a VMI from the DSX example after setting the EGM guest memory.

## Troubleshooting

| Symptom | Check |
| :------ | :---- |
| Admission rejects `graceVirtualization` | Confirm `GraceIOVirtualization` feature gate, ARM64 architecture, valid JSON, and supported keys only. |
| `vcmdq=true` rejected | Set `smmuv3=true`; for non-EGM VMs also configure hugepages as required by the admission policy. |
| `egm=true` rejected | Set `smmuv3=true`, `dedicatedCpuPlacement=true`, explicit `domain.memory.guest`, and remove hugepages. |
| VM fails with PCI allocation errors | Increase `alpha.kubevirt.io/pciHole64Size` and confirm GPU BAR sizes. |
| virt-launcher cannot open EGM devices | Confirm the device plugin mounted the expected `/dev/egm*` nodes. |
| RPM dependency resolution fails | Confirm the NEVRA in `QEMU_VERSION_AARCH64` or `LIBVIRT_VERSION_AARCH64` exactly matches repository metadata. |
