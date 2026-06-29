#!/bin/bash

[ $(id -u) -ne 0 ] && echo "FATAL: this script requires sudo privileges" >&2 && exit 1

set -xe

SCRIPT_PATH=$(dirname "$(realpath "$0")")

source ${SCRIPT_PATH}/grace-node/node.sh
source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

nodes=($(_kubectl get nodes -o custom-columns=:.metadata.name --no-headers))
node::remount_sysfs "${nodes[*]}"
node::setup_vfio "${nodes[*]}"
node::deploy_gpu_device_plugin

_kubectl get nodes -o wide
