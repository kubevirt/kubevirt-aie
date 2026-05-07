# Build NVIDIA libvirt RPMs for DSX Virtualization

Prepared by: [Fan Zhang](https://github.com/fanzhangio), NVIDIA

This document describes a public reference flow for building aarch64 libvirt RPMs with NVIDIA Grace hardware-enablement patches and consuming them from this KubeVirt branch. The exact source branch, patch stack, packaging layout, and release number depend on the NVIDIA libvirt tree being validated.

The commands below are a reference flow. The exact spec path, source archive name, and `rpmbuild` macros depend on the packaging tree. If your source is distributed as an SRPM, prefer rebuilding from the SRPM and applying the NVIDIA patch stack there.

## Why Custom libvirt RPMs Are Needed

The Grace-Blackwell flow requires libvirt support for ARM64 passthrough features used by the generated domain XML, including SMMUv3, IOMMUFD-related device wiring, and Grace-specific topology handling. KubeVirt consumes those changes through aarch64 RPM version overrides while leaving other architectures on the standard upstream versions.

## Validated Rocky/RHEL 9 Recipe

This recipe captures the DSX validation flow for libvirt `11.9.0+nvidia4-1.el9`. It builds aarch64 RPMs in an emulated Rocky Linux 9 container on an x86_64 development host. The source tree is mounted at `/src`, and `/src` must be the libvirt repository root itself.

Set local paths on the host and initialize submodules before entering the container:

```shell
export WORKDIR=$HOME/dsx-rpms
export LIBVIRT_SRC=$WORKDIR/nvidia-libvirt
export OUTPUT_DIR=$WORKDIR/output

git clone <libvirt-source-url> "$LIBVIRT_SRC"
cd "$LIBVIRT_SRC"
git checkout nvidia_stable-11.9
git submodule update --init --recursive

mkdir -p "$OUTPUT_DIR"
```

Register ARM64 binfmt handlers and start the build container:

```shell
docker run --rm --privileged tonistiigi/binfmt --install arm64

docker run -it --rm --platform linux/arm64 \
  --name build-libvirt \
  -v "$LIBVIRT_SRC":/src \
  -v "$OUTPUT_DIR":/output \
  rockylinux:9 bash
```

Inside the container, install build tooling and import the Rocky/RHEL 9 `libvirt` source RPM spec:

```shell
dnf install -y rpm-build rpmdevtools dnf-plugins-core git epel-release
dnf config-manager --set-enabled crb
dnf builddep -y libvirt
rpmdev-setuptree

dnf download --source libvirt
rpm -ivh libvirt-*.src.rpm
```

Patch the source RPM spec for the NVIDIA libvirt source:

```shell
SPEC=~/rpmbuild/SPECS/libvirt.spec

sed -i 's/^Version:.*/Version: 11.9.0+nvidia4/' "$SPEC"
sed -i 's/^Release:.*/Release: 1%{?dist}/' "$SPEC"

# Disable downstream patches. Replace the whole autosetup line to avoid a
# double -N flag when the source RPM already uses git_am.
sed -i 's/^Patch/#Patch/' "$SPEC"
sed -i 's/%autosetup -S git_am -N/%autosetup -N/' "$SPEC"

grep -E '^(Version:|Release:)' "$SPEC"
sed -n '/%prep/,/%build/p' "$SPEC" | head -6
```

Create the source tarball expected by the patched spec. The `AUTHORS.rst` file is required by the libvirt build when using this generated archive:

```shell
cd /src
echo "See git history for contributors" > AUTHORS.rst
cd /

tar --exclude='.git' \
  -cvf ~/rpmbuild/SOURCES/libvirt-11.9.0+nvidia4.tar \
  --transform 's,^src,libvirt-11.9.0+nvidia4,' src
xz -T0 -f ~/rpmbuild/SOURCES/libvirt-11.9.0+nvidia4.tar

tar -tf ~/rpmbuild/SOURCES/libvirt-11.9.0+nvidia4.tar.xz \
  | grep 'keycodemapdb/meson.build'
```

Build the RPMs and copy the output. The container is already the target architecture, so do not use `--target=aarch64` from an x86_64 host:

```shell
dnf builddep -y "$SPEC"
rpmbuild -ba "$SPEC"

cp ~/rpmbuild/RPMS/*/*.rpm /output/
cp "$SPEC" /output/
```

Expected aarch64 RPMs include the packages consumed by KubeVirt:

```text
libvirt-client-11.9.0+nvidia4-1.el9.aarch64.rpm
libvirt-daemon-driver-qemu-11.9.0+nvidia4-1.el9.aarch64.rpm
libvirt-devel-11.9.0+nvidia4-1.el9.aarch64.rpm
```

## Publish a Local RPM Repository

Create repository metadata:

```shell
export WORKDIR=$HOME/dsx-rpms
export REPO_ROOT=$WORKDIR/repo/libvirt/11.9.0-nvidia4
mkdir -p "$REPO_ROOT"
cp "$WORKDIR"/output/libvirt*.aarch64.rpm "$REPO_ROOT"/
createrepo_c "$REPO_ROOT"
python3 -m http.server 8080 --directory "$WORKDIR/repo"
```

The resulting base URL is similar to:

```text
http://<repo-host>:8080/libvirt/11.9.0-nvidia4/
```

## Consume From KubeVirt

Create a local `rpm/nvidia-repo.yaml` with the libvirt repository:

```yaml
repositories:
- arch: aarch64
  baseurl: http://<repo-host>:8080/libvirt/11.9.0-nvidia4/
  name: custom-libvirt-aarch64
  gpgcheck: 0
  repo_gpgcheck: 0
```

The same `rpm/nvidia-repo.yaml` can contain both the libvirt and QEMU custom repositories. `CUSTOM_REPO` is passed to `bazeldnf` before the default `rpm/repo.yaml`, so matching packages from the custom repository take precedence over the CentOS Stream defaults.

Run `rpm-deps` with the exact libvirt NEVRA:

```shell
make CUSTOM_REPO=rpm/nvidia-repo.yaml \
     QEMU_VERSION_AARCH64=17:10.1.0+nvidia5-1.el9 \
     LIBVIRT_VERSION_AARCH64=0:11.9.0+nvidia4-1.el9 \
     SINGLE_ARCH=aarch64 \
     rpm-deps
```

In most Grace-Blackwell builds you must also provide the matching QEMU override. See [Build NVIDIA QEMU RPMs](DSX-Virtualization-QEMU-RPM-Build.md).

`make rpm-deps` updates `WORKSPACE` with pinned RPM URLs and checksums and updates `rpm/BUILD.bazel` with the regenerated `rpmtree` targets. Review both files before committing. A normal validation loop is:

```shell
make test
bazel build //cmd/virt-launcher:virt-launcher-image
bazel build //cmd/virt-handler:virt-handler
```

## Verification

Verify repository metadata before running `rpm-deps`:

```shell
dnf repoquery \
  --repofrompath custom-libvirt-aarch64,http://<repo-host>:8080/libvirt/11.9.0-nvidia4/ \
  --repo custom-libvirt-aarch64 \
  --arch aarch64 \
  'libvirt-client*' 'libvirt-daemon-driver-qemu*' 'libvirt-devel*'
```

If `bazeldnf` cannot find `libvirt-client-0:11.9.0+nvidia4-1.el9` or
`libvirt-daemon-driver-qemu-0:11.9.0+nvidia4-1.el9`, the NEVRA passed to
`LIBVIRT_VERSION_AARCH64` does not match repository metadata exactly.

## Open-Source Hygiene

- Commit only placeholders and examples.
- Keep real `rpm/nvidia-repo.yaml` files local or in private deployment repos.
- Do not commit internal repository URLs, hostnames, registry names, or credentials.
- Keep the aarch64 override scoped to Grace hardware-enablement builds.
