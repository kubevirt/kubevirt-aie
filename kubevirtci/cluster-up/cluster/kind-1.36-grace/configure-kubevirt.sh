#!/bin/bash

set -e

SCRIPT_PATH=$(dirname "$(realpath "$0")")
source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

echo "Configuring KubeVirt for Grace GPU passthrough..."

# Wait for KubeVirt to be available
_kubectl wait -n kubevirt kv kubevirt --for condition=Available --timeout 5m

# Discover GPU PCI IDs from the host (3D controllers only)
gpu_pci_ids=$(lspci -d 10de: -nn | grep -i '3D controller' | grep -oP '\[10de:\K[0-9a-f]+(?=\])' | sort -u)

if [ -z "$gpu_pci_ids" ]; then
  echo "WARNING: No NVIDIA 3D controller GPUs found, skipping permittedHostDevices"
  exit 0
fi

# Discover resource names from node allocatable
gpu_resources=$(_kubectl get nodes -o json | jq -r '.items[].status.allocatable | to_entries[] | select(.key | startswith("nvidia.com/")) | .key' | sort -u)

if [ -z "$gpu_resources" ]; then
  echo "WARNING: No nvidia.com/* resources found on nodes, skipping permittedHostDevices"
  exit 0
fi

# Build the permittedHostDevices patch
pci_host_devices="["
first=true
for resource in $gpu_resources; do
  for pci_id in $gpu_pci_ids; do
    if [ "$first" = true ]; then
      first=false
    else
      pci_host_devices+=","
    fi
    pci_host_devices+="{\"pciVendorSelector\":\"10de:${pci_id}\",\"resourceName\":\"${resource}\"}"
  done
done
pci_host_devices+="]"

echo "Patching KubeVirt CR with permittedHostDevices:"
echo "  PCI IDs: $(echo $gpu_pci_ids | tr '\n' ' ')"
echo "  Resources: $(echo $gpu_resources | tr '\n' ' ')"

_kubectl patch kubevirt kubevirt -n kubevirt \
  --type merge -p "{\"spec\":{\"configuration\":{\"permittedHostDevices\":{\"pciHostDevices\":${pci_host_devices}}}}}"

echo "KubeVirt configured for Grace GPU passthrough"
