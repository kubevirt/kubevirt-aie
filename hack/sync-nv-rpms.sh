#!/usr/bin/env bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright 2026 Red Hat, Inc.
#
# Syncs NV-variant RPMs (el10nv) for libvirt and qemu-kvm from the
# CentOS Stream 10 internal koji build system into WORKSPACE and
# rpm/BUILD.bazel.
#
# Usage:
#   hack/sync-nv-rpms.sh
#
# Environment variables (all optional):
#   KOJI_BASE_URL        - Base URL for koji packages directory
#   LIBVIRT_NV_VERSION   - Override libvirt version (default: auto-discover)
#   QEMU_NV_VERSION      - Override qemu-kvm version (default: auto-discover)
#   LIBVIRT_NV_RELEASE   - Override libvirt release (default: auto-discover latest el10nv)
#   QEMU_NV_RELEASE      - Override qemu-kvm release (default: auto-discover latest el10nv)

set -e
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORKSPACE_FILE="${REPO_DIR}/WORKSPACE"
BUILD_FILE="${REPO_DIR}/rpm/BUILD.bazel"

KOJI_BASE_URL="${KOJI_BASE_URL:-https://kojihub.stream.centos.org/kojifiles/packages}"
LIBVIRT_NV_VERSION="${LIBVIRT_NV_VERSION:-}"
QEMU_NV_VERSION="${QEMU_NV_VERSION:-}"
LIBVIRT_NV_RELEASE="${LIBVIRT_NV_RELEASE:-}"
QEMU_NV_RELEASE="${QEMU_NV_RELEASE:-}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

# Sub-packages to sync, with their supported architectures.
# Only packages/arches that already have el10 entries in WORKSPACE are included.
# Format: "subpkg:arch1,arch2"
LIBVIRT_SUBPKGS=(
    "libvirt-client:aarch64,x86_64"
    "libvirt-client-debuginfo:aarch64,x86_64"
    "libvirt-daemon-common:aarch64,x86_64"
    "libvirt-daemon-common-debuginfo:aarch64,x86_64"
    "libvirt-daemon-driver-qemu:aarch64,x86_64"
    "libvirt-daemon-driver-qemu-debuginfo:aarch64,x86_64"
    "libvirt-daemon-driver-secret:x86_64"
    "libvirt-daemon-driver-storage-core:x86_64"
    "libvirt-daemon-log:aarch64,x86_64"
    "libvirt-daemon-log-debuginfo:aarch64,x86_64"
    "libvirt-debuginfo:aarch64,x86_64"
    "libvirt-debugsource:aarch64,x86_64"
    "libvirt-devel:aarch64,x86_64"
    "libvirt-libs:aarch64,x86_64"
    "libvirt-libs-debuginfo:aarch64,x86_64"
)

QEMU_SUBPKGS=(
    "qemu-img:aarch64,x86_64"
    "qemu-img-debuginfo:aarch64,x86_64"
    "qemu-kvm-common:aarch64,x86_64"
    "qemu-kvm-common-debuginfo:aarch64,x86_64"
    "qemu-kvm-core:aarch64,x86_64"
    "qemu-kvm-core-debuginfo:aarch64,x86_64"
    "qemu-kvm-debugsource:aarch64,x86_64"
    "qemu-kvm-device-display-virtio-gpu:aarch64,x86_64"
    "qemu-kvm-device-display-virtio-gpu-debuginfo:aarch64,x86_64"
    "qemu-kvm-device-display-virtio-gpu-pci:aarch64,x86_64"
    "qemu-kvm-device-display-virtio-gpu-pci-debuginfo:aarch64,x86_64"
    "qemu-kvm-device-display-virtio-vga:x86_64"
    "qemu-kvm-device-display-virtio-vga-debuginfo:x86_64"
    "qemu-kvm-device-usb-host:aarch64,x86_64"
    "qemu-kvm-device-usb-host-debuginfo:aarch64,x86_64"
    "qemu-kvm-device-usb-redirect:aarch64,x86_64"
    "qemu-kvm-device-usb-redirect-debuginfo:aarch64,x86_64"
    "qemu-pr-helper:aarch64,x86_64"
)

# ---------------------------------------------------------------------------
# Functions
# ---------------------------------------------------------------------------

discover_latest_version() {
    local package=$1
    curl -sSL "${KOJI_BASE_URL}/${package}/" |
        grep -oP 'href="\K[0-9][^"]*(?=/")' |
        sort -V | tail -1
}

discover_latest_release() {
    local package=$1
    local version=$2
    curl -sSL "${KOJI_BASE_URL}/${package}/${version}/" |
        grep -oP 'href="\K[^"]*\.el10nv[^"]*(?=/")' |
        grep -v ',draft_' |
        sort -V | tail -1
}

get_epoch() {
    local source_pkg=$1
    local subpkg=$2
    local version=$3
    local release=$4
    local rpm_url="${KOJI_BASE_URL}/${source_pkg}/${version}/${release}/x86_64/${subpkg}-${version}-${release}.x86_64.rpm"
    local tmpfile="${TMPDIR}/epoch.rpm"

    curl -sSL -o "${tmpfile}" "${rpm_url}"
    local epoch
    epoch=$(rpm -qp --qf '%{EPOCH}' "${tmpfile}" 2>/dev/null)
    [[ "${epoch}" == "(none)" ]] && epoch=0
    rm -f "${tmpfile}"
    echo "${epoch}"
}

# ---------------------------------------------------------------------------
# 1. Resolve versions and releases
# ---------------------------------------------------------------------------

if [[ -z "${LIBVIRT_NV_VERSION}" ]]; then
    LIBVIRT_NV_VERSION=$(discover_latest_version libvirt)
fi
if [[ -z "${LIBVIRT_NV_RELEASE}" ]]; then
    LIBVIRT_NV_RELEASE=$(discover_latest_release libvirt "${LIBVIRT_NV_VERSION}")
fi
if [[ -z "${QEMU_NV_VERSION}" ]]; then
    QEMU_NV_VERSION=$(discover_latest_version qemu-kvm)
fi
if [[ -z "${QEMU_NV_RELEASE}" ]]; then
    QEMU_NV_RELEASE=$(discover_latest_release qemu-kvm "${QEMU_NV_VERSION}")
fi

echo "Syncing libvirt ${LIBVIRT_NV_VERSION}-${LIBVIRT_NV_RELEASE} (x86_64, aarch64)"
echo "Syncing qemu-kvm ${QEMU_NV_VERSION}-${QEMU_NV_RELEASE} (x86_64, aarch64)"

# ---------------------------------------------------------------------------
# 2. Get epochs (once per source package)
# ---------------------------------------------------------------------------

LIBVIRT_EPOCH=$(get_epoch libvirt libvirt-client "${LIBVIRT_NV_VERSION}" "${LIBVIRT_NV_RELEASE}")
QEMU_EPOCH=$(get_epoch qemu-kvm qemu-kvm-core "${QEMU_NV_VERSION}" "${QEMU_NV_RELEASE}")

# ---------------------------------------------------------------------------
# 3. Determine old el10 names for BUILD.bazel replacement and insertion points
# ---------------------------------------------------------------------------

declare -A OLD_NAMES # key: "subpkg:arch" -> old el10 bazel name
declare -A LAST_EL10 # key: subpkg -> last el10 name in WORKSPACE (insertion anchor)

find_old_names() {
    local subpkg=$1
    local arches=$2

    IFS=',' read -ra arch_list <<<"${arches}"
    for arch in "${arch_list[@]}"; do
        # Check for existing el10nv entries first (from a previous sync),
        # then fall back to base el10 entries (first-time sync).
        local old_name
        old_name=$(grep -oP "name = \"\K${subpkg}-\d+__[^\"]*\.el10nv[^\"]*\.${arch}" "${WORKSPACE_FILE}" | head -1 || true)
        if [[ -z "${old_name}" ]]; then
            old_name=$(grep -oP "name = \"\K${subpkg}-\d+__[^\"]*\.el10\.${arch}" "${WORKSPACE_FILE}" | head -1 || true)
        fi
        if [[ -n "${old_name}" ]]; then
            OLD_NAMES["${subpkg}:${arch}"]="${old_name}"
        fi
    done

    # Find the last el10 entry for this sub-package across ALL arches (for insertion point)
    local last
    last=$(grep -oP "name = \"\K${subpkg}-\d+__[^\"]*\.el10\.[a-z0-9_]+" "${WORKSPACE_FILE}" | sort | tail -1 || true)
    if [[ -n "${last}" ]]; then
        LAST_EL10["${subpkg}"]="${last}"
    fi
}

for entry in "${LIBVIRT_SUBPKGS[@]}"; do
    find_old_names "${entry%%:*}" "${entry##*:}"
done
for entry in "${QEMU_SUBPKGS[@]}"; do
    find_old_names "${entry%%:*}" "${entry##*:}"
done

# ---------------------------------------------------------------------------
# 4. Download RPMs and compute checksums
# ---------------------------------------------------------------------------

echo ""
echo "Downloading and hashing RPMs..."

declare -A SHA256S   # key: "subpkg:arch" -> sha256
declare -A NEW_NAMES # key: "subpkg:arch" -> new el10nv bazel name
declare -A RPM_URLS  # key: "subpkg:arch" -> koji RPM URL

download_subpackages() {
    local source_pkg=$1
    local version=$2
    local release=$3
    local epoch=$4
    shift 4
    local subpkgs=("$@")

    for entry in "${subpkgs[@]}"; do
        local subpkg="${entry%%:*}"
        local arches="${entry##*:}"

        IFS=',' read -ra arch_list <<<"${arches}"
        for arch in "${arch_list[@]}"; do
            local rpm_filename="${subpkg}-${version}-${release}.${arch}.rpm"
            local rpm_url="${KOJI_BASE_URL}/${source_pkg}/${version}/${release}/${arch}/${rpm_filename}"
            local bazel_name="${subpkg}-${epoch}__${version}-${release}.${arch}"

            local key="${subpkg}:${arch}"
            NEW_NAMES["${key}"]="${bazel_name}"

            # Skip download if the version/release hasn't changed.
            if [[ "${OLD_NAMES[${key}]:-}" == "${bazel_name}" ]]; then
                echo "  ${rpm_filename} (unchanged, skipping download)"
                continue
            fi

            local tmpfile="${TMPDIR}/${rpm_filename}"

            echo "  ${rpm_filename}"
            curl -sSL -o "${tmpfile}" "${rpm_url}"
            local sha256
            sha256=$(sha256sum "${tmpfile}" | awk '{print $1}')
            rm -f "${tmpfile}"

            SHA256S["${key}"]="${sha256}"
            RPM_URLS["${key}"]="${rpm_url}"
        done
    done
}

download_subpackages libvirt "${LIBVIRT_NV_VERSION}" "${LIBVIRT_NV_RELEASE}" "${LIBVIRT_EPOCH}" "${LIBVIRT_SUBPKGS[@]}"
download_subpackages qemu-kvm "${QEMU_NV_VERSION}" "${QEMU_NV_RELEASE}" "${QEMU_EPOCH}" "${QEMU_SUBPKGS[@]}"

# ---------------------------------------------------------------------------
# 5. Generate insertion files for WORKSPACE
# ---------------------------------------------------------------------------

# For each sub-package, write a file containing all its el10nv rpm() entries.
# These will be inserted after the last el10 entry for that sub-package.

declare -A INSERT_FILES # key: last_el10_name -> path to insertion file

all_subpkg_names=()

generate_insert_file() {
    local subpkgs=("$@")

    for entry in "${subpkgs[@]}"; do
        local subpkg="${entry%%:*}"
        local arches="${entry##*:}"

        # Skip subpackages whose version/release hasn't changed to avoid
        # removing existing cached builddeps URLs from their rpm() blocks.
        local changed=false
        IFS=',' read -ra check_arches <<<"${arches}"
        for arch in "${check_arches[@]}"; do
            local key="${subpkg}:${arch}"
            if [[ "${OLD_NAMES[${key}]:-}" != "${NEW_NAMES[${key}]:-}" ]]; then
                changed=true
                break
            fi
        done
        if ! ${changed}; then
            echo "  ${subpkg}: unchanged, skipping"
            continue
        fi

        all_subpkg_names+=("${subpkg}")

        local last_el10="${LAST_EL10[${subpkg}]:-}"
        if [[ -z "${last_el10}" ]]; then
            # Debug packages (debuginfo/debugsource) are NV-only and have no
            # base el10 entry to anchor against.  Append their rpm() entries
            # to the insert file of the corresponding non-debug parent package
            # so they are inserted alongside it.
            local parent_pkg="${subpkg%-debuginfo}"
            parent_pkg="${parent_pkg%-debugsource}"
            if [[ "${parent_pkg}" == "${subpkg}" ]]; then
                echo "WARNING: No el10 entry found for ${subpkg}, skipping" >&2
                continue
            fi
            local parent_anchor="${LAST_EL10[${parent_pkg}]:-}"
            # Fallback: find any managed sub-package matching the parent
            # prefix (e.g. libvirt-debuginfo -> libvirt-libs).
            if [[ -z "${parent_anchor}" ]]; then
                for pkg_key in "${!LAST_EL10[@]}"; do
                    if [[ "${pkg_key}" == "${parent_pkg}" || "${pkg_key}" == "${parent_pkg}-"* ]]; then
                        if [[ -z "${parent_anchor}" ]] || [[ "${LAST_EL10[${pkg_key}]}" > "${parent_anchor}" ]]; then
                            parent_anchor="${LAST_EL10[${pkg_key}]}"
                        fi
                    fi
                done
            fi
            if [[ -z "${parent_anchor}" ]]; then
                echo "WARNING: No el10 entry found for ${subpkg} or parent ${parent_pkg}, skipping" >&2
                continue
            fi
            local insert_file="${INSERT_FILES[${parent_anchor}]:-${TMPDIR}/insert_${parent_pkg}.txt}"
            INSERT_FILES["${parent_anchor}"]="${insert_file}"
            IFS=',' read -ra arch_list <<<"${arches}"
            for arch in "${arch_list[@]}"; do
                local key="${subpkg}:${arch}"
                if [[ -z "${NEW_NAMES[${key}]:-}" ]]; then
                    continue
                fi
                cat >>"${insert_file}" <<EOF

rpm(
    name = "${NEW_NAMES[${key}]}",
    sha256 = "${SHA256S[${key}]}",
    urls = [
        "${RPM_URLS[${key}]}",
    ],
)
EOF
            done
            continue
        fi

        # Reuse an existing insert file if a debug package already claimed
        # this anchor (e.g. libvirt-debuginfo processed before libvirt-libs).
        local insert_file="${INSERT_FILES[${last_el10}]:-${TMPDIR}/insert_${subpkg}.txt}"
        if [[ -z "${INSERT_FILES[${last_el10}]:-}" ]]; then
            : >"${insert_file}"
        fi
        local has_entries=false

        IFS=',' read -ra arch_list <<<"${arches}"
        for arch in "${arch_list[@]}"; do
            local key="${subpkg}:${arch}"
            if [[ -z "${NEW_NAMES[${key}]:-}" ]]; then
                continue
            fi
            cat >>"${insert_file}" <<EOF

rpm(
    name = "${NEW_NAMES[${key}]}",
    sha256 = "${SHA256S[${key}]}",
    urls = [
        "${RPM_URLS[${key}]}",
    ],
)
EOF
            has_entries=true
        done

        if ${has_entries}; then
            INSERT_FILES["${last_el10}"]="${insert_file}"
        fi
    done
}

generate_insert_file "${LIBVIRT_SUBPKGS[@]}"
generate_insert_file "${QEMU_SUBPKGS[@]}"

# ---------------------------------------------------------------------------
# 6. Update WORKSPACE
# ---------------------------------------------------------------------------

echo ""
echo "Updating WORKSPACE..."

# Write the insert map (last_el10_name|file_path) for awk
insert_map_file="${TMPDIR}/insert_map.txt"
: >"${insert_map_file}"
for name in "${!INSERT_FILES[@]}"; do
    echo "${name}|${INSERT_FILES[${name}]}" >>"${insert_map_file}"
done

# Write the sub-package name list for awk
subpkg_list_file="${TMPDIR}/subpkg_list.txt"
printf '%s\n' "${all_subpkg_names[@]}" >"${subpkg_list_file}"

# Single awk pass to remove stale el10nv blocks and insert new ones.
# See hack/sync-nv-rpms-update-workspace.awk for details.
awk -v insert_map_file="${insert_map_file}" \
    -v subpkg_list_file="${subpkg_list_file}" \
    -f "${SCRIPT_DIR}/sync-nv-rpms-update-workspace.awk" \
    "${WORKSPACE_FILE}" >"${TMPDIR}/WORKSPACE.new"

workspace_entries=$(grep -c 'name = ".*el10nv' "${TMPDIR}/WORKSPACE.new" || true)
mv "${TMPDIR}/WORKSPACE.new" "${WORKSPACE_FILE}"

echo "Updated WORKSPACE: ${workspace_entries} rpm() entries added/updated"

# ---------------------------------------------------------------------------
# 7. Update rpm/BUILD.bazel
# ---------------------------------------------------------------------------

echo ""
echo "Updating rpm/BUILD.bazel..."

build_count=0
for key in "${!OLD_NAMES[@]}"; do
    old="${OLD_NAMES[${key}]}"
    new="${NEW_NAMES[${key}]:-}"
    if [[ -z "${new}" ]]; then
        continue
    fi
    if grep -q "@${old}//rpm" "${BUILD_FILE}"; then
        old_escaped="${old//./\\.}"
        sed -i "s|@${old_escaped}//rpm|@${new}//rpm|g" "${BUILD_FILE}"
        count=$(grep -c "@${new}//rpm" "${BUILD_FILE}" || true)
        build_count=$((build_count + count))
    fi
done

echo "Updated rpm/BUILD.bazel: ${build_count} references updated"

# ---------------------------------------------------------------------------
# 8. Update sandbox hash in hack/bootstrap.sh
# ---------------------------------------------------------------------------

echo ""
echo "Updating sandbox hash..."

sandbox_hash=$(sha256sum "${BUILD_FILE}" | head -c 40)
sed -i "/^[[:blank:]]*sandbox_hash[[:blank:]]*=/s/=.*/=\"${sandbox_hash}\"/" "${SCRIPT_DIR}/bootstrap.sh"

echo "Updated sandbox hash: ${sandbox_hash}"
