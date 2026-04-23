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

package iommu

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt/pkg/pointer"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

func createMockNUMANode(nodeID int, distances string, memTotalKB int, cpulist string) {
	nodeDir := filepath.Join(SysfsNodeBasePath, fmt.Sprintf("node%d", nodeID))
	Expect(os.MkdirAll(nodeDir, 0755)).To(Succeed())

	Expect(os.WriteFile(filepath.Join(nodeDir, "distance"), []byte(distances+"\n"), 0644)).To(Succeed())

	meminfo := fmt.Sprintf("Node %d MemTotal:       %d kB\nNode %d MemFree:        0 kB\n", nodeID, memTotalKB, nodeID)
	Expect(os.WriteFile(filepath.Join(nodeDir, "meminfo"), []byte(meminfo), 0644)).To(Succeed())

	Expect(os.WriteFile(filepath.Join(nodeDir, "cpulist"), []byte(cpulist+"\n"), 0644)).To(Succeed())
}

func createMockPCIDevice(bdf string, numaNode int) {
	devDir := filepath.Join(SysfsPCIBasePath, bdf)
	Expect(os.MkdirAll(devDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(devDir, "numa_node"), []byte(fmt.Sprintf("%d\n", numaNode)), 0644)).To(Succeed())
}

// setupGrace2Socket2GPU creates a mock sysfs layout resembling a 2-socket
// Grace system with 2 GPUs (one per socket):
//
//	Node 0: CPU socket 0 (has memory and CPUs)
//	Node 1: CPU socket 1 (has memory and CPUs)
//	Nodes 4-11: GI nodes for GPU 0 (memory-less, CPU-less)
//	Nodes 12-19: GI nodes for GPU 1 (memory-less, CPU-less)
func setupGrace2Socket2GPU() {
	Expect(os.WriteFile(filepath.Join(SysfsNodeBasePath, "online"), []byte("0-1,4-19\n"), 0644)).To(Succeed())

	//                        0   1   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19
	node0Dist := "10 40 80 80 80 80 80 80 80 80 120 120 120 120 120 120 120 120"
	node1Dist := "40 10 120 120 120 120 120 120 120 120 80 80 80 80 80 80 80 80"

	createMockNUMANode(0, node0Dist, 468713472, "0-59")
	createMockNUMANode(1, node1Dist, 468713472, "60-119")

	giDistTemplate0 := []string{
		"80 120 10 11 11 11 11 11 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 10 11 11 11 11 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 10 11 11 11 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 11 10 11 11 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 11 11 10 11 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 11 11 11 10 11 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 11 11 11 11 10 11 40 40 40 40 40 40 40 40",
		"80 120 11 11 11 11 11 11 11 10 40 40 40 40 40 40 40 40",
	}
	for i, dist := range giDistTemplate0 {
		createMockNUMANode(4+i, dist, 0, "")
	}

	giDistTemplate1 := []string{
		"120 80 40 40 40 40 40 40 40 40 10 11 11 11 11 11 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 10 11 11 11 11 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 10 11 11 11 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 11 10 11 11 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 11 11 10 11 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 11 11 11 10 11 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 11 11 11 11 10 11",
		"120 80 40 40 40 40 40 40 40 40 11 11 11 11 11 11 11 10",
	}
	for i, dist := range giDistTemplate1 {
		createMockNUMANode(12+i, dist, 0, "")
	}

	createMockPCIDevice("0008:01:00.0", 4)
	createMockPCIDevice("0009:01:00.0", 12)
}

func buildGrace2GPUDomain() *api.DomainSpec {
	domain := &api.DomainSpec{
		NUMATune: &api.NUMATune{
			MemNodes: []api.MemNode{
				{CellID: 0, Mode: "strict", NodeSet: "0"},
				{CellID: 1, Mode: "strict", NodeSet: "1"},
			},
		},
	}
	domain.CPU.NUMA = &api.NUMA{
		Cells: []api.NUMACell{
			{ID: "0", CPUs: "0-59", Memory: pointer.P(uint64(468713472)), Unit: "KiB"},
			{ID: "1", CPUs: "60-119", Memory: pointer.P(uint64(468713472)), Unit: "KiB"},
		},
	}
	for i := 2; i <= 17; i++ {
		domain.CPU.NUMA.Cells = append(domain.CPU.NUMA.Cells, api.NUMACell{
			ID:     fmt.Sprintf("%d", i),
			Memory: pointer.P(uint64(0)),
			Unit:   "KiB",
		})
	}
	domain.Devices.HostDevices = []api.HostDevice{
		{
			Source: api.HostDeviceSource{
				Address: &api.Address{Domain: "0x0008", Bus: "0x01", Slot: "0x00", Function: "0x0"},
			},
			ACPI: &api.ACPIHostDev{NodeSet: "2-9"},
		},
		{
			Source: api.HostDeviceSource{
				Address: &api.Address{Domain: "0x0009", Bus: "0x01", Slot: "0x00", Function: "0x0"},
			},
			ACPI: &api.ACPIHostDev{NodeSet: "10-17"},
		},
	}
	return domain
}

func findSibling(siblings []api.NUMACellSibling, targetID string) *api.NUMACellSibling {
	for i := range siblings {
		if siblings[i].ID == targetID {
			return &siblings[i]
		}
	}
	return nil
}

var _ = Describe("NUMA Distances", func() {
	BeforeEach(func() {
		tmpDir := GinkgoT().TempDir()

		origNodeBase := SysfsNodeBasePath
		origPCIBase := SysfsPCIBasePath
		SysfsNodeBasePath = filepath.Join(tmpDir, "sys/devices/system/node")
		SysfsPCIBasePath = filepath.Join(tmpDir, "sys/bus/pci/devices")

		Expect(os.MkdirAll(SysfsNodeBasePath, 0755)).To(Succeed())
		Expect(os.MkdirAll(SysfsPCIBasePath, 0755)).To(Succeed())

		DeferCleanup(func() {
			SysfsNodeBasePath = origNodeBase
			SysfsPCIBasePath = origPCIBase
		})
	})

	Describe("readHostNUMADistances", func() {
		It("should parse sysfs distance files correctly", func() {
			Expect(os.WriteFile(filepath.Join(SysfsNodeBasePath, "online"), []byte("0-1\n"), 0644)).To(Succeed())
			createMockNUMANode(0, "10 40", 468713472, "0-59")
			createMockNUMANode(1, "40 10", 468713472, "60-119")

			distances, err := readHostNUMADistances([]int{0, 1})
			Expect(err).NotTo(HaveOccurred())

			Expect(distances[0][0]).To(Equal(uint64(10)))
			Expect(distances[0][1]).To(Equal(uint64(40)))
			Expect(distances[1][0]).To(Equal(uint64(40)))
			Expect(distances[1][1]).To(Equal(uint64(10)))
		})
	})

	Describe("discoverGPUGINodes", func() {
		It("should find contiguous memory-less GI nodes after the primary node", func() {
			Expect(os.WriteFile(filepath.Join(SysfsNodeBasePath, "online"), []byte("0-1,4-11\n"), 0644)).To(Succeed())
			createMockNUMANode(0, "10 40", 468713472, "0-59")
			createMockNUMANode(1, "40 10", 468713472, "60-119")
			for i := 4; i <= 11; i++ {
				createMockNUMANode(i, "10", 0, "")
			}
			createMockPCIDevice("0008:01:00.0", 4)

			primary, giNodes, err := discoverGPUGINodes("0008:01:00.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(primary).To(Equal(4))
			Expect(giNodes).To(Equal([]int{5, 6, 7, 8, 9, 10, 11}))
		})
	})

	Describe("buildGuestToHostMapping", func() {
		It("should map CPU cells from NUMATune and GI cells from sysfs", func() {
			Expect(os.WriteFile(filepath.Join(SysfsNodeBasePath, "online"), []byte("0-1,4-19\n"), 0644)).To(Succeed())
			createMockNUMANode(0, "10 40", 468713472, "0-59")
			createMockNUMANode(1, "40 10", 468713472, "60-119")
			for i := 4; i <= 11; i++ {
				createMockNUMANode(i, "10", 0, "")
			}
			for i := 12; i <= 19; i++ {
				createMockNUMANode(i, "10", 0, "")
			}
			createMockPCIDevice("0008:01:00.0", 4)
			createMockPCIDevice("0009:01:00.0", 12)

			domain := buildGrace2GPUDomain()
			mapping, err := buildGuestToHostMapping(domain)
			Expect(err).NotTo(HaveOccurred())

			Expect(mapping[0]).To(Equal(0))
			Expect(mapping[1]).To(Equal(1))
			for i := 0; i < 8; i++ {
				Expect(mapping[2+i]).To(Equal(4+i), "GPU 0 GI cell %d", 2+i)
			}
			for i := 0; i < 8; i++ {
				Expect(mapping[10+i]).To(Equal(12+i), "GPU 1 GI cell %d", 10+i)
			}
		})
	})

	Describe("applyNUMADistances", func() {
		Context("with a Grace 2-socket 2-GPU topology", func() {
			BeforeEach(func() {
				setupGrace2Socket2GPU()
			})

			It("should set correct distances on CPU cell 0", func() {
				domain := buildGrace2GPUDomain()
				applyNUMADistances(domain)

				cell0 := domain.CPU.NUMA.Cells[0]
				Expect(cell0.Distances).NotTo(BeNil())
				Expect(findSibling(cell0.Distances.Siblings, "0").Value).To(Equal(uint64(10)))   // self
				Expect(findSibling(cell0.Distances.Siblings, "1").Value).To(Equal(uint64(40)))   // cross-socket
				Expect(findSibling(cell0.Distances.Siblings, "2").Value).To(Equal(uint64(80)))   // GPU 0 GI (local)
				Expect(findSibling(cell0.Distances.Siblings, "10").Value).To(Equal(uint64(120))) // GPU 1 GI (remote)
			})

			It("should set correct distances on CPU cell 1", func() {
				domain := buildGrace2GPUDomain()
				applyNUMADistances(domain)

				cell1 := domain.CPU.NUMA.Cells[1]
				Expect(cell1.Distances).NotTo(BeNil())
				Expect(findSibling(cell1.Distances.Siblings, "0").Value).To(Equal(uint64(40)))  // cross-socket
				Expect(findSibling(cell1.Distances.Siblings, "1").Value).To(Equal(uint64(10)))  // self
				Expect(findSibling(cell1.Distances.Siblings, "2").Value).To(Equal(uint64(120))) // GPU 0 GI (remote)
				Expect(findSibling(cell1.Distances.Siblings, "10").Value).To(Equal(uint64(80))) // GPU 1 GI (local)
			})

			It("should set correct distances on GI cell 2 (GPU 0 primary)", func() {
				domain := buildGrace2GPUDomain()
				applyNUMADistances(domain)

				cell2 := domain.CPU.NUMA.Cells[2]
				Expect(cell2.Distances).NotTo(BeNil())
				Expect(findSibling(cell2.Distances.Siblings, "0").Value).To(Equal(uint64(80)))  // local CPU
				Expect(findSibling(cell2.Distances.Siblings, "1").Value).To(Equal(uint64(120))) // remote CPU
				Expect(findSibling(cell2.Distances.Siblings, "2").Value).To(Equal(uint64(10)))  // self
				Expect(findSibling(cell2.Distances.Siblings, "3").Value).To(Equal(uint64(11)))  // same GPU group
				Expect(findSibling(cell2.Distances.Siblings, "10").Value).To(Equal(uint64(40))) // GPU 1 GI
			})

			It("should set correct distances on GI cell 10 (GPU 1 primary)", func() {
				domain := buildGrace2GPUDomain()
				applyNUMADistances(domain)

				cell10 := domain.CPU.NUMA.Cells[10]
				Expect(cell10.Distances).NotTo(BeNil())
				Expect(findSibling(cell10.Distances.Siblings, "0").Value).To(Equal(uint64(120))) // remote CPU
				Expect(findSibling(cell10.Distances.Siblings, "1").Value).To(Equal(uint64(80)))  // local CPU
				Expect(findSibling(cell10.Distances.Siblings, "2").Value).To(Equal(uint64(40)))  // GPU 0 GI
				Expect(findSibling(cell10.Distances.Siblings, "10").Value).To(Equal(uint64(10))) // self
				Expect(findSibling(cell10.Distances.Siblings, "11").Value).To(Equal(uint64(11))) // same GPU group
			})

			It("should set distances on all cells", func() {
				domain := buildGrace2GPUDomain()
				applyNUMADistances(domain)

				for i, cell := range domain.CPU.NUMA.Cells {
					Expect(cell.Distances).NotTo(BeNil(), "cell %d (%s) has no distances", i, cell.ID)
				}
			})
		})

		Context("graceful degradation", func() {
			It("should be a no-op when there are no NUMA cells", func() {
				domain := &api.DomainSpec{}
				applyNUMADistances(domain)

				domain.CPU.NUMA = &api.NUMA{Cells: []api.NUMACell{}}
				applyNUMADistances(domain)
			})

			It("should still apply CPU-to-CPU distances when there are no GI devices", func() {
				Expect(os.WriteFile(filepath.Join(SysfsNodeBasePath, "online"), []byte("0-1\n"), 0644)).To(Succeed())
				createMockNUMANode(0, "10 40", 468713472, "0-59")
				createMockNUMANode(1, "40 10", 468713472, "60-119")

				domain := &api.DomainSpec{
					NUMATune: &api.NUMATune{
						MemNodes: []api.MemNode{
							{CellID: 0, Mode: "strict", NodeSet: "0"},
							{CellID: 1, Mode: "strict", NodeSet: "1"},
						},
					},
				}
				domain.CPU.NUMA = &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-59", Memory: pointer.P(uint64(468713472)), Unit: "KiB"},
						{ID: "1", CPUs: "60-119", Memory: pointer.P(uint64(468713472)), Unit: "KiB"},
					},
				}

				applyNUMADistances(domain)

				Expect(domain.CPU.NUMA.Cells[0].Distances).NotTo(BeNil())
				Expect(findSibling(domain.CPU.NUMA.Cells[0].Distances.Siblings, "0").Value).To(Equal(uint64(10)))
				Expect(findSibling(domain.CPU.NUMA.Cells[0].Distances.Siblings, "1").Value).To(Equal(uint64(40)))
			})
		})
	})
})
