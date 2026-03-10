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

package hardware

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("Hardware utils test", func() {

	Context("cpuset parser", func() {
		It("shoud parse cpuset correctly", func() {
			expectedList := []int{0, 1, 2, 7, 12, 13, 14}
			cpusetLine := "0-2,7,12-14"
			lst, err := ParseCPUSetLine(cpusetLine, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(lst).To(HaveLen(7))
			Expect(lst).To(Equal(expectedList))
		})

		It("should reject expanding arbitrary ranges which would overload a machine", func() {
			cpusetLine := "0-100000000000"
			_, err := ParseCPUSetLine(cpusetLine, 100)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("safety"))
		})
	})

	Context("count vCPUs", func() {
		It("shoud count vCPUs correctly", func() {
			vCPUs := GetNumberOfVCPUs(&v1.CPU{
				Sockets: 2,
				Cores:   2,
				Threads: 2,
			})
			Expect(vCPUs).To(Equal(int64(8)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Sockets: 2,
			})
			Expect(vCPUs).To(Equal(int64(2)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Cores: 2,
			})
			Expect(vCPUs).To(Equal(int64(2)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Threads: 2,
			})
			Expect(vCPUs).To(Equal(int64(2)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Sockets: 2,
				Threads: 2,
			})
			Expect(vCPUs).To(Equal(int64(4)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Sockets: 2,
				Cores:   2,
			})
			Expect(vCPUs).To(Equal(int64(4)), "Expect vCPUs")

			vCPUs = GetNumberOfVCPUs(&v1.CPU{
				Cores:   2,
				Threads: 2,
			})
			Expect(vCPUs).To(Equal(int64(4)), "Expect vCPUs")
		})
	})

	Context("parse PCI address", func() {
		It("shoud return an array of PCI DBSF fields (domain, bus, slot, function) or an error for malformed address", func() {
			testData := []struct {
				addr        string
				expectation []string
			}{
				{"05EA:Fc:1d.6", []string{"05EA", "Fc", "1d", "6"}},
				{"", nil},
				{"invalid address", nil},
				{" 05EA:Fc:1d.6", nil}, // leading symbol
				{"05EA:Fc:1d.6 ", nil}, // trailing symbol
				{"00Z0:00:1d.6", nil},  // invalid digit in domain
				{"0000:z0:1d.6", nil},  // invalid digit in bus
				{"0000:00:Zd.6", nil},  // invalid digit in slot
				{"05EA:Fc:1d:6", nil},  // colon ':' instead of dot '.' after slot
				{"0000:00:1d.9", nil},  // invalid function
			}

			for _, t := range testData {
				res, err := ParsePciAddress(t.addr)
				Expect(res).To(Equal(t.expectation))
				if t.expectation == nil {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})
	})

	Context("mdev parent detection", func() {
		var (
			tmpDir  string
			oldBase string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "kubevirt-mdev-*")
			Expect(err).ToNot(HaveOccurred())
			oldBase = mdevDevicesBasePath
			mdevDevicesBasePath = tmpDir
		})

		AfterEach(func() {
			mdevDevicesBasePath = oldBase
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("extracts the parent pci address from the mdev_type symlink real path", func() {
			uuid := "435a944e-07a7-4aab-8502-7c2b327ae40a"
			mdevDir := filepath.Join(tmpDir, uuid)
			Expect(os.MkdirAll(mdevDir, 0o755)).To(Succeed())

			parentBDF := "0000:82:00.4"
			target := filepath.Join(tmpDir, "devices", "pci0000:80", "0000:80:03.1", parentBDF, "mdev_supported_types", "nvidia-1168")
			Expect(os.MkdirAll(target, 0o755)).To(Succeed())
			Expect(os.Symlink(target, filepath.Join(mdevDir, "mdev_type"))).To(Succeed())

			addr, err := GetMdevParentPCIAddress(uuid)
			Expect(err).ToNot(HaveOccurred())
			Expect(addr).To(Equal("0000:82:00.4"))
		})

		It("returns canonical lower-case bdf", func() {
			uuid := "canonical-bdf"
			mdevDir := filepath.Join(tmpDir, uuid)
			Expect(os.MkdirAll(mdevDir, 0o755)).To(Succeed())

			parentBDF := "0000:AF:00.0"
			target := filepath.Join(tmpDir, "sys", "devices", "pci0000:80", "0000:80:03.1", parentBDF, "mdev_supported_types", "type-1")
			Expect(os.MkdirAll(target, 0o755)).To(Succeed())
			Expect(os.Symlink(target, filepath.Join(mdevDir, "mdev_type"))).To(Succeed())

			addr, err := GetMdevParentPCIAddress(uuid)
			Expect(err).ToNot(HaveOccurred())
			Expect(addr).To(Equal("0000:af:00.0"))
		})

		It("fails when the mdev_type symlink is missing", func() {
			_, err := GetMdevParentPCIAddress("missing")
			Expect(err).To(HaveOccurred())
		})

		It("fails when the parent path is not a pci address", func() {
			uuid := "invalid-parent"
			mdevDir := filepath.Join(tmpDir, uuid)
			Expect(os.MkdirAll(mdevDir, 0o755)).To(Succeed())

			target := filepath.Join(tmpDir, "weird", "path", "not-a-bdf", "mdev_supported_types", "type-1")
			Expect(os.MkdirAll(target, 0o755)).To(Succeed())
			Expect(os.Symlink(target, filepath.Join(mdevDir, "mdev_type"))).To(Succeed())

			_, err := GetMdevParentPCIAddress(uuid)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("PCI topology grouping", func() {
		var (
			tmpDir  string
			oldBase string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "kubevirt-pci-topology-*")
			Expect(err).ToNot(HaveOccurred())

			oldBase = pciDevicesBasePath

			// Build a fake hierarchy resembling:
			// 0000:00:01.0 -> 0000:01:00.0 -> 0000:02:01.0 -> 0000:03:00.0
			//                                     |
			//                                     -> 0000:02:02.0 -> 0000:04:00.0
			// 0000:00:01.0 -> 0000:01:01.0 -> 0000:05:00.0
			root := filepath.Join(tmpDir, "sys", "devices", "pci0000:00")
			Expect(os.MkdirAll(filepath.Join(root, "0000:00:01.0", "0000:01:00.0", "0000:02:01.0", "0000:03:00.0"), 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(root, "0000:00:01.0", "0000:01:00.0", "0000:02:02.0", "0000:04:00.0"), 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(root, "0000:00:01.0", "0000:01:01.0", "0000:05:00.0"), 0o755)).To(Succeed())

			devicesBase := filepath.Join(tmpDir, "sys", "bus", "pci", "devices")
			Expect(os.MkdirAll(devicesBase, 0o755)).To(Succeed())

			Expect(os.Symlink(filepath.Join(root, "0000:00:01.0", "0000:01:00.0", "0000:02:01.0", "0000:03:00.0"),
				filepath.Join(devicesBase, "0000:03:00.0"))).To(Succeed())
			Expect(os.Symlink(filepath.Join(root, "0000:00:01.0", "0000:01:00.0", "0000:02:02.0", "0000:04:00.0"),
				filepath.Join(devicesBase, "0000:04:00.0"))).To(Succeed())
			Expect(os.Symlink(filepath.Join(root, "0000:00:01.0", "0000:01:01.0", "0000:05:00.0"),
				filepath.Join(devicesBase, "0000:05:00.0"))).To(Succeed())

			pciDevicesBasePath = devicesBase
		})

		AfterEach(func() {
			pciDevicesBasePath = oldBase
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("returns the canonical hierarchy for a device", func() {
			hierarchy, err := GetDevicePCIPathHierarchy("0000:03:00.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(hierarchy).To(Equal([]string{
				"0000:00:01.0",
				"0000:01:00.0",
				"0000:02:01.0",
				"0000:03:00.0",
			}))
		})

		It("groups devices by their upstream root port and first switch", func() {
			groupA, err := GetDevicePCITopologyGroup("0000:03:00.0")
			Expect(err).ToNot(HaveOccurred())
			groupB, err := GetDevicePCITopologyGroup("0000:04:00.0")
			Expect(err).ToNot(HaveOccurred())
			groupC, err := GetDevicePCITopologyGroup("0000:05:00.0")
			Expect(err).ToNot(HaveOccurred())

			Expect(groupA).To(Equal("0000:00:01.0/0000:01:00.0"))
			Expect(groupB).To(Equal("0000:00:01.0/0000:01:00.0"))
			Expect(groupC).To(Equal("0000:00:01.0/0000:01:01.0"))
		})

	})

	Context("GetDeviceIOMMUGroupInfo", func() {
		var (
			tmpDir string
			oldDir string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "kubevirt-iommu-*")
			Expect(err).ToNot(HaveOccurred())

			root := filepath.Join(tmpDir, "sys", "devices", "pci0000:00", "0000:00:01.0", "0000:01:00.0", "0000:02:01.0", "0000:03:00.0")
			Expect(os.MkdirAll(root, 0o755)).To(Succeed())
			other := filepath.Join(tmpDir, "sys", "devices", "pci0000:00", "0000:00:01.0", "0000:01:00.0", "0000:02:02.0", "0000:04:00.0")
			Expect(os.MkdirAll(other, 0o755)).To(Succeed())

			devicesBase := filepath.Join(tmpDir, "sys", "bus", "pci", "devices")
			Expect(os.MkdirAll(devicesBase, 0o755)).To(Succeed())

			Expect(os.Symlink(root, filepath.Join(devicesBase, "0000:03:00.0"))).To(Succeed())
			Expect(os.Symlink(other, filepath.Join(devicesBase, "0000:04:00.0"))).To(Succeed())

			groupDir := filepath.Join(tmpDir, "sys", "kernel", "iommu_groups", "7")
			Expect(os.MkdirAll(filepath.Join(groupDir, "devices"), 0o755)).To(Succeed())

			groupLinkTarget := groupDir
			Expect(os.Symlink(groupLinkTarget, filepath.Join(root, "iommu_group"))).To(Succeed())

			Expect(os.Symlink(filepath.Join(devicesBase, "0000:03:00.0"), filepath.Join(groupDir, "devices", "0000:03:00.0"))).To(Succeed())

			oldDir = pciDevicesBasePath
			pciDevicesBasePath = devicesBase
		})

		AfterEach(func() {
			pciDevicesBasePath = oldDir
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("returns group information when present", func() {
			groupID, peers, err := GetDeviceIOMMUGroupInfo("0000:03:00.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(groupID).To(Equal(7))
			Expect(peers).To(ContainElement("0000:03:00.0"))
		})

		It("handles devices without groups gracefully", func() {
			groupID, peers, err := GetDeviceIOMMUGroupInfo("0000:04:00.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(groupID).To(Equal(-1))
			Expect(peers).To(BeEmpty())
		})
	})
})
