package converter

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt/pkg/util/hardware"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	iommupci "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/iommu-pci"
)

type devicePlacementTestCase struct {
	name                  string
	numaCells             []api.NUMACell
	vcpuPins              []api.CPUTuneVCPUPin
	devices               []api.HostDevice
	expectedControllers   int
	expectedExpanderBuses int
	expectedRootPorts     int
	expectedError         string
}

type addDevicesTestCase struct {
	name            string
	devices         []api.HostDevice
	numaCells       []api.NUMACell
	vcpuPins        []api.CPUTuneVCPUPin
	expectedDevices int
	description     string
}

var _ = Describe("PCIe Expander Bus Assigner", func() {
	var (
		originalPciBasePath  string
		originalNodeBasePath string
		fakePciBasePath      string
		fakeNodeBasePath     string
	)

	createDomainSpecWithNUMA := func(numaCells []api.NUMACell, vcpuPins []api.CPUTuneVCPUPin) *api.DomainSpec {
		spec := &api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{},
			},
		}
		if len(numaCells) > 0 {
			spec.CPU = api.CPU{
				NUMA: &api.NUMA{Cells: numaCells},
			}
		}
		if len(vcpuPins) > 0 {
			spec.CPUTune = &api.CPUTune{VCPUPin: vcpuPins}
		}
		return spec
	}

	createPCIDevice := func(alias, bus string) api.HostDevice {
		return api.HostDevice{
			Type:  api.HostDevicePCI,
			Alias: api.NewUserDefinedAlias(alias),
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Domain: "0x0000", Bus: bus,
					Slot: "0x00", Function: "0x0",
				},
			},
		}
	}

	createIOMMUPCIDevice := func(alias, bus string) api.HostDevice {
		dev := createPCIDevice(alias, bus)
		dev.ACPI = &api.ACPIHostDev{NodeSet: "tofill"}
		return dev
	}

	createNonPCIDevice := func(deviceType string) api.HostDevice {
		return api.HostDevice{
			Type: deviceType,
		}
	}

	createPCIDeviceWithoutAddress := func(alias string) api.HostDevice {
		return api.HostDevice{
			Type:  api.HostDevicePCI,
			Alias: api.NewUserDefinedAlias(alias),
		}
	}

	setupFakeSysfs := func() {
		var err error
		fakePciBasePath, err = os.MkdirTemp("", "pci_devices")
		Expect(err).ToNot(HaveOccurred())

		fakeNodeBasePath, err = os.MkdirTemp("", "numa_nodes")
		Expect(err).ToNot(HaveOccurred())

		// Create test PCI devices with NUMA nodes
		//
		// Devices on NUMA nodes 0 and 1 (CPU nodes):
		//   0000:01:00.0 -> NUMA 0 (has CPUs 0-3)
		//   0000:02:00.0 -> NUMA 1 (has CPUs 4-7)
		//   0000:03:00.0 -> NUMA 0
		//   0000:04:00.0 -> NUMA 1
		//   0000:05:00.0 -> NUMA 0
		//
		// Devices on GPU NUMA nodes (no CPUs, modelling GB200 topology):
		//   0000:06:00.0 -> NUMA 2 (GPU HBM node, no CPUs)
		//   0000:07:00.0 -> NUMA 10 (GPU HBM node, no CPUs)
		testDevices := map[string]string{
			"0000:01:00.0": "0",
			"0000:02:00.0": "1",
			"0000:03:00.0": "0",
			"0000:04:00.0": "1",
			"0000:05:00.0": "0",
			"0000:06:00.0": "2",
			"0000:07:00.0": "10",
		}

		for pciAddr, numaNode := range testDevices {
			pciDevicePath := filepath.Join(fakePciBasePath, pciAddr)
			err = os.MkdirAll(pciDevicePath, 0o755)
			Expect(err).ToNot(HaveOccurred())

			numaNodeFile := filepath.Join(pciDevicePath, "numa_node")
			err = os.WriteFile(numaNodeFile, []byte(numaNode+"\n"), 0o644)
			Expect(err).ToNot(HaveOccurred())
		}

		// Create NUMA node directories
		// Nodes 0 and 1 have CPUs (Grace CPU sockets)
		// Nodes 2 and 10 have no CPUs (GPU HBM nodes on a GB200)
		for numaID, cpuList := range map[string]string{
			"0":  "0-3",
			"1":  "4-7",
			"2":  "",
			"10": "",
		} {
			numaNodePath := filepath.Join(fakeNodeBasePath, "node"+numaID)
			err = os.MkdirAll(numaNodePath, 0o755)
			Expect(err).ToNot(HaveOccurred())

			cpuListFile := filepath.Join(numaNodePath, "cpulist")
			err = os.WriteFile(cpuListFile, []byte(cpuList+"\n"), 0o644)
			Expect(err).ToNot(HaveOccurred())
		}
	}

	BeforeEach(func() {
		originalPciBasePath = hardware.PciBasePath
		originalNodeBasePath = hardware.NodeBasePath
		setupFakeSysfs()
		hardware.PciBasePath = fakePciBasePath
		hardware.NodeBasePath = fakeNodeBasePath
	})

	AfterEach(func() {
		hardware.PciBasePath = originalPciBasePath
		hardware.NodeBasePath = originalNodeBasePath
		if fakePciBasePath != "" {
			os.RemoveAll(fakePciBasePath)
		}
		if fakeNodeBasePath != "" {
			os.RemoveAll(fakeNodeBasePath)
		}
	})

	Describe("getCurrentControllerIndex", func() {
		It("should return the highest index of the existing controllers", func() {
			domainSpec := &api.DomainSpec{
				Devices: api.Devices{
					Controllers: []api.Controller{
						{Model: api.ControllerModelPCIeRoot, Index: "0"},
						{Model: api.ControllerModelPCIeRootPort, Index: "4"},
					},
				},
			}

			Expect(getCurrentControllerIndex(domainSpec)).To(Equal(uint32(4)))
		})
	})

	Describe("expanderBusAssigner", func() {
		var (
			assigner   *expanderBusAssigner
			domainSpec *api.DomainSpec
			iommuPCI   *iommupci.IommuPCI
		)

		BeforeEach(func() {
			domainSpec = createDomainSpecWithNUMA(
				[]api.NUMACell{{ID: "0", CPUs: "0-1"}, {ID: "1", CPUs: "2-3"}},
				[]api.CPUTuneVCPUPin{{VCPU: 0, CPUSet: "0"}, {VCPU: 2, CPUSet: "4"}},
			)
			iommuPCI = iommupci.NewIommuPCI(runtime.GOARCH)
			assigner = newExpanderBusAssigner(domainSpec, iommuPCI)
		})

		DescribeTable("addDevices",
			func(testCase addDevicesTestCase) {
				if testCase.numaCells != nil || testCase.vcpuPins != nil {
					domainSpec = createDomainSpecWithNUMA(testCase.numaCells, testCase.vcpuPins)
					domainSpec.Devices.HostDevices = testCase.devices
					assigner = newExpanderBusAssigner(domainSpec, iommuPCI)
				}

				assigner.addDevices(testCase.devices)
				Expect(assigner.devices).To(HaveLen(testCase.expectedDevices), testCase.description)
			},
			Entry("filters non-PCI devices", addDevicesTestCase{
				name: "mixed device types",
				devices: []api.HostDevice{
					createPCIDevice("pci1", "0x01"),
					createNonPCIDevice("usb"),
					createPCIDevice("pci2", "0x02"),
					createNonPCIDevice("scsi"),
				},
				expectedDevices: 2,
				description:     "should only accept PCI devices",
			}),
			Entry("filters devices without source address", addDevicesTestCase{
				name: "devices without address",
				devices: []api.HostDevice{
					createPCIDevice("pci1", "0x01"),
					createPCIDeviceWithoutAddress("pci2"),
				},
				expectedDevices: 1,
				description:     "should skip devices without source address",
			}),
			Entry("filters devices without NUMA affinity", addDevicesTestCase{
				name:            "devices without NUMA topology",
				devices:         []api.HostDevice{createPCIDevice("pci1", "0x01")},
				numaCells:       []api.NUMACell{},
				vcpuPins:        []api.CPUTuneVCPUPin{},
				expectedDevices: 0,
				description:     "should skip devices when no NUMA topology is configured",
			}),
			Entry("accepts PCI devices with NUMA affinity", addDevicesTestCase{
				name: "valid PCI devices",
				devices: []api.HostDevice{
					createPCIDevice("device1", "0x01"),
					createPCIDevice("device2", "0x02"),
				},
				expectedDevices: 2,
				description:     "should accept all valid PCI devices with NUMA affinity",
			}),
			Entry("accepts cross-NUMA device by creating CPU-less NUMA cell", addDevicesTestCase{
				name: "device on NUMA node without vCPUs (cross-NUMA)",
				devices: []api.HostDevice{
					createPCIDevice("gpu1", "0x02"), // NUMA 1, no vCPUs pinned there
				},
				numaCells: []api.NUMACell{{ID: "0", CPUs: "0-1"}},
				vcpuPins: []api.CPUTuneVCPUPin{
					{VCPU: 0, CPUSet: "0"},
					{VCPU: 1, CPUSet: "1"},
				},
				expectedDevices: 1,
				description:     "should accept cross-NUMA device by creating a CPU-less guest NUMA cell",
			}),
			Entry("accepts both co-located and cross-NUMA devices", addDevicesTestCase{
				name: "mixed NUMA alignment",
				devices: []api.HostDevice{
					createPCIDevice("gpu_numa0", "0x01"), // NUMA 0, vCPUs present
					createPCIDevice("gpu_numa1", "0x02"), // NUMA 1, no vCPUs
				},
				numaCells: []api.NUMACell{{ID: "0", CPUs: "0-1"}},
				vcpuPins: []api.CPUTuneVCPUPin{
					{VCPU: 0, CPUSet: "0"},
					{VCPU: 1, CPUSet: "1"},
				},
				expectedDevices: 2,
				description:     "should accept both co-located and cross-NUMA devices",
			}),
		)

		DescribeTable("PlaceNumaAlignedDevices",
			func(testCase devicePlacementTestCase) {
				if testCase.numaCells != nil || testCase.vcpuPins != nil {
					domainSpec = createDomainSpecWithNUMA(testCase.numaCells, testCase.vcpuPins)
					assigner = newExpanderBusAssigner(domainSpec, iommuPCI)
				}

				domainSpec.Devices.HostDevices = testCase.devices

				err := assigner.PlaceNumaAlignedDevices()

				if testCase.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(testCase.expectedError))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}

				if testCase.expectedControllers >= 0 {
					Expect(domainSpec.Devices.Controllers).To(HaveLen(testCase.expectedControllers))
				}

				if testCase.expectedExpanderBuses > 0 {
					expanderBusCount := 0
					for _, controller := range domainSpec.Devices.Controllers {
						if controller.Model == api.ControllerModelPCIeExpanderBus {
							expanderBusCount++
						}
					}
					Expect(expanderBusCount).To(Equal(testCase.expectedExpanderBuses))
				}

				if testCase.expectedRootPorts > 0 {
					rootPortCount := 0
					for _, controller := range domainSpec.Devices.Controllers {
						if controller.Model == api.ControllerModelPCIeRootPort {
							rootPortCount++
						}
					}
					Expect(rootPortCount).To(Equal(testCase.expectedRootPorts))
				}
			},
			Entry("handles empty device list", devicePlacementTestCase{
				name:                "no devices",
				devices:             []api.HostDevice{},
				expectedControllers: 0,
			}),
			Entry("places single device on single NUMA node", devicePlacementTestCase{
				name:                  "single device",
				devices:               []api.HostDevice{createPCIDevice("device1", "0x01")},
				expectedControllers:   2,
				expectedExpanderBuses: 1,
				expectedRootPorts:     1,
			}),
			Entry("places multiple devices on same NUMA node", devicePlacementTestCase{
				name: "multiple devices same NUMA",
				devices: []api.HostDevice{
					createPCIDevice("device1", "0x01"),
					createPCIDevice("device2", "0x03"),
				},
				expectedControllers:   3,
				expectedExpanderBuses: 1,
				expectedRootPorts:     2,
			}),
			Entry("places devices on different NUMA nodes", devicePlacementTestCase{
				name: "devices on different NUMA nodes",
				devices: []api.HostDevice{
					createPCIDevice("device_numa0", "0x01"),
					createPCIDevice("device_numa1", "0x02"),
				},
				expectedControllers:   4,
				expectedExpanderBuses: 2,
				expectedRootPorts:     2,
			}),
			Entry("handles domain spec without NUMA topology", devicePlacementTestCase{
				name:                "no NUMA topology",
				numaCells:           []api.NUMACell{},
				vcpuPins:            []api.CPUTuneVCPUPin{},
				devices:             []api.HostDevice{createPCIDevice("device1", "0x01")},
				expectedControllers: 0,
			}),
			Entry("handles domain spec without CPU affinity", devicePlacementTestCase{
				name:                "no CPU affinity",
				numaCells:           []api.NUMACell{{ID: "0", CPUs: "0-1"}},
				vcpuPins:            []api.CPUTuneVCPUPin{},
				devices:             []api.HostDevice{createPCIDevice("device1", "0x01")},
				expectedControllers: 0,
			}),
			Entry("places cross-NUMA device with CPU-less NUMA cell", devicePlacementTestCase{
				name: "cross-NUMA: device on NUMA 1, vCPUs only on NUMA 0",
				numaCells: []api.NUMACell{{ID: "0", CPUs: "0-1"}},
				vcpuPins: []api.CPUTuneVCPUPin{
					{VCPU: 0, CPUSet: "0"},
					{VCPU: 1, CPUSet: "1"},
				},
				devices:               []api.HostDevice{createPCIDevice("gpu1", "0x02")},
				expectedControllers:   2,
				expectedExpanderBuses: 1,
				expectedRootPorts:     1,
			}),
			Entry("places both co-located and cross-NUMA devices", devicePlacementTestCase{
				name: "mixed: one device co-located, one cross-NUMA",
				numaCells: []api.NUMACell{{ID: "0", CPUs: "0-1"}},
				vcpuPins: []api.CPUTuneVCPUPin{
					{VCPU: 0, CPUSet: "0"},
					{VCPU: 1, CPUSet: "1"},
				},
				devices: []api.HostDevice{
					createPCIDevice("gpu_numa0", "0x01"), // NUMA 0, vCPUs present
					createPCIDevice("gpu_numa1", "0x02"), // NUMA 1, no vCPUs
				},
				expectedControllers:   4,
				expectedExpanderBuses: 2,
				expectedRootPorts:     2,
			}),
		)
	})

	Describe("PlacePCIDevicesWithNUMAAlignment", func() {
		var (
			domainSpec *api.DomainSpec
			iommuPCI   *iommupci.IommuPCI
		)

		BeforeEach(func() {
			domainSpec = createDomainSpecWithNUMA(
				[]api.NUMACell{{ID: "0", CPUs: "0-1"}, {ID: "1", CPUs: "2-3"}},
				[]api.CPUTuneVCPUPin{{VCPU: 0, CPUSet: "0"}, {VCPU: 2, CPUSet: "4"}},
			)
			iommuPCI = iommupci.NewIommuPCI(runtime.GOARCH)
		})

		It("should return error when controller index exceeds the last expander bus number", func() {
			// Set current controller index to the maximum to trigger the validation
			domainSpec.Devices.Controllers = []api.Controller{
				{Model: api.ControllerModelPCIeRootPort, Index: strconv.Itoa(maxExpanderBusNr)},
			}

			// Add a device, this would require creating new controllers
			domainSpec.Devices.HostDevices = []api.HostDevice{createPCIDevice("device1", "0x01")}

			err := PlacePCIDevicesWithNUMAAlignment(domainSpec, iommuPCI)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("insufficient bus numbers for NUMA-aligned PCIe topology"))
			Expect(err.Error()).To(ContainSubstring("current controller index 256"))
			Expect(err.Error()).To(ContainSubstring("last assigned expander bus number 255"))
		})

		It("should assign bus numbers for expander buses calculated as maxBusNr - controllerCount + 1", func() {
			domainSpec.Devices.HostDevices = []api.HostDevice{
				createPCIDevice("device1", "0x01"),
				createPCIDevice("device2", "0x02"),
			}

			err := PlacePCIDevicesWithNUMAAlignment(domainSpec, iommuPCI)
			Expect(err).ToNot(HaveOccurred())

			// Bus numbers calculated as 254 - controllerCount + 1:
			// NUMA 0: 255 - 2 + 1 = 254 (after creating expander bus + root port)
			// NUMA 1: 255 - 4 + 1 = 252 (after creating 2nd expander bus + root port)
			expectedBusNumbers := map[uint32]bool{254: false, 252: false}
			for _, controller := range domainSpec.Devices.Controllers {
				if controller.Model == api.ControllerModelPCIeExpanderBus {
					Expect(controller.Target).ToNot(BeNil())
					Expect(controller.Target.BusNr).ToNot(BeNil())
					busNr := *controller.Target.BusNr
					_, expected := expectedBusNumbers[busNr]
					Expect(expected).To(BeTrue(), "Bus number %d should be one of the expected values (253, 251)", busNr)
					expectedBusNumbers[busNr] = true
				}
			}

			// Ensure both expected bus numbers were assigned
			for busNr, assigned := range expectedBusNumbers {
				Expect(assigned).To(BeTrue(), "Expected bus number %d was not assigned", busNr)
			}
		})

		It("should assign devices to correct root ports", func() {
			domainSpec.Devices.HostDevices = []api.HostDevice{
				createPCIDevice("device1", "0x01"),
				createPCIDevice("device2", "0x02"),
			}

			err := PlacePCIDevicesWithNUMAAlignment(domainSpec, iommuPCI)
			Expect(err).ToNot(HaveOccurred())

			for _, device := range domainSpec.Devices.HostDevices {
				Expect(device.Address).ToNot(BeNil())
				Expect(device.Address.Type).To(Equal(api.AddressPCI))
				Expect(device.Address.Bus).ToNot(BeEmpty())
				Expect(device.Address.Slot).To(Equal("0x00"))
			}
		})

		It("should place cross-NUMA device under expander bus with CPU-less NUMA cell", func() {
			// Simulate a topology like NVIDIA GB200 where GPUs are on
			// separate NUMA nodes (2, 10) from the CPUs (NUMA 0, 1).
			// See: https://docs.nvidia.com/dccpu/grace-perf-tuning-guide/system.html
			//
			// On a 2-superchip GB200:
			//   NUMA 0: Grace CPU #1 (CPUs 0-71, ~490 GB LPDDR5X)
			//   NUMA 1: Grace CPU #2 (CPUs 72-143, ~490 GB LPDDR5X)
			//   NUMA 2: Blackwell GPU #1 (no CPUs, ~188 GB HBM)
			//   NUMA 10: Blackwell GPU #2 (no CPUs, ~188 GB HBM)
			domainSpec = createDomainSpecWithNUMA(
				[]api.NUMACell{
					{ID: "0", CPUs: "0-1"},
					{ID: "1", CPUs: "2-3"},
				},
				[]api.CPUTuneVCPUPin{
					{VCPU: 0, CPUSet: "0"},
					{VCPU: 1, CPUSet: "1"},
					{VCPU: 2, CPUSet: "4"},
					{VCPU: 3, CPUSet: "5"},
				},
			)

			gpu1 := createIOMMUPCIDevice("gpu_numa2", "0x06")  // NUMA 2, no CPUs (GPU HBM node)
			gpu2 := createIOMMUPCIDevice("gpu_numa10", "0x07") // NUMA 10, no CPUs (GPU HBM node)

			domainSpec.Devices.HostDevices = []api.HostDevice{gpu1, gpu2}

			err := PlacePCIDevicesWithNUMAAlignment(domainSpec, iommuPCI)
			Expect(err).ToNot(HaveOccurred())

			// Both GPUs should be placed under expander buses
			for i, device := range domainSpec.Devices.HostDevices {
				Expect(device.Address).ToNot(BeNil(), "device %d should have an address assigned", i)
				Expect(device.Address.Type).To(Equal(api.AddressPCI))
				Expect(device.Address.Slot).To(Equal("0x00"))
			}

			// Verify two expander buses were created (one per GPU NUMA node)
			expanderBuses := make(map[uint32]*api.Controller)
			for i := range domainSpec.Devices.Controllers {
				ctrl := &domainSpec.Devices.Controllers[i]
				if ctrl.Model == api.ControllerModelPCIeExpanderBus {
					Expect(ctrl.Target).ToNot(BeNil())
					Expect(ctrl.Target.NUMANode).ToNot(BeNil())
					expanderBuses[*ctrl.Target.NUMANode] = ctrl
				}
			}
			Expect(expanderBuses).To(HaveLen(2), "expected 2 expander buses (one per GPU)")

			// Each GPU gets its own SMMUv3 since each is on its own NUMA node
			Expect(domainSpec.Devices.IOMMU).To(HaveLen(2), "each GPU should have its own smmuv3")

			// Verify CPU-less NUMA cells were created for the GPU NUMA nodes
			Expect(domainSpec.CPU.NUMA).ToNot(BeNil())
			Expect(domainSpec.CPU.NUMA.Cells).To(HaveLen(4), "expected 4 NUMA cells (2 CPU + 2 GPU)")

			for _, cell := range domainSpec.CPU.NUMA.Cells[2:] {
				Expect(cell.CPUs).To(Equal(""), "GPU NUMA cell should have no CPUs")
			}

			// Verify the two GPU devices are on different expander buses
			Expect(domainSpec.Devices.HostDevices[0].Address.Bus).ToNot(
				Equal(domainSpec.Devices.HostDevices[1].Address.Bus),
				"GPUs should be on separate expander buses")
		})
	})
})
