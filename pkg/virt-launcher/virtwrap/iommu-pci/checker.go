//go:build !arm64

/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright The KubeVirt Authors.
 *
 */

package iommu_pci

func IommufdEnabled() (bool, error) {
	return false, nil
}

// This file provides stub implementations of IOMMU/PCI checker functions
// for non-ARM64 architectures (x86_64, s390x, etc.).
//
// On ARM64, these functions perform real IOMMU capability checks, NUMA
// configuration inference, and PCI memory hole calculations. On other
// architectures, these checks are either not applicable or not yet
// implemented, so we provide no-op stubs that return safe default values.
//
// This allows the same API to be used across all architectures without
// build errors, while the actual functionality is only active on ARM64
// where it's needed for NVIDIA GPU virtualization with vfio-pci/iommufd.

// parseConfigHybrid is a stub for non-ARM64 architectures.
// On ARM64, this function reads PCI extended configuration space to determine
// IOMMU capabilities (ATS, PASID, IOMMUFD support).
//
// Returns all capabilities as false/disabled since IOMMU checking is not
// implemented for this architecture.
func parseConfigHybrid(_ string) (atsSupported, atsEnabled, pasidSupported bool, ssidSize, oasBits int, err error) {
	return false, false, false, 0, 0, nil
}

// isSMMUv3Enabled is a stub for non-ARM64 architectures.
// On ARM64, this checks if the ARM SMMUv3 (System Memory Management Unit v3)
// is available by examining /sys/class/iommu.
//
// Returns false since SMMUv3 is an ARM-specific IOMMU implementation.
func isSMMUv3Enabled() (bool, error) {
	return false, nil
}

// CalculatePCIHole64Size is a stub for non-ARM64 architectures.
// On ARM64, this computes the total size of 64-bit PCI memory regions for
// a device by parsing /sys/bus/pci/devices/<bdf>/resource.
//
// Returns 0 since PCI hole size calculation is not implemented for this
// architecture.
func CalculatePCIHole64Size(_ string) (uint64, error) {
	return 0, nil
}

// CalculateTotalPCIHole64Size is a stub for non-ARM64 architectures.
// On ARM64, this computes the total PCI hole size with margin and rounds
// up to the next power of 2 for proper memory alignment.
//
// Returns 0 since this calculation is not applicable for this architecture.
func CalculateTotalPCIHole64Size(bdfComputedSize uint64, marginKiB uint64) uint64 {
	return 0
}

// InferExtraNUMANodes is a stub for non-ARM64 architectures.
// On ARM64, this determines the NUMA node configuration needed for proper
// IOMMU setup, inferring additional virtual NUMA nodes for devices with
// SMMUv3 and PASID support.
//
// Returns empty extraNodes list and -1 for mainNode since NUMA inference
// is not implemented for this architecture.
func InferExtraNUMANodes(_ string) (extraNodes []int, mainNode int, err error) {
	return []int{}, -1, nil
}
