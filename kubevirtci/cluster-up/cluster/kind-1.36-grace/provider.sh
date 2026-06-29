#!/usr/bin/env bash

set -e

DEFAULT_CLUSTER_NAME="kind-1.36-grace"
DEFAULT_HOST_PORT=5000
ALTERNATE_HOST_PORT=5001
export CLUSTER_NAME=${CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}

if [ $CLUSTER_NAME == $DEFAULT_CLUSTER_NAME ]; then
  export HOST_PORT=$DEFAULT_HOST_PORT
else
  export HOST_PORT=$ALTERNATE_HOST_PORT
fi

function set_kind_params() {
  version=$(cat "${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/version")
  export KIND_VERSION="${KIND_VERSION:-$version}"

  image=$(cat "${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/image")
  export KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-$image}"
}

function configure_registry_proxy() {
  [ "$CI" != "true" ] && return

  echo "Configuring cluster nodes to work with CI mirror-proxy..."

  local -r ci_proxy_hostname="docker-mirror-proxy.kubevirt-prow.svc"
  local -r kind_binary_path="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kind"
  local -r configure_registry_proxy_script="${KUBEVIRTCI_PATH}/cluster/kind/configure-registry-proxy.sh"

  KIND_BIN="$kind_binary_path" PROXY_HOSTNAME="$ci_proxy_hostname" $configure_registry_proxy_script
}

function _add_grace_mounts() {
  local kind_config="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/kind.yaml"

  if [ -e /dev/iommu ]; then
    cat <<EOF >> "$kind_config"
  - containerPath: /dev/iommu
    hostPath: /dev/iommu
EOF
  fi

  if [ -d /dev/vfio ]; then
    cat <<EOF >> "$kind_config"
  - containerPath: /dev/vfio
    hostPath: /dev/vfio
EOF
  fi

  cat <<EOF >> "$kind_config"
  - containerPath: /sys/bus/pci/devices
    hostPath: /sys/bus/pci/devices
  - containerPath: /sys/class/iommu
    hostPath: /sys/class/iommu
    readOnly: true
  - containerPath: /sys/kernel/iommu_groups
    hostPath: /sys/kernel/iommu_groups
    readOnly: true
  - containerPath: /sys/devices
    hostPath: /sys/devices
    readOnly: true
EOF
}

function _add_grace_topology_manager_options() {
  local kind_config="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/kind.yaml"

  # CONFIG_TOPOLOGY_MANAGER_POLICY=restricted causes _add_kubeadm_config_patches
  # in common.sh to write a KubeletConfiguration block with the policy. We inject
  # the policy options into that same block by replacing the trailing --- separator.
  sed -i '/^  topologyManagerPolicy: restricted$/a\  topologyManagerPolicyOptions:\n    max-allowable-numa-nodes: "64"\n    prefer-closest-numa-nodes: "true"' "$kind_config"
}

function _load_grace_modules() {
  modprobe -q iommufd || true
  # nvgrace-gpu-vfio-pci handles Grace-specific VFIO reset and C2C coherent
  # link setup. Using generic vfio-pci causes kbusVerifyCoherentLink crashes
  # in the NVIDIA guest driver (VOYAGER-1197).
  modprobe -q nvgrace-gpu-vfio-pci || true
}

function _bind_gpus_to_vfio() {
  local gpu_bdfs
  gpu_bdfs=$(lspci -d 10de: -D | grep -i '3D controller' | awk '{print $1}')

  if [ -z "$gpu_bdfs" ]; then
    echo "WARNING: No NVIDIA 3D controller GPUs found"
    return
  fi

  # Bind via new_id rather than driver_override — nvgrace-gpu-vfio-pci
  # cannot reclaim devices from vfio-pci without a full reboot.
  echo "10de 2941" > /sys/bus/pci/drivers/nvgrace_gpu_vfio_pci/new_id 2>/dev/null || true
  sleep 2

  for bdf in $gpu_bdfs; do
    local current_driver
    current_driver=$(basename "$(readlink /sys/bus/pci/devices/$bdf/driver 2>/dev/null)" 2>/dev/null || true)
    echo "$bdf bound to: ${current_driver:-UNBOUND}"
  done

  # VFIO cdev devices are created with 0600 permissions but virt-launcher
  # runs as non-root (UID 107) and needs to open them.
  chmod 666 /dev/vfio/devices/* 2>/dev/null || true

  echo "VFIO cdevs: $(ls /dev/vfio/devices/ 2>/dev/null)"
}

function up() {
  _load_grace_modules
  _bind_gpus_to_vfio

  echo 'Discovering NVIDIA GPUs for Grace passthrough...'
  lspci -d 10de: -nn || true
  echo ""

  # CPU manager requires a worker node — the JoinConfiguration kubelet patch
  # only applies to workers, not the control-plane. Override the default of 1
  # unless the user explicitly set KUBEVIRT_NUM_NODES > 1.
  if [ "${KUBEVIRT_NUM_NODES}" -le 1 ] 2>/dev/null; then
    export KUBEVIRT_NUM_NODES=2
  fi

  cp $KIND_MANIFESTS_DIR/kind.yaml ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/kind.yaml
  _add_kubeadm_cpu_manager_config_patch
  _add_extra_mounts
  _add_grace_mounts
  _add_extra_portmapping
  export CONFIG_WORKER_CPU_MANAGER=true
  export CONFIG_TOPOLOGY_MANAGER_POLICY=restricted

  _fetch_kind
  _prepare_kind_config
  _add_grace_topology_manager_options
  setup_kind

  configure_registry_proxy

  ${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/config_grace_cluster.sh

  echo "$KUBEVIRT_PROVIDER cluster '$CLUSTER_NAME' is ready"
}

set_kind_params

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh
