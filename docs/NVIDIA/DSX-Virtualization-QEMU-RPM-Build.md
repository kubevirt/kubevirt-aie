# Build NVIDIA QEMU RPMs for DSX Virtualization

Prepared by: [Fan Zhang](https://github.com/fanzhangio), NVIDIA

This document describes a public reference flow for building aarch64 QEMU RPMs with NVIDIA Grace hardware-enablement patches and consuming them from this KubeVirt branch. The exact source branch, patch stack, packaging layout, and release number depend on the NVIDIA QEMU tree being validated.

The commands below are a reference flow. The exact spec path, source archive name, and `rpmbuild` macros depend on the packaging tree. If your source is distributed as an SRPM, prefer rebuilding from the SRPM and applying the NVIDIA
patch stack there.

## Why Custom QEMU RPMs Are Needed

Grace-Blackwell virtualization requires QEMU support that may not be present in the stock CentOS Stream 9 QEMU packages used by KubeVirt `release-1.7`. The DSX development flow keeps the upstream split QEMU package layout and overrides only the aarch64 package versions used by `rpm-deps`.

The expected package layout is:

- `qemu-img`
- `qemu-kvm-core`
- `qemu-kvm-device-*`
- `qemu-pr-helper`

Do not convert the build to a bundled QEMU package layout unless the KubeVirt RPM dependency lists are changed accordingly.

## Validated Rocky/RHEL 9 Recipe

This recipe captures the DSX validation flow for QEMU `10.1.0+nvidia5-1.el9`. It builds aarch64 RPMs in an emulated Rocky Linux 9 container on an x86_64 development host. The source tree is mounted at `/src`, and `/src` must be the QEMU repository root itself.

Set local paths on the host:

```shell
export WORKDIR=$HOME/dsx-rpms
export QEMU_SRC=$WORKDIR/nvidia-qemu
export OUTPUT_DIR=$WORKDIR/output

mkdir -p "$OUTPUT_DIR"
```

Clone the QEMU tree that carries the NVIDIA Grace enablement patches and check out the validated branch or tag:

```shell
git clone <qemu-source-url> "$QEMU_SRC"
cd "$QEMU_SRC"
git checkout nvidia_stable-10.1
```

Register ARM64 binfmt handlers and start the build container:

```shell
docker run --rm --privileged tonistiigi/binfmt --install arm64

docker run -it --rm --platform linux/arm64 \
  -v "$QEMU_SRC":/src \
  -v "$OUTPUT_DIR":/output \
  rockylinux:9 bash
```

Inside the container, install build tooling and import the Rocky/RHEL 9 `qemu-kvm` source RPM spec:

```shell
dnf install -y rpm-build rpmdevtools dnf-plugins-core git epel-release
dnf config-manager --set-enabled crb
dnf builddep -y qemu-kvm
dnf install -y python3-pip usbredir-devel
rpmdev-setuptree

# NVIDIA QEMU 10.1 requires a newer Meson than Rocky/RHEL 9 ships.
python3 -m pip install --upgrade 'meson>=1.5,<2'
hash -r
meson --version

dnf download --source qemu-kvm
rpm -ivh qemu-kvm-*.src.rpm
```

Create the source tarball expected by the patched spec:

```shell
cd /src
git config --global --add safe.directory /src
git submodule update --init --recursive

rm -f ~/rpmbuild/SOURCES/qemu-10.1.0+nvidia5.tar \
      ~/rpmbuild/SOURCES/qemu-10.1.0+nvidia5.tar.xz

tar -C /src --exclude='.git' \
  -cvf ~/rpmbuild/SOURCES/qemu-10.1.0+nvidia5.tar \
  --transform 's,^\./,qemu-10.1.0+nvidia5/,' .
xz -T0 -f ~/rpmbuild/SOURCES/qemu-10.1.0+nvidia5.tar

tar -tf ~/rpmbuild/SOURCES/qemu-10.1.0+nvidia5.tar.xz \
  | grep '^qemu-10.1.0+nvidia5/meson.build$'
```

Patch the source RPM spec for the NVIDIA QEMU source. These edits are intentionally explicit so reviewers can see why the resulting RPMs remain compatible with KubeVirt's split QEMU package layout:

```shell
SPEC=~/rpmbuild/SPECS/qemu-kvm.spec

sed -i 's/^Version:.*/Version: 10.1.0+nvidia5/' "$SPEC"
sed -i 's/^Release:.*/Release: 1%{?dist}/' "$SPEC"
sed -i 's|^Source0:.*|Source0: qemu-10.1.0+nvidia5.tar.xz|' "$SPEC"

# Disable downstream patches from the source RPM.
sed -i '/^Patch/s/^/#/' "$SPEC"
sed -i '/^%autopatch /d' "$SPEC"

# Ensure x86_64 builds still resolve %{kvm_target}.
grep -q '^    %global kvm_target    x86_64$' "$SPEC" || \
sed -i '/^%ifarch %{power64}/i %ifarch x86_64\n    %global kvm_target    x86_64\n%endif\n' "$SPEC"

# Remove RHEL-only options and unsupported old configure switches.
sed -i '/--with-devices-%{kvm_target}=%{kvm_target}-rh-devices/d' "$SPEC"
sed -i '/--rhel-version=/d' "$SPEC"
sed -i '/--disable-avx2/d' "$SPEC"
sed -i '/--disable-avx512bw/d' "$SPEC"
sed -i '/--disable-sanitizers/d' "$SPEC"

# Use upstream default devices and keep split module subpackages.
sed -i '/^  --with-devices-%{kvm_target}=default \\/d' "$SPEC"
sed -i '/^  --enable-attr \\/i\  --with-devices-%{kvm_target}=default \\' "$SPEC"
sed -i 's/--without-default-devices/--with-default-devices/' "$SPEC"

# Remove stale downstream packaging entries.
sed -i '/systemtap\/script.d/d' "$SPEC"
sed -i '/systemtap\/conf.d/d' "$SPEC"
sed -i 's/ README.systemtap//' "$SPEC"
sed -i '/^[[:space:]]*%{_libdir}\/%{name}\/accel-tcg-%{kvm_target}\.so$/d' "$SPEC"

# QEMU 10.1 renamed tests/avocado to tests/functional.
sed -i 's|tests/avocado|tests/functional|g' "$SPEC"
sed -i 's|avocado_qemu|functional|g' "$SPEC"
sed -i 's|cp -R %{qemu_kvm_build}/tests/functional/|cp -R tests/functional/|' "$SPEC"
sed -i 's|cp -R %{qemu_kvm_build}/python/|cp -R python/|' "$SPEC"
sed -i 's|cp -R %{qemu_kvm_build}/scripts/qmp/|cp -R scripts/qmp/|' "$SPEC"

# Restrict %check to stable suites if tests are enabled.
sed -i '/MTESTARGS/!s|make V=1 check|MTESTARGS="--suite qtest --suite softfloat --suite qapi-schema" make V=1 check|' "$SPEC"
sed -i 's|meson test -C %{qemu_kvm_build} --print-errorlogs --no-suite block --no-suite block-slow --no-suite block-thorough|MTESTARGS="--suite qtest --suite softfloat --suite qapi-schema" make V=1 check|' "$SPEC"

# Ensure meson subprojects are available during %build.
grep -q '^meson subprojects download$' "$SPEC" || \
sed -i '/^%build$/a meson subprojects download' "$SPEC"
```

Remove files produced by NVIDIA QEMU that the downstream spec does not own, and keep x86 firmware blobs inside x86-only file sections:

```shell
grep -q 'hw-uefi-vars.so' "$SPEC" || \
perl -0pi -e 's@rm -rf %\{buildroot\}%\{_datadir\}/%\{name\}/qboot\.rom\n@${&}rm -rf %{buildroot}%{_datadir}/%{name}/dtb/bamboo.dtb\nrm -rf %{buildroot}%{_datadir}/%{name}/dtb/canyonlands.dtb\nrm -rf %{buildroot}%{_datadir}/%{name}/dtb/petalogix-ml605.dtb\nrm -rf %{buildroot}%{_datadir}/%{name}/dtb/petalogix-s3adsp1800.dtb\nrm -rf %{buildroot}%{_datadir}/%{name}/ast27x0_bootrom.bin\nrm -rf %{buildroot}%{_datadir}/%{name}/npcm8xx_bootrom.bin\nrm -rf %{buildroot}%{_datadir}/%{name}/pnv-pnor.bin\nrm -rf %{buildroot}%{_libdir}/%{name}/hw-uefi-vars.so\nrm -f  %{buildroot}%{_libdir}/%{name}/accel-tcg-%{kvm_target}.so\n@' "$SPEC"

sed -i '/^[[:space:]]*%{_datadir}\/%{name}\/kvmvapic\.bin$/d' "$SPEC"
sed -i '/^[[:space:]]*%{_datadir}\/%{name}\/linuxboot\.bin$/d' "$SPEC"
sed -i '/^[[:space:]]*%{_datadir}\/%{name}\/multiboot\.bin$/d' "$SPEC"
sed -i '/^[[:space:]]*%{_datadir}\/%{name}\/multiboot_dma\.bin$/d' "$SPEC"
sed -i '/^[[:space:]]*%{_datadir}\/%{name}\/pvh\.bin$/d' "$SPEC"

perl -0pi -e '
  s/^%ifarch x86_64\n%endif\n//gm;
  s/^(%\{_datadir\}\/icons\/\*)$/$1\n%ifarch x86_64\n    %{_datadir}\/%{name}\/kvmvapic.bin\n    %{_datadir}\/%{name}\/linuxboot.bin\n    %{_datadir}\/%{name}\/multiboot.bin\n    %{_datadir}\/%{name}\/multiboot_dma.bin\n    %{_datadir}\/%{name}\/pvh.bin\n%endif/m;
' "$SPEC"
```

Verify the important spec changes:

```shell
sed -n '1,220p' "$SPEC" | grep -E '^(Version:|Release:|Source0:)'
grep -n -- '^  --with-devices-%{kvm_target}=default \\' "$SPEC"
grep -nE 'rh-devices|rhel-version|disable-avx2|disable-avx512bw|disable-sanitizers' "$SPEC" \
  | grep -v '^#' | grep -v '%changelog' || true

awk '/^%ifarch x86_64/{inside=1} /^%endif/{inside=0} /(kvmvapic|linuxboot|multiboot|multiboot_dma|pvh)\.bin/{printf "%s line %d: %s\n", (inside?"OK ":"BAD"), NR, $0}' "$SPEC"
```

Build the RPMs. The container is already the target architecture, so do not use
`--target=aarch64` from an x86_64 host:

```shell
dnf builddep -y "$SPEC"
rpmbuild -ba --nocheck "$SPEC"

cp ~/rpmbuild/RPMS/*/*.rpm /output/
cp "$SPEC" /output/
```

`--nocheck` is used in this validated emulated-container flow because some QEMU test reference data and qtests do not match the modified default device set or time out under user-mode emulation. Drop `--nocheck` only if you intend to fix or update the affected test data.

Expected aarch64 RPMs include:

```text
qemu-img-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-core-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-common-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-pci-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-device-usb-host-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-kvm-device-usb-redirect-10.1.0+nvidia5-1.el9.aarch64.rpm
qemu-pr-helper-10.1.0+nvidia5-1.el9.aarch64.rpm
```

## Publish a Local RPM Repository

Create a repository directory and generate metadata:

```shell
export WORKDIR=$HOME/dsx-rpms
export REPO_ROOT=$WORKDIR/repo/qemu/10.1.0-nvidia5
mkdir -p "$REPO_ROOT"
cp "$WORKDIR"/output/qemu*.aarch64.rpm "$REPO_ROOT"/
createrepo_c "$REPO_ROOT"
python3 -m http.server 8080 --directory "$WORKDIR/repo"
```

The resulting base URL is similar to:

```text
http://<repo-host>:8080/qemu/10.1.0-nvidia5/
```

## Consume From KubeVirt

Create a local `rpm/nvidia-repo.yaml` that points to the repository:

```yaml
repositories:
- arch: aarch64
  baseurl: http://<repo-host>:8080/qemu/10.1.0-nvidia5/
  name: custom-qemu-aarch64
  gpgcheck: 0
  repo_gpgcheck: 0
```

The same `rpm/nvidia-repo.yaml` can contain both the QEMU and libvirt custom repositories. `CUSTOM_REPO` is passed to `bazeldnf` before the default `rpm/repo.yaml`, so matching packages from the custom repository take precedence over the CentOS Stream defaults.

Run `rpm-deps` with the exact QEMU NEVRA:

```shell
make CUSTOM_REPO=rpm/nvidia-repo.yaml \
     QEMU_VERSION_AARCH64=17:10.1.0+nvidia5-1.el9 \
     LIBVIRT_VERSION_AARCH64=0:11.9.0+nvidia4-1.el9 \
     SINGLE_ARCH=aarch64 \
     rpm-deps
```

In most Grace-Blackwell builds you must also provide the matching libvirt override. See [Build NVIDIA libvirt RPMs](DSX-Virtualization-Libvirt-RPM-Build.md).

`make rpm-deps` updates `WORKSPACE` with pinned RPM URLs and checksums and updates `rpm/BUILD.bazel` with the regenerated `rpmtree` targets. Review both files before committing. A normal validation loop is:

```shell
make test
bazel build //cmd/virt-launcher:virt-launcher-image
bazel build //cmd/virt-handler:virt-handler
```

## Verification

Before running `rpm-deps`, verify the repository metadata contains the expected package names and versions:

```shell
dnf repoquery \
  --repofrompath custom-qemu-aarch64,http://<repo-host>:8080/qemu/10.1.0-nvidia5/ \
  --repo custom-qemu-aarch64 \
  --arch aarch64 \
  'qemu-img*' 'qemu-kvm-core*'
```

If `bazeldnf` reports that `qemu-img-17:10.1.0+nvidia5-1.el9` does not exist,
the version string passed to `QEMU_VERSION_AARCH64` does not match the RPM
repository metadata exactly.

## Open-Source Hygiene

- Commit only generic examples such as `rpm/nvidia-repo.yaml.example`.
- Do not commit internal repository URLs, credentials, hostnames, or registry names.
- Keep aarch64 overrides separate from x86_64 and s390x defaults.
- Keep the split QEMU package layout unless the RPM dependency rules are updated in the same change.
