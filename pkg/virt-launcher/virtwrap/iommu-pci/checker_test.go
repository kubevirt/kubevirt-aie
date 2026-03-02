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
 * Copyright 2021
 *
 */

package iommu_pci_test

import (
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	iommupci "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/iommu-pci"
)

// isARM64 returns true if running on ARM64 architecture
func isARM64() bool {
	return runtime.GOARCH == "arm64"
}

var _ = Describe("IOMMU PCI Checker", func() {

	Describe("CalculateTotalPCIHole64Size", func() {
		Context("when computing PCI hole sizes", func() {
			It("should return 0 when both size and margin are 0", func() {
				result := iommupci.CalculateTotalPCIHole64Size(0, 0)
				Expect(result).To(Equal(uint64(0)))
			})

			It("should round up to nearest power of 2", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// 1024 bytes (1 KiB) + 0 margin = 1 KiB, already power of 2
				result := iommupci.CalculateTotalPCIHole64Size(1024, 0)
				Expect(result).To(Equal(uint64(1)))

				// 1024 bytes (1 KiB) + 1 KiB margin = 2 KiB, already power of 2
				result = iommupci.CalculateTotalPCIHole64Size(1024, 1)
				Expect(result).To(Equal(uint64(2)))

				// 1536 bytes (1.5 KiB) + 0 margin = rounds up to 2 KiB
				result = iommupci.CalculateTotalPCIHole64Size(1536, 0)
				Expect(result).To(Equal(uint64(2)))

				// 3 MiB (3072 KiB) should round up to 4096 KiB (4 MiB)
				result = iommupci.CalculateTotalPCIHole64Size(3*1024*1024, 0)
				Expect(result).To(Equal(uint64(4096)))
			})

			It("should add margin before rounding", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// 1024 bytes (1 KiB) + 1024 KiB margin = 1025 KiB, rounds to 2048 KiB
				result := iommupci.CalculateTotalPCIHole64Size(1024, 1024)
				Expect(result).To(Equal(uint64(2048)))

				// 512 KiB device + 512 KiB margin = 1024 KiB, already power of 2
				result = iommupci.CalculateTotalPCIHole64Size(512*1024, 512)
				Expect(result).To(Equal(uint64(1024)))
			})

			It("should handle large GPU memory sizes", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// Simulate a 16 GB GPU BAR (typical for modern GPUs)
				// 16 GB = 16 * 1024 * 1024 KiB = 16777216 KiB
				deviceSize := uint64(16 * 1024 * 1024 * 1024) // 16 GB in bytes
				margin := uint64(1024 * 1024)                 // 1 GB margin in KiB

				result := iommupci.CalculateTotalPCIHole64Size(deviceSize, margin)

				// 16 GB = 16777216 KiB, + 1048576 KiB = 17825792 KiB
				// Next power of 2 is 2^25 = 33554432 KiB (32 GB)
				Expect(result).To(Equal(uint64(33554432)))
			})

			It("should handle multiple small devices", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// Simulate 4 devices with 256 MB each = 1 GB total
				totalSize := uint64(4 * 256 * 1024 * 1024) // 1 GB in bytes
				margin := uint64(0)

				result := iommupci.CalculateTotalPCIHole64Size(totalSize, margin)

				// 1 GB = 1048576 KiB, already power of 2
				Expect(result).To(Equal(uint64(1048576)))
			})
		})
	})

	Describe("NewBDFDevice", func() {
		Context("when creating a BDF device", func() {
			It("should initialize with the correct ID", func() {
				bdf := iommupci.NewBDFDevice("0000:3b:00.0")
				Expect(bdf).NotTo(BeNil())
				Expect(bdf.ID).To(Equal("0000:3b:00.0"))
			})

			It("should handle different BDF formats", func() {
				testCases := []string{
					"0000:00:00.0",
					"0000:ff:1f.7",
					"0001:3b:00.0",
					"ffff:ff:ff.f",
				}

				for _, testCase := range testCases {
					bdf := iommupci.NewBDFDevice(testCase)
					Expect(bdf).NotTo(BeNil())
					Expect(bdf.ID).To(Equal(testCase))
				}
			})
		})
	})

	Describe("NewIommuPCI", func() {
		Context("when initializing IOMMU checker", func() {
			It("should create instance for arm64", func() {
				iommu := iommupci.NewIommuPCI("arm64")
				Expect(iommu).NotTo(BeNil())
				Expect(iommu.PCIHoleSize).To(Equal(uint64(0)))
			})

			It("should create instance for x86_64", func() {
				iommu := iommupci.NewIommuPCI("x86_64")
				Expect(iommu).NotTo(BeNil())
				// On x86_64, SMMU should not be enabled
				Expect(iommu.SMMUEnabled).To(BeFalse())
			})

			It("should create instance for s390x", func() {
				iommu := iommupci.NewIommuPCI("s390x")
				Expect(iommu).NotTo(BeNil())
				Expect(iommu.SMMUEnabled).To(BeFalse())
			})

			It("should handle unknown architectures", func() {
				iommu := iommupci.NewIommuPCI("unknown")
				Expect(iommu).NotTo(BeNil())
			})
		})
	})

	Describe("BDF Operations", func() {
		Context("when working with BDF devices", func() {
			var bdf *iommupci.BDF

			BeforeEach(func() {
				bdf = iommupci.NewBDFDevice("0000:3b:00.0")
			})

			It("should have correct initial state", func() {
				Expect(bdf.ID).To(Equal("0000:3b:00.0"))
				Expect(bdf.IOMMUFDEnabled).To(BeFalse())
				Expect(bdf.ATSSupported).To(BeFalse())
				Expect(bdf.ATSEnabled).To(BeFalse())
				Expect(bdf.PASIDSupported).To(BeFalse())
				Expect(bdf.SSIDSize).To(Equal(0))
			})
		})
	})

	Describe("IommuPCI Operations", func() {
		Context("when working with IOMMU configuration", func() {
			var iommu *iommupci.IommuPCI

			BeforeEach(func() {
				iommu = iommupci.NewIommuPCI("arm64")
			})

			It("should initialize with zero PCI hole size", func() {
				Expect(iommu.PCIHoleSize).To(Equal(uint64(0)))
			})

			It("should allow accumulating PCI hole sizes", func() {
				// Simulate adding multiple devices
				iommu.PCIHoleSize += 1024 * 1024     // 1 GB
				iommu.PCIHoleSize += 512 * 1024      // 512 MB
				iommu.PCIHoleSize += 2 * 1024 * 1024 // 2 GB

				expectedTotal := uint64(3*1024*1024 + 512*1024) // 3.5 GB in KiB
				Expect(iommu.PCIHoleSize).To(Equal(expectedTotal))
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("when handling boundary conditions", func() {
			It("should handle maximum uint64 values safely", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// This shouldn't panic or overflow
				result := iommupci.CalculateTotalPCIHole64Size(1<<63, 0)
				// Result should be a valid power of 2
				Expect(result).To(BeNumerically(">", 0))
			})

			It("should handle very small sizes", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// 1 byte should round up to 1 KiB
				result := iommupci.CalculateTotalPCIHole64Size(1, 0)
				Expect(result).To(Equal(uint64(1)))
			})

			It("should handle exact power of 2 sizes", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// Powers of 2 should remain unchanged (after KiB conversion)
				testCases := []struct {
					bytes    uint64
					expected uint64
				}{
					{1024, 1},             // 1 KiB
					{2048, 2},             // 2 KiB
					{4096, 4},             // 4 KiB
					{1048576, 1024},       // 1 MiB
					{1073741824, 1048576}, // 1 GiB
				}

				for _, tc := range testCases {
					result := iommupci.CalculateTotalPCIHole64Size(tc.bytes, 0)
					Expect(result).To(Equal(tc.expected))
				}
			})
		})
	})

	Describe("Realistic GPU Scenarios", func() {
		Context("when configuring for real GPU models", func() {
			It("should handle NVIDIA A100 configuration (40 GB)", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// NVIDIA A100 has a 40 GB BAR
				a100BAR := uint64(40) * 1024 * 1024 * 1024 // 40 GB
				margin := uint64(1024 * 1024)              // 1 GB margin

				result := iommupci.CalculateTotalPCIHole64Size(a100BAR, margin)

				// 40 GB + 1 GB = 41 GB = 43008 MiB = 44040192 KiB
				// Next power of 2 is 2^26 = 67108864 KiB (64 GB)
				Expect(result).To(Equal(uint64(67108864)))
			})

			It("should handle NVIDIA H100 configuration (80 GB)", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// NVIDIA H100 has an 80 GB BAR
				h100BAR := uint64(80) * 1024 * 1024 * 1024 // 80 GB
				margin := uint64(1024 * 1024)              // 1 GB margin

				result := iommupci.CalculateTotalPCIHole64Size(h100BAR, margin)

				// 80 GB + 1 GB = 81 GB = 85983232 KiB
				// Next power of 2 is 2^27 = 134217728 KiB (128 GB)
				Expect(result).To(Equal(uint64(134217728)))
			})

			It("should handle multiple small GPUs", func() {
				if !isARM64() {
					Skip("CalculateTotalPCIHole64Size only implemented on ARM64")
				}
				// 4x NVIDIA T4 (16 GB each) = 64 GB total
				iommu := iommupci.NewIommuPCI("arm64")

				for i := 0; i < 4; i++ {
					// Each T4 contributes 16 GB
					iommu.PCIHoleSize += uint64(16 * 1024 * 1024) // in KiB
				}

				result := iommupci.CalculateTotalPCIHole64Size(
					iommu.PCIHoleSize*1024, // convert KiB to bytes
					uint64(1024*1024),      // 1 GB margin
				)

				// 64 GB + 1 GB = 65 GB = 68157440 KiB
				// Next power of 2 is 2^27 = 134217728 KiB (128 GB)
				Expect(result).To(Equal(uint64(134217728)))
			})
		})
	})
})
