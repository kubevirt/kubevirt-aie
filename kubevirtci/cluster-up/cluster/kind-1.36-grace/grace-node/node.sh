#!/bin/bash

function node::remount_sysfs() {
  local -r nodes_array=($1)
  for node in "${nodes_array[@]}"; do
    ${CRI_BIN} exec $node mount -o remount,rw /sys
  done
}

function node::setup_vfio() {
  local -r nodes_array=($1)
  for node in "${nodes_array[@]}"; do
    ${CRI_BIN} exec $node chmod 666 /dev/iommu 2>/dev/null || true
    ${CRI_BIN} exec $node chmod 666 /dev/vfio/vfio 2>/dev/null || true
    ${CRI_BIN} exec $node bash -c 'for f in /dev/vfio/*; do chmod 666 "$f" 2>/dev/null; done' || true
    ${CRI_BIN} exec $node bash -c 'for f in /dev/vfio/devices/*; do chmod 666 "$f" 2>/dev/null; done' || true
  done
}

function node::deploy_gpu_device_plugin() {
  _kubectl apply -f - <<'EOF'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kubevirt-gpu-device-plugin
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: kubevirt-gpu-device-plugin
  template:
    metadata:
      labels:
        name: kubevirt-gpu-device-plugin
    spec:
      tolerations:
      - key: nvidia.com/gpu
        operator: Exists
        effect: NoSchedule
      containers:
      - name: kubevirt-gpu-device-plugin
        image: ghcr.io/nvidia/kubevirt-gpu-device-plugin:94985728
        securityContext:
          privileged: true
        volumeMounts:
        - name: device-plugin
          mountPath: /var/lib/kubelet/device-plugins
        - name: pci-devices
          mountPath: /sys/bus/pci/devices
          readOnly: true
        - name: dev-vfio
          mountPath: /dev/vfio
      volumes:
      - name: device-plugin
        hostPath:
          path: /var/lib/kubelet/device-plugins
      - name: pci-devices
        hostPath:
          path: /sys/bus/pci/devices
      - name: dev-vfio
        hostPath:
          path: /dev/vfio
EOF

  echo "Waiting for kubevirt-gpu-device-plugin to be ready..."
  _kubectl -n kube-system rollout status daemonset/kubevirt-gpu-device-plugin --timeout=120s || true

  echo "GPU extended resources:"
  _kubectl get nodes -o json | jq '.items[].status.allocatable | to_entries[] | select(.key | startswith("nvidia.com/"))' || true
}
