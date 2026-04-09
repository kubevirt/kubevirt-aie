package hostdevice

import (
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/util/hardware"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

func mustApplyNUMAHostDeviceTopology(t *testing.T, vmi *v1.VirtualMachineInstance, domain *api.Domain) {
	t.Helper()
	if err := ApplyNUMAHostDeviceTopology(vmi, domain); err != nil {
		t.Fatalf("ApplyNUMAHostDeviceTopology returned error: %v", err)
	}
}

func mustPrepareGraceGuestNUMATopology(t *testing.T, vmi *v1.VirtualMachineInstance, domain *api.Domain) {
	t.Helper()
	if err := PrepareGraceGuestNUMATopology(vmi, domain); err != nil {
		t.Fatalf("PrepareGraceGuestNUMATopology returned error: %v", err)
	}
}

func mustNUMACellByID(t *testing.T, domain *api.Domain, id string) api.NUMACell {
	t.Helper()
	if domain == nil || domain.Spec.CPU.NUMA == nil {
		t.Fatalf("domain NUMA topology is nil")
	}
	for _, cell := range domain.Spec.CPU.NUMA.Cells {
		if cell.ID == id {
			return cell
		}
	}
	t.Fatalf("NUMA cell %s not found", id)
	return api.NUMACell{}
}

func mustNUMADistanceValue(t *testing.T, cell api.NUMACell, dst string) uint64 {
	t.Helper()
	if cell.Distances == nil {
		t.Fatalf("NUMA cell %s has no distances", cell.ID)
	}
	for _, sibling := range cell.Distances.Siblings {
		if sibling.ID == dst {
			return sibling.Value
		}
	}
	t.Fatalf("NUMA cell %s has no sibling distance for %s", cell.ID, dst)
	return 0
}

func mustHostDeviceNodeSets(t *testing.T, domain *api.Domain) []string {
	t.Helper()
	nodeSets := make([]string, 0, len(domain.Spec.Devices.HostDevices))
	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		if dev.ACPI == nil {
			t.Fatalf("host device %d has no ACPI nodeset", i)
		}
		nodeSets = append(nodeSets, dev.ACPI.NodeSet)
	}
	return nodeSets
}

func TestNormalizeHotplugRootPortAlias(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":                            "",
		"hotplug-rp-numa-0":           "hotplug-rp-numa-0",
		"ua-hotplug-rp-numa-1":        "hotplug-rp-numa-1",
		"ua-ua-hotplug-rp-numa-2":     "hotplug-rp-numa-2",
		"ua-hotplug-rp-legacy":        "hotplug-rp-legacy",
		"ua-ua-ua-hotplug-rp-random":  "hotplug-rp-random",
		"unrelated-prefix-hotplug-rp": "unrelated-prefix-hotplug-rp",
		"ua-not-hotplug":              "not-hotplug",
		"ua-ua-not-hotplug":           "not-hotplug",
	}
	for input, expected := range cases {
		if got := NormalizeHotplugRootPortAlias(input); got != expected {
			t.Fatalf("NormalizeHotplugRootPortAlias(%q) = %q, expected %q", input, got, expected)
		}
	}
}

func TestIsHotplugRootPortAlias(t *testing.T) {
	t.Parallel()
	cases := []struct {
		alias    string
		expected bool
	}{
		{alias: "", expected: false},
		{alias: "hotplug-rp-numa-0", expected: true},
		{alias: "ua-hotplug-rp-numa-1", expected: true},
		{alias: "ua-ua-hotplug-rp-numa-2", expected: true},
		{alias: fmt.Sprintf("%s0", NUMAHotplugRootPortAliasPrefix), expected: true},
		{alias: "something-else", expected: false},
		{alias: "ua-something-else", expected: false},
	}
	for _, tc := range cases {
		if got := IsHotplugRootPortAlias(tc.alias); got != tc.expected {
			t.Fatalf("IsHotplugRootPortAlias(%q) = %t, expected %t", tc.alias, got, tc.expected)
		}
	}
}

func TestIsNUMARootPortAlias(t *testing.T) {
	t.Parallel()
	cases := []struct {
		alias    string
		expected bool
	}{
		{alias: "", expected: false},
		{alias: "ua-numa-rp-1234", expected: true},
		{alias: "numa-rp-abcdef", expected: true},
		{alias: "hotplug-rp-numa-0", expected: false},
		{alias: "ua-hotplug-rp-numa-1", expected: false},
		{alias: "random", expected: false},
	}
	for _, tc := range cases {
		if got := IsNUMARootPortAlias(tc.alias); got != tc.expected {
			t.Fatalf("IsNUMARootPortAlias(%q) = %t, expected %t", tc.alias, got, tc.expected)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyDisabled(t *testing.T) {
	defer restoreNUMAHelpers()
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		t.Fatal("GetDeviceNumaNodeInt should not be called when feature disabled")
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{}
	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0"},
						{ID: "1"},
					},
				},
			},
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
			NUMATune: &api.NUMATune{
				MemNodes: []api.MemNode{
					{CellID: 0, Mode: "strict", NodeSet: "0"},
					{CellID: 1, Mode: "strict", NodeSet: "1"},
				},
			},
		},
	}

	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})
	stubPCIPath("0000:02:00.0", []string{"0000:00:02.0", "0000:02:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged, got %d", len(domain.Spec.Devices.Controllers))
	}
	if domain.Spec.Devices.HostDevices[0].Address != nil {
		t.Fatalf("expected host device address to remain unset when feature disabled")
	}
}

func TestApplyNUMAHostDeviceTopologyCreatesPXBs(t *testing.T) {
	defer restoreNUMAHelpers()

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:01:00.0":
			return 0, nil
		case "0000:02:00.0":
			return 1, nil
		default:
			return -1, nil
		}
	}
	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:" + strings.TrimPrefix(addr.Bus, "0x") + ":" + strings.TrimPrefix(addr.Slot, "0x") + ".0", nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x02",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})
	stubPCIPath("0000:02:00.0", []string{"0000:00:02.0", "0000:02:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var pxbCount int
	var numaNodes []int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
			if ctrl.Target == nil || ctrl.Target.Node == nil {
				t.Fatalf("expected PXB controller to have target node")
			}
			numaNodes = append(numaNodes, *ctrl.Target.Node)
		}
	}
	if pxbCount != 2 {
		t.Fatalf("expected two expander buses, got %d", pxbCount)
	}
	if !(containsInt(numaNodes, 0) && containsInt(numaNodes, 1)) {
		t.Fatalf("expected expander buses for NUMA nodes 0 and 1, got %v", numaNodes)
	}

	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("expected host device %d to have an address assigned", i)
		}
		// devices should use Bus attribute, not controller
		if dev.Address.Bus == "" {
			t.Fatalf("expected host device %d to have bus assigned (NVIDIA docs pattern)", i)
		}
		if dev.Address.Controller != "" {
			t.Fatalf("expected host device %d to leave controller empty (using Bus instead), got %s", i, dev.Address.Controller)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyUsesDedicatedPXBForLargeMMIODevices(t *testing.T) {
	defer restoreNUMAHelpers()

	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:" + strings.TrimPrefix(addr.Bus, "0x") + ":" + strings.TrimPrefix(addr.Slot, "0x") + ".0", nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x01"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x02"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})
	stubPCIPath("0000:02:00.0", []string{"0000:00:02.0", "0000:02:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var pxbBuses []int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-expander-bus" {
			continue
		}
		if ctrl.Target == nil || ctrl.Target.Node == nil || *ctrl.Target.Node != 0 {
			t.Fatalf("expected dedicated PXB to target NUMA node 0, got %+v", ctrl.Target)
		}
		busNr, err := strconv.Atoi(ctrl.Target.BusNr)
		if err != nil {
			t.Fatalf("failed to parse dedicated PXB busNr %q: %v", ctrl.Target.BusNr, err)
		}
		pxbBuses = append(pxbBuses, busNr)
	}

	if len(pxbBuses) != 2 {
		t.Fatalf("expected two dedicated PXB controllers for large-MMIO devices, got %d", len(pxbBuses))
	}
	if pxbBuses[0] == pxbBuses[1] {
		t.Fatalf("expected dedicated PXBs to use distinct bus numbers, got %v", pxbBuses)
	}
}

func TestAllocateDedicatedPXBBusNumberAvoidsAdjacentBusConflicts(t *testing.T) {
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
			},
		},
	}
	planner := newNUMAPCIPlanner(domain)

	baseBus := pxbBusNumberBase + pxbBusNumberSpacing // NUMA node 1 base, 0x40
	planner.reservePCIBus(baseBus)

	first, err := planner.allocateDedicatedPXBBusNumber(baseBus)
	if err != nil {
		t.Fatalf("failed to allocate first dedicated PXB bus: %v", err)
	}
	planner.reservePCIBus(first)

	second, err := planner.allocateDedicatedPXBBusNumber(baseBus)
	if err != nil {
		t.Fatalf("failed to allocate second dedicated PXB bus: %v", err)
	}

	if second-first == 1 {
		t.Fatalf("expected non-adjacent dedicated PXB buses, got 0x%02x and 0x%02x", first, second)
	}
	if first != baseBus+1 || second != baseBus+3 {
		t.Fatalf("expected dedicated PXB buses 0x%02x and 0x%02x, got 0x%02x and 0x%02x",
			baseBus+1, baseBus+3, first, second)
	}
}

func TestReservePXBBusNumbersReservesSharedRootPortWindowBeforeDedicatedPXB(t *testing.T) {
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
			},
		},
	}
	planner := newNUMAPCIPlanner(domain)

	// NUMA node 1 with:
	// - 2 shared groups (pxbGroup="") that will live behind the shared PXB
	// - 2 dedicated groups (pxbGroup!= "") for large-MMIO devices (e.g. GPUs)
	groupKeys := []deviceGroupKey{
		{hostNUMANode: 1, guestNUMANode: 1, pathKey: "ib-a", pxbGroup: ""},
		{hostNUMANode: 1, guestNUMANode: 1, pathKey: "ib-b", pxbGroup: ""},
		{hostNUMANode: 1, guestNUMANode: 1, pathKey: "gpu-a", pxbGroup: "gpu-a"},
		{hostNUMANode: 1, guestNUMANode: 1, pathKey: "gpu-b", pxbGroup: "gpu-b"},
	}
	planner.reservePXBBusNumbers(groupKeys)

	baseBus := pxbBusNumberBase + pxbBusNumberSpacing // NUMA node 1 base, 0x40
	for _, bus := range []int{baseBus, baseBus + 1, baseBus + 2} {
		if _, used := planner.usedPCIBuses[bus]; !used {
			t.Fatalf("expected bus 0x%02x to be pre-reserved", bus)
		}
	}

	firstDedicated, err := planner.allocateDedicatedPXBBusNumber(baseBus)
	if err != nil {
		t.Fatalf("failed to allocate first dedicated PXB bus: %v", err)
	}
	if firstDedicated != baseBus+3 {
		t.Fatalf("expected first dedicated PXB bus 0x%02x, got 0x%02x", baseBus+3, firstDedicated)
	}
	planner.reservePCIBus(firstDedicated)

	secondDedicated, err := planner.allocateDedicatedPXBBusNumber(baseBus)
	if err != nil {
		t.Fatalf("failed to allocate second dedicated PXB bus: %v", err)
	}
	if secondDedicated != baseBus+5 {
		t.Fatalf("expected second dedicated PXB bus 0x%02x, got 0x%02x", baseBus+5, secondDedicated)
	}
}

func TestApplyNUMAHostDeviceTopologySingleGuestCellPreservesHostNUMA(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:03:00.0", "0000:04:00.0":
			return 0, nil
		case "0000:83:00.0", "0000:84:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{
							ID:     "0",
							CPUs:   "0-19",
							Memory: 5242880,
							Unit:   "KiB",
						},
					},
				},
			},
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu1", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu2", "0x0000", "0x04"),
					newTestPCIHostDevice("gpu3", "0x0000", "0x83"),
					newTestPCIHostDevice("gpu4", "0x0000", "0x84"),
				},
			},
		},
	}

	domain.Spec.NUMATune = &api.NUMATune{
		MemNodes: []api.MemNode{
			{CellID: 0, Mode: "strict", NodeSet: "0"},
		},
	}
	stubPCIPath("0000:03:00.0", []string{"0000:00:03.0", "0000:03:00.0"})
	stubPCIPath("0000:04:00.0", []string{"0000:00:04.0", "0000:04:00.0"})
	stubPCIPath("0000:83:00.0", []string{"0000:80:83.0", "0000:83:00.0"})
	stubPCIPath("0000:84:00.0", []string{"0000:80:84.0", "0000:84:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			t.Fatalf("expected NUMA planner to skip when guest exposes a single NUMA cell, found expander bus %v", ctrl)
		}
	}
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address != nil {
			t.Fatalf("expected host device %d to retain unmanaged address when guest NUMA topology lacks multiple cells", i)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyGroupsByTopologyWithinNUMA(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		return fmt.Sprintf("%s:%s:00.0", domain, bus), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }

	groupMap := map[string]string{
		"0000:03:00.0": "switch-a",
		"0000:04:00.0": "switch-b",
		"0000:05:00.0": "switch-a",
	}
	getDevicePCIProximityGroupFunc = func(bdf string) (string, error) {
		if group, ok := groupMap[bdf]; ok {
			return group, nil
		}
		return "default", nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu1", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu2", "0x0000", "0x04"),
					newTestPCIHostDevice("gpu3", "0x0000", "0x05"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:01:00.0", "0000:03:00.0"})
	stubPCIPath("0000:04:00.0", []string{"0000:00:02.0", "0000:02:00.0", "0000:04:00.0"})
	stubPCIPath("0000:05:00.0", []string{"0000:00:01.0", "0000:01:00.0", "0000:05:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var pxbControllers []api.Controller
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbControllers = append(pxbControllers, ctrl)
		}
	}
	if len(pxbControllers) != 1 {
		t.Fatalf("expected a single PXB for NUMA node 0, got %d", len(pxbControllers))
	}
	if pxbControllers[0].Target == nil || pxbControllers[0].Target.Node == nil || *pxbControllers[0].Target.Node != 0 {
		t.Fatalf("expected PXB to target NUMA node 0, got %+v", pxbControllers[0].Target)
	}

	rootPorts := make(map[string]*api.Address)
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-root-port" && ctrl.Address != nil &&
			ctrl.Alias != nil && strings.HasPrefix(ctrl.Alias.GetName(), numaRootPortPrefix) {
			rootPorts[ctrl.Index] = ctrl.Address
		}
	}
	if len(rootPorts) != 2 {
		t.Fatalf("expected two root ports (one per unique upstream path), got %d", len(rootPorts))
	}

	gpu1 := domain.Spec.Devices.HostDevices[0]
	gpu2 := domain.Spec.Devices.HostDevices[1]
	gpu3 := domain.Spec.Devices.HostDevices[2]

	if gpu1.Address == nil || gpu2.Address == nil || gpu3.Address == nil {
		t.Fatalf("expected all devices to receive guest addresses")
	}
}

func TestApplyNUMAHostDeviceTopologyHandlesMdev(t *testing.T) {
	defer restoreNUMAHelpers()

	getMdevParentPCIAddressFunc = func(uuid string) (string, error) {
		if uuid != "mdev-uuid" {
			t.Fatalf("unexpected mdev uuid %s", uuid)
		}
		return "0000:03:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		if bdf == "0000:03:00.0" {
			return 0, nil
		}
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDeviceMDev,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								UUID: "mdev-uuid",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var pxbCount int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
		}
	}
	if pxbCount != 1 {
		t.Fatalf("expected a single expander bus for the mdev device, got %d", pxbCount)
	}
	if domain.Spec.Devices.HostDevices[0].Address == nil {
		t.Fatalf("expected host dev address assigned for mdev device")
	}
	if domain.Spec.Devices.HostDevices[0].Source.Address.UUID != "mdev-uuid" {
		t.Fatalf("expected source address UUID to remain unchanged")
	}
}

func TestApplyNUMAHostDeviceTopologyInjectsArm64GraceHostDeviceSettings(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:03:00.0":
			return 0, nil
		case "0000:83:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x83"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	for node := 2; node <= 16; node++ {
		appendGuestNUMACells(domain, node)
	}
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})
	stubPCIPath("0000:83:00.0", []string{"0000:80:01.0", "0000:83:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	expectedNodeSets := []string{"0", "1"}
	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		if dev.Driver != nil && dev.Driver.IOMMUFD != "" {
			t.Fatalf("expected host device %d to skip iommufd for non-large MMIO device, got %q", i, dev.Driver.IOMMUFD)
		}
		if dev.ACPI == nil || dev.ACPI.NodeSet != expectedNodeSets[i] {
			t.Fatalf("expected host device %d to have ACPI nodeset %s, got %+v", i, expectedNodeSets[i], dev.ACPI)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyInjectsArm64GraceHostDeviceSettingsFromGraceVirtualizationAnnotation(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:03:00.0":
			return 0, nil
		case "0000:83:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x83"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	for node := 2; node <= 16; node++ {
		appendGuestNUMACells(domain, node)
	}
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})
	stubPCIPath("0000:83:00.0", []string{"0000:80:01.0", "0000:83:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	expectedNodeSets := []string{"0", "1"}
	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		if dev.Driver != nil && dev.Driver.IOMMUFD != "" {
			t.Fatalf("expected host device %d to skip iommufd for non-large MMIO device, got %q", i, dev.Driver.IOMMUFD)
		}
		if dev.ACPI == nil || dev.ACPI.NodeSet != expectedNodeSets[i] {
			t.Fatalf("expected host device %d to have ACPI nodeset %s, got %+v", i, expectedNodeSets[i], dev.ACPI)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyAcceptsGraceVirtualizationAnnotationWithSMMUv3Fields(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return false }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true,"vcmdq":true,"egm":false}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	dev := domain.Spec.Devices.HostDevices[0]
	if dev.Driver != nil && dev.Driver.IOMMUFD != "" {
		t.Fatalf("expected host device to skip iommufd injection for non-large MMIO device, got %q", dev.Driver.IOMMUFD)
	}
	if dev.ACPI == nil || dev.ACPI.NodeSet != "0" {
		t.Fatalf("expected host device ACPI nodeset to be injected, got %+v", dev.ACPI)
	}
}

func TestApplyNUMAHostDeviceTopologySkipsGraceHostDeviceSettingsWithoutGraceVirtualizationAnnotation(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				IOMMUs: []api.IOMMU{
					{
						Model: "smmuv3",
						Driver: &api.IOMMUDriver{
							PCIBus: "7",
						},
					},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
			QEMUCmd: &api.Commandline{
				QEMUArg: []api.Arg{
					{Value: "-device"},
					{Value: "arm-smmuv3,id=smmu-test"},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	dev := domain.Spec.Devices.HostDevices[0]
	if dev.Driver != nil {
		t.Fatalf("expected iommufd injection to be skipped without graceVirtualization annotation, got %+v", dev.Driver)
	}
	if dev.ACPI != nil {
		t.Fatalf("expected ACPI nodeset injection to be skipped without graceVirtualization annotation, got %+v", dev.ACPI)
	}
	if len(domain.Spec.Devices.IOMMUs) != 1 || domain.Spec.Devices.IOMMUs[0].Model != "smmuv3" {
		t.Fatalf("expected existing smmuv3 iommu entries to be preserved without graceVirtualization annotation, got %+v", domain.Spec.Devices.IOMMUs)
	}
	if !hasArmSMMUv3QEMUArgs(domain) {
		t.Fatalf("expected existing arm-smmuv3 qemu args to be preserved without graceVirtualization annotation")
	}
}

func hasArmSMMUv3QEMUArgs(domain *api.Domain) bool {
	if domain == nil || domain.Spec.QEMUCmd == nil {
		return false
	}
	for i := range domain.Spec.QEMUCmd.QEMUArg {
		if strings.Contains(domain.Spec.QEMUCmd.QEMUArg[i].Value, "arm-smmuv3") {
			return true
		}
	}
	return false
}

func hasQEMUOverrideProperty(domain *api.Domain, alias, name, propType, value string) bool {
	if domain == nil || domain.Spec.QEMUOverride == nil {
		return false
	}
	for _, device := range domain.Spec.QEMUOverride.Devices {
		if device.Alias != alias {
			continue
		}
		for _, prop := range device.Frontend.Properties {
			if prop.Name == name && prop.Type == propType && prop.Value == value {
				return true
			}
		}
	}
	return false
}

func TestApplyNUMAHostDeviceTopologyInjectsSMMUv3IOMMUsAndVCMDQ(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	readPCIDeviceFileFunc = func(path string) ([]byte, error) {
		switch {
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_speed"),
			strings.HasSuffix(path, "/0000:03:00.0/max_link_speed"):
			return []byte("32.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_width"),
			strings.HasSuffix(path, "/0000:03:00.0/max_link_width"):
			return []byte("16\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true,"vcmdq":true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	if len(domain.Spec.Devices.IOMMUs) == 0 {
		t.Fatalf("expected libvirt iommu entries for smmuv3")
	}

	var hasCMDQVEnabled bool
	for _, iommu := range domain.Spec.Devices.IOMMUs {
		if iommu.Model != "smmuv3" {
			t.Fatalf("expected iommu model smmuv3, got %q", iommu.Model)
		}
		if iommu.Driver == nil {
			t.Fatalf("expected iommu driver settings")
		}
		if iommu.Driver.PCIBus == "" {
			t.Fatalf("expected iommu driver pciBus to be set")
		}
		if iommu.Driver.CMDQV == "on" {
			hasCMDQVEnabled = true
		}
	}
	if !hasCMDQVEnabled {
		t.Fatalf("expected at least one smmuv3 iommu with cmdqv=on")
	}
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-expander-bus" {
			continue
		}
		if ctrl.Alias != nil {
			t.Fatalf("expected Grace smmuv3 flow to keep default PXB ids (no user alias), got alias %q", ctrl.Alias.GetName())
		}
	}
	if hasArmSMMUv3QEMUArgs(domain) {
		t.Fatalf("did not expect raw arm-smmuv3 qemu args when using libvirt-native smmuv3 iommu entries")
	}
	var hasRootPortLinkSpeed bool
	var hasRootPortLinkWidth bool
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-root-port" || ctrl.Alias == nil {
			continue
		}
		qemuID := api.UserAliasPrefix + ctrl.Alias.GetName()
		if hasQEMUOverrideProperty(domain, qemuID, "x-speed", "string", "32") {
			hasRootPortLinkSpeed = true
		}
		if hasQEMUOverrideProperty(domain, qemuID, "x-width", "string", "16") {
			hasRootPortLinkWidth = true
		}
	}
	if !hasRootPortLinkSpeed {
		t.Fatalf("expected derived per-root-port x-speed qemu override property to be injected")
	}
	if !hasRootPortLinkWidth {
		t.Fatalf("expected derived per-root-port x-width qemu override property to be injected")
	}
}

func TestApplyNUMAHostDeviceTopologyDerivesRootPortLinkFromHostPathBottleneck(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	readPCIDeviceFileFunc = func(path string) ([]byte, error) {
		switch {
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_speed"):
			return []byte("16.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:03:00.0/max_link_speed"):
			return []byte("32.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_width"):
			return []byte("8\n"), nil
		case strings.HasSuffix(path, "/0000:03:00.0/max_link_width"):
			return []byte("16\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("nic0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var hasBottleneckSpeed bool
	var hasBottleneckWidth bool
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-root-port" || ctrl.Alias == nil {
			continue
		}
		qemuID := api.UserAliasPrefix + ctrl.Alias.GetName()
		if hasQEMUOverrideProperty(domain, qemuID, "x-speed", "string", "16") {
			hasBottleneckSpeed = true
		}
		if hasQEMUOverrideProperty(domain, qemuID, "x-width", "string", "8") {
			hasBottleneckWidth = true
		}
	}
	if !hasBottleneckSpeed {
		t.Fatalf("expected root port x-speed to follow the host path bottleneck")
	}
	if !hasBottleneckWidth {
		t.Fatalf("expected root port x-width to follow the host path bottleneck")
	}
}

func TestReadPCIeLinkCharacteristicsPreferCurrentOverMax(t *testing.T) {
	defer restoreNUMAHelpers()

	readPCIDeviceFileFunc = func(path string) ([]byte, error) {
		switch {
		case strings.HasSuffix(path, "/0000:00:01.0/current_link_speed"):
			return []byte("16.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_speed"):
			return []byte("32.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/current_link_width"):
			return []byte("8\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_width"):
			return []byte("16\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	if got := readPCIeLinkSpeedForBDF("0000:00:01.0"); got != "16" {
		t.Fatalf("expected current link speed to be preferred, got %q", got)
	}
	if got := readPCIeLinkWidthForBDF("0000:00:01.0"); got != "8" {
		t.Fatalf("expected current link width to be preferred, got %q", got)
	}
}

func TestReadPCIeLinkCharacteristicsFallbackToMax(t *testing.T) {
	defer restoreNUMAHelpers()

	readPCIDeviceFileFunc = func(path string) ([]byte, error) {
		switch {
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_speed"):
			return []byte("32.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_width"):
			return []byte("16\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	if got := readPCIeLinkSpeedForBDF("0000:00:01.0"); got != "32" {
		t.Fatalf("expected max link speed fallback, got %q", got)
	}
	if got := readPCIeLinkWidthForBDF("0000:00:01.0"); got != "16" {
		t.Fatalf("expected max link width fallback, got %q", got)
	}
}

func TestApplyNUMAHostDeviceTopologyIgnoresDownstreamSwitchBottlenecksForRootPortModeling(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	readPCIDeviceFileFunc = func(path string) ([]byte, error) {
		switch {
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_speed"),
			strings.HasSuffix(path, "/0000:01:00.0/max_link_speed"):
			return []byte("32.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:02:00.0/max_link_speed"),
			strings.HasSuffix(path, "/0000:02:01.0/max_link_speed"),
			strings.HasSuffix(path, "/0000:03:00.0/max_link_speed"),
			strings.HasSuffix(path, "/0000:04:00.0/max_link_speed"):
			return []byte("16.0 GT/s PCIe\n"), nil
		case strings.HasSuffix(path, "/0000:00:01.0/max_link_width"),
			strings.HasSuffix(path, "/0000:01:00.0/max_link_width"):
			return []byte("16\n"), nil
		case strings.HasSuffix(path, "/0000:02:00.0/max_link_width"),
			strings.HasSuffix(path, "/0000:02:01.0/max_link_width"),
			strings.HasSuffix(path, "/0000:03:00.0/max_link_width"),
			strings.HasSuffix(path, "/0000:04:00.0/max_link_width"):
			return []byte("8\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x04"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0"})
	stubPCIPath("0000:04:00.0", []string{"0000:00:01.0", "0000:01:00.0", "0000:02:01.0", "0000:04:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var hasSharedUpstreamSpeed bool
	var hasSharedUpstreamWidth bool
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-root-port" || ctrl.Alias == nil {
			continue
		}
		qemuID := api.UserAliasPrefix + ctrl.Alias.GetName()
		if hasQEMUOverrideProperty(domain, qemuID, "x-speed", "string", "32") {
			hasSharedUpstreamSpeed = true
		}
		if hasQEMUOverrideProperty(domain, qemuID, "x-width", "string", "16") {
			hasSharedUpstreamWidth = true
		}
	}
	if !hasSharedUpstreamSpeed {
		t.Fatalf("expected root port x-speed to follow the shared upstream link, not slower downstream switch links")
	}
	if !hasSharedUpstreamWidth {
		t.Fatalf("expected root port x-width to follow the shared upstream link, not narrower downstream switch links")
	}
}

func TestApplyNUMAHostDeviceTopologyScopesVCMDQToDedicatedPXBBuses(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	getDevicePCITotalMMIOSizeFunc = func(bdf string) (uint64, error) {
		if bdf == "0000:08:00.0" {
			return largeMMIOPXBIsolationThreshold, nil
		}
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true,"vcmdq":true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("ib0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu0", "0x0000", "0x08"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})
	stubPCIPath("0000:08:00.0", []string{"0000:00:02.0", "0000:08:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var pxbControllers int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbControllers++
		}
	}
	if pxbControllers != 2 {
		t.Fatalf("expected isolated PXB hierarchies per device in mixed Grace topology, got %d controllers", pxbControllers)
	}

	if len(domain.Spec.Devices.IOMMUs) == 0 {
		t.Fatalf("expected at least one smmuv3 iommu entry, got %d", len(domain.Spec.Devices.IOMMUs))
	}

	var cmdqvOn, cmdqvOff int
	var accelOn, accelOff int
	var rilOffOnDedicated int
	for _, iommu := range domain.Spec.Devices.IOMMUs {
		if iommu.Model != "smmuv3" || iommu.Driver == nil {
			continue
		}
		if iommu.Driver.CMDQV == "on" {
			cmdqvOn++
		} else {
			cmdqvOff++
		}
		if iommu.Driver.Accel == "on" {
			accelOn++
			if iommu.Driver.CMDQV == "on" {
				if iommu.Driver.RIL == "off" && iommu.Driver.ATS == "on" && iommu.Driver.PASID == "on" && iommu.Driver.OAS == "48" {
					rilOffOnDedicated++
				}
			}
		} else if iommu.Driver.Accel == "off" {
			accelOff++
		}
	}

	if cmdqvOn == 0 {
		t.Fatalf("expected cmdqv enabled on at least one dedicated PXB bus")
	}
	if cmdqvOff != 0 {
		t.Fatalf("did not expect shared/non-dedicated smmuv3 entries in mixed topology")
	}
	if accelOn == 0 {
		t.Fatalf("expected accel=on on at least one dedicated PXB bus")
	}
	if accelOff != 0 {
		t.Fatalf("did not expect accel=off when /dev/iommu is available")
	}
	if rilOffOnDedicated == 0 {
		t.Fatalf("expected dedicated smmuv3 buses to keep accel features (ats=on,pasid=on,ril=off,oas=48)")
	}
}

func TestApplyNUMAHostDeviceTopologyInjectsSMMUv3IOMMUsWithoutVCMDQByDefault(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	if len(domain.Spec.Devices.IOMMUs) == 0 {
		t.Fatalf("expected libvirt iommu entries for smmuv3")
	}

	for _, iommu := range domain.Spec.Devices.IOMMUs {
		if iommu.Driver != nil && iommu.Driver.CMDQV == "on" {
			t.Fatalf("did not expect cmdqv=on unless explicitly requested")
		}
	}
	if hasArmSMMUv3QEMUArgs(domain) {
		t.Fatalf("did not expect raw arm-smmuv3 qemu args when using libvirt-native smmuv3 iommu entries")
	}
}

func TestApplyNUMAHostDeviceTopologyFallsBackToSMMUv3AccelOffWithoutIOMMUDevice(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}
	isIOMMUFDDeviceAvailableFunc = func() bool {
		return false
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("ib0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	if len(domain.Spec.Devices.IOMMUs) == 0 {
		t.Fatalf("expected smmuv3 iommu entries to be generated")
	}
	for _, iommu := range domain.Spec.Devices.IOMMUs {
		if iommu.Driver == nil {
			t.Fatalf("expected iommu driver settings")
		}
		if iommu.Driver.Accel != "off" {
			t.Fatalf("expected smmuv3 accel=off when /dev/iommu is unavailable, got %q", iommu.Driver.Accel)
		}
		if iommu.Driver.RIL != "on" {
			t.Fatalf("expected smmuv3 ril=on when accel=off, got %q", iommu.Driver.RIL)
		}
		if iommu.Driver.ATS != "" || iommu.Driver.PASID != "" || iommu.Driver.CMDQV != "" || iommu.Driver.OAS != "" {
			t.Fatalf("expected accel=off buses to omit ats/pasid/cmdqv/oas, got %+v", iommu.Driver)
		}
	}
	for _, dev := range domain.Spec.Devices.HostDevices {
		if dev.Driver != nil {
			t.Fatalf("expected host device driver to be omitted when /dev/iommu is unavailable, got %+v", dev.Driver)
		}
	}
}

func TestCollectNUMAPXBPciBusesUsesControllerIndices(t *testing.T) {
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{
						Type:  "pci",
						Index: "2",
						Model: "pcie-expander-bus",
						Target: &api.ControllerTarget{
							BusNr: "33",
						},
					},
					{
						Type:  "pci",
						Index: "6",
						Model: "pcie-expander-bus",
						Alias: api.NewUserDefinedAlias("numa-pxb-1-1-b2"),
						Target: &api.ControllerTarget{
							BusNr: "65",
						},
					},
				},
			},
		},
	}

	got := collectNUMAPXBPciBuses(domain)
	expected := []string{"2", "6"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected PXB bus IDs %v, got %v", expected, got)
	}
}

func TestApplyNUMAHostDeviceTopologyInjectsArm64GraceGINodeSetsForLargeMMIOGPUs(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:03:00.0":
			return 0, nil
		case "0000:83:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x83"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})
	stubPCIPath("0000:83:00.0", []string{"0000:80:01.0", "0000:83:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	expectedNodeSets := []string{"2-9", "10-17"}
	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		if dev.Driver == nil || dev.Driver.IOMMUFD != "yes" {
			t.Fatalf("expected host device %d to have iommufd enabled", i)
		}
		if dev.ACPI == nil || dev.ACPI.NodeSet != expectedNodeSets[i] {
			t.Fatalf("expected host device %d to have ACPI nodeset %s, got %+v", i, expectedNodeSets[i], dev.ACPI)
		}
	}

	for id := 0; id <= 17; id++ {
		idStr := strconv.Itoa(id)
		found := false
		for _, cell := range domain.Spec.CPU.NUMA.Cells {
			if cell.ID != idStr {
				continue
			}
			found = true
			if id >= 2 {
				if cell.Memory != 0 || cell.Unit != "KiB" {
					t.Fatalf("expected GI NUMA cell %d to be zero-memory KiB cell, got memory=%d unit=%q", id, cell.Memory, cell.Unit)
				}
			}
			break
		}
		if !found {
			t.Fatalf("expected guest NUMA cell %d to exist after GI allocation", id)
		}
	}

	cell0 := mustNUMACellByID(t, domain, "0")
	if got := mustNUMADistanceValue(t, cell0, "1"); got != graceNUMADistanceRemoteNode {
		t.Fatalf("expected cell 0 -> 1 distance %d, got %d", graceNUMADistanceRemoteNode, got)
	}
	if got := mustNUMADistanceValue(t, cell0, "2"); got != graceNUMADistanceLocalToGPU {
		t.Fatalf("expected cell 0 -> 2 distance %d, got %d", graceNUMADistanceLocalToGPU, got)
	}
	if got := mustNUMADistanceValue(t, cell0, "10"); got != graceNUMADistanceRemoteToGPU {
		t.Fatalf("expected cell 0 -> 10 distance %d, got %d", graceNUMADistanceRemoteToGPU, got)
	}

	cell2 := mustNUMACellByID(t, domain, "2")
	if got := mustNUMADistanceValue(t, cell2, "0"); got != graceNUMADistanceLocalToGPU {
		t.Fatalf("expected cell 2 -> 0 distance %d, got %d", graceNUMADistanceLocalToGPU, got)
	}
	if got := mustNUMADistanceValue(t, cell2, "3"); got != graceNUMADistanceSameGPUGroup {
		t.Fatalf("expected cell 2 -> 3 distance %d, got %d", graceNUMADistanceSameGPUGroup, got)
	}
	if got := mustNUMADistanceValue(t, cell2, "10"); got != graceNUMADistanceRemoteNode {
		t.Fatalf("expected cell 2 -> 10 distance %d, got %d", graceNUMADistanceRemoteNode, got)
	}
}

func TestApplyNUMAHostDeviceTopologyInjectsIOMMUFDOnlyForLargeMMIOGraceDevices(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:03:00.0", "0000:08:00.0":
			return 0, nil
		case "0000:18:00.0", "0000:83:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}
	getDevicePCITotalMMIOSizeFunc = func(bdf string) (uint64, error) {
		switch bdf {
		case "0000:08:00.0", "0000:18:00.0":
			return largeMMIOPXBIsolationThreshold, nil
		default:
			return 0, nil
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("ib0", "0x0000", "0x03"),
					newTestPCIHostDevice("gpu0", "0x0000", "0x08"),
					newTestPCIHostDevice("ib1", "0x0000", "0x83"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x18"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	stubPCIPath("0000:03:00.0", []string{"0000:00:03.0", "0000:03:00.0"})
	stubPCIPath("0000:08:00.0", []string{"0000:00:08.0", "0000:08:00.0"})
	stubPCIPath("0000:18:00.0", []string{"0000:00:18.0", "0000:18:00.0"})
	stubPCIPath("0000:83:00.0", []string{"0000:80:03.0", "0000:83:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	type expectedHostDevice struct {
		nodeSet string
		iommufd string
	}
	expectedByBDF := map[string]expectedHostDevice{
		"0000:03:00.0": {nodeSet: "0", iommufd: ""},
		"0000:08:00.0": {nodeSet: "2-9", iommufd: "yes"},
		"0000:18:00.0": {nodeSet: "10-17", iommufd: "yes"},
		"0000:83:00.0": {nodeSet: "1", iommufd: ""},
	}

	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		bdf := fmt.Sprintf("%s:%s:00.0",
			strings.TrimPrefix(dev.Source.Address.Domain, "0x"),
			strings.TrimPrefix(dev.Source.Address.Bus, "0x"),
		)
		expected, ok := expectedByBDF[bdf]
		if !ok {
			t.Fatalf("unexpected host device bdf %s", bdf)
		}

		if dev.ACPI == nil || dev.ACPI.NodeSet != expected.nodeSet {
			t.Fatalf("expected host device %s to have ACPI nodeset %s, got %+v", bdf, expected.nodeSet, dev.ACPI)
		}

		if expected.iommufd == "" {
			if dev.Driver != nil {
				t.Fatalf("expected host device %s to have no driver element when iommufd is disabled, got %+v", bdf, dev.Driver)
			}
			continue
		}
		if dev.Driver == nil || dev.Driver.IOMMUFD != expected.iommufd {
			t.Fatalf("expected host device %s to have iommufd %q, got %+v", bdf, expected.iommufd, dev.Driver)
		}
	}
}

func TestApplyNUMAHostDeviceTopologyAssignsUniqueGINodeSetsPerLargeMMIOGPU(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:08:00.0", "0000:09:00.0":
			return 0, nil
		case "0000:18:00.0", "0000:19:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x19"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x08"),
					newTestPCIHostDevice("gpu2", "0x0000", "0x18"),
					newTestPCIHostDevice("gpu3", "0x0000", "0x09"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	stubPCIPath("0000:08:00.0", []string{"0000:00:08.0", "0000:08:00.0"})
	stubPCIPath("0000:09:00.0", []string{"0000:00:09.0", "0000:09:00.0"})
	stubPCIPath("0000:18:00.0", []string{"0000:00:18.0", "0000:18:00.0"})
	stubPCIPath("0000:19:00.0", []string{"0000:00:19.0", "0000:19:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	expectedByBDF := map[string]string{
		"0000:08:00.0": "2-9",
		"0000:09:00.0": "10-17",
		"0000:18:00.0": "18-25",
		"0000:19:00.0": "26-33",
	}
	seenNodeSets := make(map[string]struct{}, len(domain.Spec.Devices.HostDevices))
	for i := range domain.Spec.Devices.HostDevices {
		dev := domain.Spec.Devices.HostDevices[i]
		if dev.Driver == nil || dev.Driver.IOMMUFD != "yes" {
			t.Fatalf("expected host device %d to have iommufd enabled", i)
		}
		if dev.ACPI == nil {
			t.Fatalf("expected host device %d to have ACPI nodeset", i)
		}
		bdf := fmt.Sprintf("%s:%s:00.0",
			strings.TrimPrefix(dev.Source.Address.Domain, "0x"),
			strings.TrimPrefix(dev.Source.Address.Bus, "0x"),
		)
		expectedNodeSet, ok := expectedByBDF[bdf]
		if !ok {
			t.Fatalf("unexpected host device bdf %s", bdf)
		}
		if dev.ACPI.NodeSet != expectedNodeSet {
			t.Fatalf("expected host device %s to have ACPI nodeset %s, got %s", bdf, expectedNodeSet, dev.ACPI.NodeSet)
		}
		if _, exists := seenNodeSets[dev.ACPI.NodeSet]; exists {
			t.Fatalf("expected unique GI nodesets per GPU, found duplicate %s", dev.ACPI.NodeSet)
		}
		seenNodeSets[dev.ACPI.NodeSet] = struct{}{}
	}
	expectedNodeSetOrder := []string{"2-9", "10-17", "18-25", "26-33"}
	if got := mustHostDeviceNodeSets(t, domain); !reflect.DeepEqual(got, expectedNodeSetOrder) {
		t.Fatalf("expected host devices to be ordered by GI nodeset %v, got %v", expectedNodeSetOrder, got)
	}

	for id := 0; id <= 33; id++ {
		idStr := strconv.Itoa(id)
		found := false
		for _, cell := range domain.Spec.CPU.NUMA.Cells {
			if cell.ID == idStr {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected guest NUMA cell %d to exist after GI allocation", id)
		}
	}
}

func TestPrepareGraceGuestNUMATopologyKeepsStableGINodeSetsWhenAppliedBeforeNUMAHostDeviceTopology(t *testing.T) {
	defer restoreNUMAHelpers()
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		switch bdf {
		case "0000:08:00.0", "0000:09:00.0":
			return 0, nil
		case "0000:18:00.0", "0000:19:00.0":
			return 1, nil
		default:
			return -1, fmt.Errorf("unexpected bdf %s", bdf)
		}
	}
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "arm64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3":true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{Cells: []api.NUMACell{
					{ID: "0", CPUs: "0-59", Memory: 1, Unit: "GiB"},
					{ID: "1", CPUs: "60-119", Memory: 1, Unit: "GiB"},
				}},
			},
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x19"),
					newTestPCIHostDevice("gpu1", "0x0000", "0x08"),
					newTestPCIHostDevice("gpu2", "0x0000", "0x18"),
					newTestPCIHostDevice("gpu3", "0x0000", "0x09"),
				},
			},
		},
	}

	stubPCIPath("0000:08:00.0", []string{"0000:00:08.0", "0000:08:00.0"})
	stubPCIPath("0000:09:00.0", []string{"0000:00:09.0", "0000:09:00.0"})
	stubPCIPath("0000:18:00.0", []string{"0000:00:18.0", "0000:18:00.0"})
	stubPCIPath("0000:19:00.0", []string{"0000:00:19.0", "0000:19:00.0"})

	mustPrepareGraceGuestNUMATopology(t, vmi, domain)
	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	expectedNodeSetOrder := []string{"2-9", "10-17", "18-25", "26-33"}
	if got := mustHostDeviceNodeSets(t, domain); !reflect.DeepEqual(got, expectedNodeSetOrder) {
		t.Fatalf("expected host devices to keep stable GI nodeset order %v, got %v", expectedNodeSetOrder, got)
	}

	for id := 0; id <= 33; id++ {
		cell := mustNUMACellByID(t, domain, strconv.Itoa(id))
		if id >= 2 && id <= 33 && (cell.Memory != 0 || cell.Unit != "KiB") {
			t.Fatalf("expected prepared GI NUMA cell %d to stay zero-memory KiB, got memory=%d unit=%q", id, cell.Memory, cell.Unit)
		}
	}
}

func TestApplyNUMAHostDeviceTopologySkipsArm64GraceHostDeviceSettingsOnNonArm64(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: "amd64",
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	dev := domain.Spec.Devices.HostDevices[0]
	if dev.Driver != nil {
		t.Fatalf("expected host device driver not to be injected on non-arm64 architectures")
	}
	if dev.ACPI != nil {
		t.Fatalf("expected host device ACPI nodeset not to be injected on non-arm64 architectures")
	}
}

const defaultTopologyGroup = "default-topology"

var testPCIHierarchy map[string][]string

func setDefaultTopologyGrouping() {
	if testPCIHierarchy == nil {
		testPCIHierarchy = map[string][]string{}
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) {
		return defaultTopologyGroup, nil
	}
	getDevicePCIPathHierarchyFunc = func(bdf string) ([]string, error) {
		if path, ok := testPCIHierarchy[bdf]; ok {
			return path, nil
		}
		parts := strings.Split(bdf, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid bdf %q", bdf)
		}
		bus := parts[1]
		root := fmt.Sprintf("0000:00:%s.0", bus)
		return []string{root, bdf}, nil
	}
	getDeviceIOMMUGroupInfoFunc = func(string) (int, []string, error) {
		return -1, nil, nil
	}
}

func init() {
	setDefaultTopologyGrouping()
}

func restoreNUMAHelpers() {
	formatPCIAddressFunc = hardware.FormatPCIAddress
	getDeviceNumaNodeIntFunc = hardware.GetDeviceNumaNodeInt
	getMdevParentPCIAddressFunc = hardware.GetMdevParentPCIAddress
	getDevicePCIPathHierarchyFunc = hardware.GetDevicePCIPathHierarchy
	getDeviceIOMMUGroupInfoFunc = hardware.GetDeviceIOMMUGroupInfo
	getDevicePCITotalMMIOSizeFunc = getDevicePCITotalMMIOSize
	isIOMMUFDDeviceAvailableFunc = isIOMMUFDDeviceAvailable
	discoverEGMDevicesFunc = hardware.DiscoverEGMDevices
	statEGMDevicePathFunc = os.Stat
	readPCIDeviceFileFunc = os.ReadFile
	testPCIHierarchy = map[string][]string{}
	setDefaultTopologyGrouping()
}

func stubPCIPath(bdf string, path []string) {
	if testPCIHierarchy == nil {
		testPCIHierarchy = map[string][]string{}
	}
	testPCIHierarchy[bdf] = path
}

func newTestPCIHostDevice(name, domain, bus string) api.HostDevice {
	return api.HostDevice{
		Type: api.HostDevicePCI,
		Source: api.HostDeviceSource{
			Address: &api.Address{
				Type:     api.AddressPCI,
				Domain:   domain,
				Bus:      bus,
				Slot:     "0x00",
				Function: "0x0",
			},
		},
		Alias: api.NewUserDefinedAlias("hostdevice-" + name),
	}
}

func containsInt(list []int, value int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func assignNUMAMapping(domain *api.Domain, mapping map[int]int) {
	if domain.Spec.CPU.NUMA == nil {
		domain.Spec.CPU.NUMA = &api.NUMA{}
	}

	guestSet := make(map[int]struct{})
	hostByGuest := make(map[int][]int)
	for host, guest := range mapping {
		guestSet[guest] = struct{}{}
		hostByGuest[guest] = append(hostByGuest[guest], host)
	}

	guests := make([]int, 0, len(guestSet))
	for guest := range guestSet {
		guests = append(guests, guest)
	}
	sort.Ints(guests)

	cells := make([]api.NUMACell, 0, len(guests))
	for _, guest := range guests {
		cells = append(cells, api.NUMACell{ID: strconv.Itoa(guest)})
	}
	domain.Spec.CPU.NUMA.Cells = cells

	memNodes := make([]api.MemNode, 0, len(guests))
	for _, guest := range guests {
		hosts := hostByGuest[guest]
		sort.Ints(hosts)
		nodeSetParts := make([]string, len(hosts))
		for i, host := range hosts {
			nodeSetParts[i] = strconv.Itoa(host)
		}
		memNodes = append(memNodes, api.MemNode{
			CellID:  uint32(guest),
			Mode:    "strict",
			NodeSet: strings.Join(nodeSetParts, ","),
		})
	}
	domain.Spec.NUMATune = &api.NUMATune{
		MemNodes: memNodes,
	}
}

func appendGuestNUMACells(domain *api.Domain, ids ...int) {
	if domain.Spec.CPU.NUMA == nil {
		domain.Spec.CPU.NUMA = &api.NUMA{}
	}
	existing := make(map[string]struct{}, len(domain.Spec.CPU.NUMA.Cells))
	for _, cell := range domain.Spec.CPU.NUMA.Cells {
		existing[cell.ID] = struct{}{}
	}
	for _, id := range ids {
		idStr := strconv.Itoa(id)
		if _, ok := existing[idStr]; ok {
			continue
		}
		domain.Spec.CPU.NUMA.Cells = append(domain.Spec.CPU.NUMA.Cells, api.NUMACell{ID: idStr})
	}
	sort.Slice(domain.Spec.CPU.NUMA.Cells, func(i, j int) bool {
		idi, _ := strconv.Atoi(domain.Spec.CPU.NUMA.Cells[i].ID)
		idj, _ := strconv.Atoi(domain.Spec.CPU.NUMA.Cells[j].ID)
		return idi < idj
	})
}

// Error Handling Tests

func TestApplyNUMAHostDeviceTopologyDeviceResolutionFailure(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "", fmt.Errorf("device resolution failed")
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		t.Fatal("GetDeviceNumaNodeInt should not be called when device resolution fails")
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers when device resolution fails
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged when device resolution fails, got %d", len(domain.Spec.Devices.Controllers))
	}
	if domain.Spec.Devices.HostDevices[0].Address != nil {
		t.Fatalf("expected host device address to remain unset when device resolution fails")
	}
}

func TestApplyNUMAHostDeviceTopologyNumaDetectionFailure(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return -1, fmt.Errorf("NUMA detection failed")
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers when NUMA detection fails
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged when NUMA detection fails, got %d", len(domain.Spec.Devices.Controllers))
	}
	if domain.Spec.Devices.HostDevices[0].Address != nil {
		t.Fatalf("expected host device address to remain unset when NUMA detection fails")
	}
}

func TestApplyNUMAHostDeviceTopologyMdevParentResolutionFailure(t *testing.T) {
	defer restoreNUMAHelpers()

	getMdevParentPCIAddressFunc = func(uuid string) (string, error) {
		return "", fmt.Errorf("mdev parent resolution failed")
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		t.Fatal("GetDeviceNumaNodeInt should not be called when mdev parent resolution fails")
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDeviceMDev,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								UUID: "mdev-uuid",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers when mdev parent resolution fails
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged when mdev parent resolution fails, got %d", len(domain.Spec.Devices.Controllers))
	}
	if domain.Spec.Devices.HostDevices[0].Address != nil {
		t.Fatalf("expected host device address to remain unset when mdev parent resolution fails")
	}
}

// Edge Case Tests

func TestApplyNUMAHostDeviceTopologyNoHostDevices(t *testing.T) {
	defer restoreNUMAHelpers()

	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		t.Fatal("GetDeviceNumaNodeInt should not be called when no host devices")
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers when no host devices
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged when no host devices, got %d", len(domain.Spec.Devices.Controllers))
	}
}

func TestApplyNUMAHostDeviceTopologyNoNumaAffinity(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return -1, nil // No NUMA affinity
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers when devices have no NUMA affinity
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged when devices have no NUMA affinity, got %d", len(domain.Spec.Devices.Controllers))
	}
	if domain.Spec.Devices.HostDevices[0].Address != nil {
		t.Fatalf("expected host device address to remain unset when devices have no NUMA affinity")
	}
}

func TestApplyNUMAHostDeviceTopologyUnsupportedDeviceTypes(t *testing.T) {
	defer restoreNUMAHelpers()

	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		t.Fatal("GetDeviceNumaNodeInt should not be called for unsupported device types")
		return -1, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDeviceUSB, // Unsupported device type
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:   api.AddressPCI,
								Bus:    "1",
								Device: "2",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create any new controllers for unsupported device types
	if len(domain.Spec.Devices.Controllers) != 1 {
		t.Fatalf("expected controllers unchanged for unsupported device types, got %d", len(domain.Spec.Devices.Controllers))
	}
}

// Slot Allocation Tests

func TestApplyNUMAHostDeviceTopologySlotExhaustion(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	// Create domain with all root bus slots occupied (0x0a to 0x1f)
	controllers := []api.Controller{
		{Type: "pci", Index: "0", Model: "pcie-root"},
	}

	// Fill all available slots
	for slot := 0x0a; slot <= 0x1f; slot++ {
		controllers = append(controllers, api.Controller{
			Type:  "pci",
			Index: fmt.Sprintf("%d", slot),
			Model: "pcie-expander-bus",
			Address: &api.Address{
				Type:     api.AddressPCI,
				Domain:   "0x0000",
				Bus:      "0x00",
				Slot:     fmt.Sprintf("0x%02x", slot),
				Function: "0x0",
			},
		})
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: controllers,
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should not create new controllers when slots are exhausted
	// Original controllers + all occupied slots = expected count
	expectedControllers := len(controllers)
	if len(domain.Spec.Devices.Controllers) != expectedControllers {
		t.Fatalf("expected controllers unchanged when slots exhausted, got %d, expected %d",
			len(domain.Spec.Devices.Controllers), expectedControllers)
	}
}

func TestApplyNUMAHostDeviceTopologyReservesImplicitRootBusSlots(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{
						Type:  "pci",
						Index: "1",
						Model: "pcie-expander-bus",
						Address: &api.Address{
							Type:     api.AddressPCI,
							Domain:   "0x0000",
							Slot:     "0x0a",
							Function: "0x0",
						},
					},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	var reservedSlot string
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-expander-bus" || ctrl.Alias == nil {
			continue
		}
		if !strings.HasPrefix(ctrl.Alias.GetName(), numaPXBAliasPrefix) {
			continue
		}
		if ctrl.Address == nil {
			t.Fatalf("expected NUMA PXB controller to have an address")
		}
		reservedSlot = ctrl.Address.Slot
	}

	if reservedSlot == "" {
		t.Fatalf("expected NUMA PXB controller to be created")
	}

	if reservedSlot == "0x0a" {
		t.Fatalf("expected NUMA PXB controller to use a different slot than existing root bus devices")
	}
}

func TestApplyNUMAHostDeviceTopologyHandlesExistingHotplugSlots(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	controllers := []api.Controller{
		{Type: "pci", Index: "0", Model: "pcie-root"},
	}
	for i := 0; i < 3; i++ {
		controllers = append(controllers, api.Controller{
			Type:  "pci",
			Index: fmt.Sprintf("%d", i+1),
			Model: "pcie-root-port",
			Alias: api.NewUserDefinedAlias(fmt.Sprintf("%s%d", HotplugRootPortAliasPrefix, i)),
			Address: &api.Address{
				Type:     api.AddressPCI,
				Domain:   "0x0000",
				Bus:      "0x00",
				Slot:     fmt.Sprintf("0x%02x", rootHotplugSlotStart+i),
				Function: "0x0",
			},
		})
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: controllers,
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0})
	stubPCIPath("0000:01:00.0", []string{"0000:00:01.0", "0000:01:00.0"})

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	foundPXB := false
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model != "pcie-expander-bus" || ctrl.Alias == nil {
			continue
		}
		if !strings.HasPrefix(ctrl.Alias.GetName(), numaPXBAliasPrefix) {
			continue
		}
		if ctrl.Address == nil {
			t.Fatalf("expected NUMA PXB controller to have an address")
		}
		slot, err := parsePCISlot(ctrl.Address.Slot)
		if err != nil {
			t.Fatalf("failed to parse slot %s: %v", ctrl.Address.Slot, err)
		}
		if slot >= rootHotplugSlotStart {
			t.Fatalf("expected NUMA PXB controller slot to remain below hotplug range, got %s", ctrl.Address.Slot)
		}
		foundPXB = true
	}

	if !foundPXB {
		t.Fatalf("expected NUMA planner to allocate a PXB controller")
	}
}

func TestApplyNUMAHostDeviceTopologyExistingControllerSlots(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		return "0000:01:00.0", nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	// Create domain with existing controller using slot 0x0a
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{
						Type:  "pci",
						Index: "1",
						Model: "pcie-expander-bus",
						Address: &api.Address{
							Type:     api.AddressPCI,
							Domain:   "0x0000",
							Bus:      "0x00",
							Slot:     "0x0a", // Using default slot
							Function: "0x0",
						},
					},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Existing PXB should remain; planner may reuse it or allocate the next available slot depending on collapse behaviour
	hasSlot0A := false
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" && ctrl.Address != nil {
			switch ctrl.Address.Slot {
			case "0x0a":
				hasSlot0A = true
			}
		}
	}
	if !hasSlot0A {
		t.Fatalf("expected original expander bus at slot 0x0a to remain, controllers=%#v", domain.Spec.Devices.Controllers)
	}
	// Presence of 0x0b indicates planner allocated a fresh PXB; when collapsing to a single NUMA node we accept either behaviour.

	// PCI placement will assign downstream addresses after NUMA planning, so we only ensure no resources were lost.
}

func TestApplyNUMAHostDeviceTopologyMultipleDevicesPerNode(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		return fmt.Sprintf("0000:%s:%s.0", bus, slot), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		// Simulate 7 devices on NUMA node 0, 7 devices on NUMA node 1
		// NUMA node 0: 0000:03, 0000:04, 0000:05, 0000:06, 0000:07, 0000:08, 0000:41
		// NUMA node 1: 0000:83, 0000:84, 0000:85, 0000:86, 0000:87, 0000:88, 0000:89
		switch {
		case strings.HasPrefix(bdf, "0000:03:") || strings.HasPrefix(bdf, "0000:04:") ||
			strings.HasPrefix(bdf, "0000:05:") || strings.HasPrefix(bdf, "0000:06:") ||
			strings.HasPrefix(bdf, "0000:07:") || strings.HasPrefix(bdf, "0000:08:") ||
			strings.HasPrefix(bdf, "0000:41:"):
			return 0, nil
		case strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") ||
			strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") ||
			strings.HasPrefix(bdf, "0000:87:") || strings.HasPrefix(bdf, "0000:88:") ||
			strings.HasPrefix(bdf, "0000:89:"):
			return 1, nil
		default:
			return -1, nil
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	// Create 7 devices for NUMA node 0 and 7 for NUMA node 1
	hostDevices := []api.HostDevice{}

	// NUMA node 0 devices (0000:03-06, 0000:07-08, 0000:41)
	for _, bus := range []string{"0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x41"} {
		hostDevices = append(hostDevices, api.HostDevice{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      bus,
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		})
	}

	// NUMA node 1 devices (0000:83-86, 0000:87-88, 0000:89)
	for _, bus := range []string{"0x83", "0x84", "0x85", "0x86", "0x87", "0x88", "0x89"} {
		hostDevices = append(hostDevices, api.HostDevice{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      bus,
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		})
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: hostDevices,
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	for _, dev := range domain.Spec.Devices.HostDevices {
		addr := dev.Source.Address
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		bdf := fmt.Sprintf("0000:%s:%s.%s", bus, slot, function)
		stubPCIPath(bdf, []string{fmt.Sprintf("0000:00:%s.0", bus), bdf})
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Should create 2 PXB controllers (one per NUMA node)
	var pxbCount int
	var numaNodes []int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
			if ctrl.Target != nil && ctrl.Target.Node != nil {
				numaNodes = append(numaNodes, *ctrl.Target.Node)
			}
		}
	}
	if pxbCount != 2 {
		t.Fatalf("expected 2 PXB controllers, got %d", pxbCount)
	}
	if !(containsInt(numaNodes, 0) && containsInt(numaNodes, 1)) {
		t.Fatalf("expected PXB controllers for NUMA nodes 0 and 1, got %v", numaNodes)
	}

	// Should create 7 root ports per NUMA node (14 total)
	var rootPortCount int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-root-port" && ctrl.Alias != nil &&
			strings.HasPrefix(ctrl.Alias.GetName(), numaRootPortPrefix) {
			rootPortCount++
		}
	}
	if rootPortCount != 14 {
		t.Fatalf("expected 14 root ports (7 per NUMA node), got %d", rootPortCount)
	}

	// All devices should have addresses assigned
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("expected host device %d to have an address assigned", i)
		}
	}
}

// Controller Management Tests

func TestNewNUMAPCIPlannerControllerIndexCalculation(t *testing.T) {
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{Type: "pci", Index: "5", Model: "pcie-expander-bus"},
					{Type: "pci", Index: "10", Model: "pcie-root-port"},
				},
			},
		},
	}

	planner := newNUMAPCIPlanner(domain)

	// Should start from maxIndex + 1 (10 + 1 = 11)
	if planner.nextControllerIndex != 11 {
		t.Fatalf("expected nextControllerIndex to be 11, got %d", planner.nextControllerIndex)
	}
}

func TestNewNUMAPCIPlannerExistingControllerDetection(t *testing.T) {
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{
						Type:  "pci",
						Index: "1",
						Model: "pcie-expander-bus",
						Address: &api.Address{
							Type:     api.AddressPCI,
							Domain:   "0x0000",
							Bus:      "0x00",
							Slot:     "0x0a",
							Function: "0x0",
						},
					},
					{
						Type:  "pci",
						Index: "2",
						Model: "pcie-expander-bus",
						Address: &api.Address{
							Type:     api.AddressPCI,
							Domain:   "0x0000",
							Bus:      "0x00",
							Slot:     "0x0b",
							Function: "0x0",
						},
					},
				},
			},
		},
	}

	planner := newNUMAPCIPlanner(domain)

	// Should detect used slots 0x0a and 0x0b
	slots, ok := planner.usedBusSlots[busSlotKey{controller: rootBusControllerMarker, bus: 0}]
	if !ok {
		t.Fatalf("expected root bus slot tracking to be initialized")
	}
	if _, used := slots[0x0a]; !used {
		t.Fatalf("expected slot 0x0a to be marked as used")
	}
	if _, used := slots[0x0b]; !used {
		t.Fatalf("expected slot 0x0b to be marked as used")
	}

	// Should set nextPXBSlot to 0x0c (0x0b + 1)
	if planner.nextPXBSlot != 0x0c {
		t.Fatalf("expected nextPXBSlot to be 0x0c, got 0x%02x", planner.nextPXBSlot)
	}
}

// Address Assignment Tests

func TestAssignHostDeviceToRootPort(t *testing.T) {
	dev := &api.HostDevice{Type: api.HostDevicePCI}

	port := &rootPortInfo{controllerIndex: 5, downstreamBus: 0x90}

	assignHostDeviceToRootPort(dev, port)

	if dev.Address == nil {
		t.Fatalf("expected device address to be set")
	}
	if dev.Address.Type != api.AddressPCI {
		t.Fatalf("expected address type to be PCI")
	}
	if dev.Address.Domain != "0x0000" {
		t.Fatalf("expected domain to be 0x0000, got %s", dev.Address.Domain)
	}
	// After fix to match NVIDIA documentation: devices use Bus attribute (not Controller)
	if dev.Address.Bus != strconv.Itoa(port.controllerIndex) {
		t.Fatalf("expected bus to be %d (NVIDIA docs pattern), got %s", port.controllerIndex, dev.Address.Bus)
	}
	if dev.Address.Controller != "" {
		t.Fatalf("expected controller to be empty (using Bus instead), got %s", dev.Address.Controller)
	}
	if dev.Address.Slot != "0x00" {
		t.Fatalf("expected slot to be 0x00 for root port downstream placement, got %s", dev.Address.Slot)
	}
	if dev.Address.Function != "0x0" {
		t.Fatalf("expected function to be 0x0 for root port downstream placement, got %s", dev.Address.Function)
	}
}

func TestHostDeviceAddressFormat(t *testing.T) {
	dev := &api.HostDevice{Type: api.HostDevicePCI}
	port := &rootPortInfo{controllerIndex: 10, downstreamBus: 0xa1}

	assignHostDeviceToRootPort(dev, port)

	if dev.Address == nil {
		t.Fatalf("expected device address to be set")
	}
	// After fix to match NVIDIA documentation: devices use Bus attribute (not Controller)
	if dev.Address.Bus != strconv.Itoa(port.controllerIndex) {
		t.Fatalf("expected bus to be %d (NVIDIA docs pattern), got %s", port.controllerIndex, dev.Address.Bus)
	}
	if dev.Address.Controller != "" {
		t.Fatalf("expected controller to be empty (using Bus instead), got %s", dev.Address.Controller)
	}
	if dev.Address.Slot != "0x00" {
		t.Fatalf("expected slot to be 0x00 for root port downstream placement, got %s", dev.Address.Slot)
	}
	if dev.Address.Function != "0x0" {
		t.Fatalf("expected function to be 0x0 for root port downstream placement, got %s", dev.Address.Function)
	}
}

// Real-world scenario test based on provided device examples

func TestApplyNUMAHostDeviceTopologyRealWorldScenario(t *testing.T) {
	defer restoreNUMAHelpers()

	// Mock the hardware functions to return the exact NUMA nodes from the provided examples
	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		// Map the exact BDFs from the provided examples to their NUMA nodes
		switch bdf {
		// NUMA Node 0 devices
		case "0000:03:00.0", "0000:04:00.0", "0000:05:00.0", "0000:06:00.0": // NVIDIA GPUs
			return 0, nil
		case "0000:07:00.0", "0000:08:00.0": // Mellanox IB devices
			return 0, nil
		case "0000:41:00.0": // Ethernet device
			return 0, nil
		// NUMA Node 1 devices
		case "0000:83:00.0", "0000:84:00.0", "0000:85:00.0", "0000:86:00.0": // NVIDIA GPUs
			return 1, nil
		case "0000:87:00.0", "0000:88:00.0": // Mellanox IB devices
			return 1, nil
		case "0000:89:00.0": // Additional device
			return 1, nil
		default:
			return -1, fmt.Errorf("unknown device BDF: %s", bdf)
		}
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	// Create host devices matching the exact BDFs from the provided examples, provided by device plugin or DRA driver
	hostDevices := []api.HostDevice{
		// NUMA Node 0 - NVIDIA GPUs
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x03",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x04",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x05",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x06",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		// NUMA Node 0 - Mellanox IB devices
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x07",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x08",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		// NUMA Node 0 - BlueField device
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x41",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		// NUMA Node 1 - NVIDIA GPUs
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x83",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x84",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x85",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x86",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		// NUMA Node 1 - Mellanox IB devices
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x87",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x88",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
		// NUMA Node 1 - Additional device
		{
			Type: api.HostDevicePCI,
			Source: api.HostDeviceSource{
				Address: &api.Address{
					Type:     api.AddressPCI,
					Domain:   "0x0000",
					Bus:      "0x89",
					Slot:     "0x00",
					Function: "0x0",
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
				},
				HostDevices: hostDevices,
			},
		},
	}

	assignNUMAMapping(domain, map[int]int{0: 0, 1: 1})
	for _, dev := range domain.Spec.Devices.HostDevices {
		addr := dev.Source.Address
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		bdf := fmt.Sprintf("0000:%s:%s.%s", bus, slot, function)
		stubPCIPath(bdf, []string{fmt.Sprintf("0000:80:%s.0", bus), bdf})
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Verify PXB controllers are created for both NUMA nodes
	var pxbCount int
	var numaNodes []int
	var pxbSlots []string
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
			if ctrl.Target != nil && ctrl.Target.Node != nil {
				numaNodes = append(numaNodes, *ctrl.Target.Node)
			}
			if ctrl.Address != nil {
				pxbSlots = append(pxbSlots, ctrl.Address.Slot)
			}
		}
	}

	if pxbCount != 2 {
		t.Fatalf("expected 2 PXB controllers (one per NUMA node), got %d", pxbCount)
	}
	if !(containsInt(numaNodes, 0) && containsInt(numaNodes, 1)) {
		t.Fatalf("expected PXB controllers for NUMA nodes 0 and 1, got %v", numaNodes)
	}

	// Verify root ports are created (7 devices per NUMA node = 14 total root ports)
	var rootPortCount int
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-root-port" && ctrl.Alias != nil &&
			strings.HasPrefix(ctrl.Alias.GetName(), numaRootPortPrefix) {
			rootPortCount++
		}
	}

	if rootPortCount != len(hostDevices) {
		t.Fatalf("expected %d root ports, got %d", len(hostDevices), rootPortCount)
	}

	// Verify all host devices have addresses assigned
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("expected host device %d to have an address assigned", i)
		}
		// After fix to match NVIDIA documentation: devices use Bus attribute (not Controller)
		if dev.Address.Bus == "" {
			t.Fatalf("expected host device %d to have bus assigned (NVIDIA docs pattern)", i)
		}
		if dev.Address.Controller != "" {
			t.Fatalf("expected host device %d to leave controller empty (using Bus instead), got %s", i, dev.Address.Controller)
		}
	}
}

// Conflict Detection Tests
func TestNUMAPCIPlannerConflictDetection(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		return fmt.Sprintf("0000:%s:%s.0", bus, slot), nil
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		return 0, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	// Create domain with existing controllers that use slots and chassis
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{
					{Type: "pci", Index: "0", Model: "pcie-root"},
					{
						Type:  "pci",
						Index: "1",
						Model: "pcie-expander-bus",
						Address: &api.Address{
							Type:     api.AddressPCI,
							Domain:   "0x0000",
							Bus:      "0x00",
							Slot:     "0x0a", // Using default slot
							Function: "0x0",
						},
					},
					{
						Type:  "pci",
						Index: "2",
						Model: "pcie-root-port",
						Target: &api.ControllerTarget{
							Chassis: "5", // Using chassis 5
						},
					},
				},
				HostDevices: []api.HostDevice{
					{
						Type: api.HostDevicePCI,
						Source: api.HostDeviceSource{
							Address: &api.Address{
								Type:     api.AddressPCI,
								Domain:   "0x0000",
								Bus:      "0x01",
								Slot:     "0x00",
								Function: "0x0",
							},
						},
					},
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Verify no slot conflicts
	var usedSlots = make(map[string]bool)
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Address != nil && ctrl.Address.Slot != "" {
			slotKey := fmt.Sprintf("%s:%s", ctrl.Address.Bus, ctrl.Address.Slot)
			if usedSlots[slotKey] {
				t.Fatalf("slot conflict detected: %s", slotKey)
			}
			usedSlots[slotKey] = true
		}
	}

	// Verify no chassis conflicts
	var usedChassis = make(map[string]bool)
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Target != nil && ctrl.Target.Chassis != "" {
			if usedChassis[ctrl.Target.Chassis] {
				t.Fatalf("chassis conflict detected: %s", ctrl.Target.Chassis)
			}
			usedChassis[ctrl.Target.Chassis] = true
		}
	}

	// Verify no controller index conflicts
	var usedIndices = make(map[string]bool)
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if usedIndices[ctrl.Index] {
			t.Fatalf("controller index conflict detected: %s", ctrl.Index)
		}
		usedIndices[ctrl.Index] = true
	}

	t.Logf("No slot conflicts detected")
	t.Logf("No chassis conflicts detected")
	t.Logf("No controller index conflicts detected")
}

func TestNUMAPCIPlannerBusNumberAllocation(t *testing.T) {
	t.Parallel()

	// Create a domain with no controllers - planner will add everything
	domain := &api.Domain{
		Spec: api.DomainSpec{
			Devices: api.Devices{
				Controllers: []api.Controller{},
			},
		},
	}

	planner := newNUMAPCIPlanner(domain)

	// Simulate what ApplyNUMAHostDeviceTopology does:
	// 1. Reserve PXB bus numbers for NUMA 0 and 1
	groupKeys := []deviceGroupKey{
		{hostNUMANode: 0, guestNUMANode: 0, pathKey: "path0"},
		{hostNUMANode: 1, guestNUMANode: 1, pathKey: "path1"},
	}
	planner.reservePXBBusNumbers(groupKeys)

	// 2. Add default root port
	planner.addDefaultRootPort()

	// 3. Create PXBs
	pxb0, err := planner.ensurePXB(0, 0)
	if err != nil {
		t.Fatalf("Failed to create PXB for NUMA 0: %v", err)
	}
	_, err = planner.ensurePXB(1, 1)
	if err != nil {
		t.Fatalf("Failed to create PXB for NUMA 1: %v", err)
	}

	// 4. Add some root ports to consume bus numbers
	for i := 0; i < 3; i++ {
		alias := fmt.Sprintf("rp-numa0-%d", i)
		_, err := planner.addRootPort(pxb0, alias)
		if err != nil {
			t.Fatalf("Failed to add root port %s: %v", alias, err)
		}
	}

	// Collect all bus numbers used
	busNumbers := make(map[int]string) // bus number -> controller description
	var defaultRootPortBus int = -1
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Target != nil && ctrl.Target.BusNr != "" {
			busNr, err := strconv.Atoi(ctrl.Target.BusNr)
			if err == nil {
				desc := fmt.Sprintf("Controller index=%s model=%s alias=%s",
					ctrl.Index, ctrl.Model, ctrl.Alias.GetName())
				if existing, exists := busNumbers[busNr]; exists {
					t.Fatalf("Bus number %d (0x%02x) used by multiple controllers:\n  %s\n  %s",
						busNr, busNr, existing, desc)
				}
				busNumbers[busNr] = desc

				// Track default root port bus
				if ctrl.Alias != nil && ctrl.Alias.GetName() == "default-root-port" {
					defaultRootPortBus = busNr
				}
			}
		}
	}

	// Verify PXB buses are at expected locations
	pxbNuma0Bus := pxbBusNumberBase + (0 * pxbBusNumberSpacing) // 0x20 = 32
	pxbNuma1Bus := pxbBusNumberBase + (1 * pxbBusNumberSpacing) // 0x40 = 64

	if _, exists := busNumbers[pxbNuma0Bus]; !exists {
		t.Errorf("Expected PXB bus 0x%02x (%d) for NUMA 0 not found", pxbNuma0Bus, pxbNuma0Bus)
	}
	if _, exists := busNumbers[pxbNuma1Bus]; !exists {
		t.Errorf("Expected PXB bus 0x%02x (%d) for NUMA 1 not found", pxbNuma1Bus, pxbNuma1Bus)
	}

	// Verify default root port does NOT have a busNr in its target
	// The default root port lets libvirt auto-assign the downstream bus number
	// because specifying busNr is not supported (or required) for simple root ports
	if defaultRootPortBus >= 0 {
		t.Errorf("Default root port should not have busNr in target, but got 0x%02x (%d)",
			defaultRootPortBus, defaultRootPortBus)
	}

	// Verify PXB buses use their pre-calculated values (not affected by sequential allocation)
	if _, exists := busNumbers[pxbNuma0Bus]; !exists {
		t.Errorf("PXB NUMA 0 bus 0x%02x was not allocated correctly", pxbNuma0Bus)
	}
	if _, exists := busNumbers[pxbNuma1Bus]; !exists {
		t.Errorf("PXB NUMA 1 bus 0x%02x was not allocated correctly", pxbNuma1Bus)
	}

	t.Logf("✓ Bus allocation verified correctly:")
	t.Logf("  Default root port: no busNr (libvirt auto-assigns)")
	t.Logf("  PXB NUMA 0: 0x%02x (%d)", pxbNuma0Bus, pxbNuma0Bus)
	t.Logf("  PXB NUMA 1: 0x%02x (%d)", pxbNuma1Bus, pxbNuma1Bus)
	t.Logf("  Total controllers: %d", len(domain.Spec.Devices.Controllers))
}

// TestHGXH200TopologyWithPCIeSwitches tests the HGX H200 topology pattern:
// - 8 GPUs (4 per NUMA node) behind PCIe switches
// - 4 IB NICs (2 per NUMA node) behind PCIe switches
// - Devices sharing same physical switch get virtual switch in guest
func TestHGXH200TopologyWithPCIeSwitches(t *testing.T) {
	defer restoreNUMAHelpers()

	// Mock HGX H200 physical topology:
	// NUMA 0: 4 GPUs sharing switch at 0000:17:00.0, 2 IB NICs sharing switch at 0000:18:00.0
	// NUMA 1: 4 GPUs sharing switch at 0000:85:00.0, 2 IB NICs sharing switch at 0000:86:00.0

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		// GPUs and NICs on NUMA 0
		if strings.HasPrefix(bdf, "0000:06:") || strings.HasPrefix(bdf, "0000:07:") ||
			strings.HasPrefix(bdf, "0000:08:") || strings.HasPrefix(bdf, "0000:09:") {
			return 0, nil // 4 GPUs on NUMA 0
		}
		if strings.HasPrefix(bdf, "0000:0a:") || strings.HasPrefix(bdf, "0000:0b:") {
			return 0, nil // 2 IB NICs on NUMA 0
		}
		// GPUs and NICs on NUMA 1
		if strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") ||
			strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") {
			return 1, nil // 4 GPUs on NUMA 1
		}
		if strings.HasPrefix(bdf, "0000:87:") || strings.HasPrefix(bdf, "0000:88:") {
			return 1, nil // 2 IB NICs on NUMA 1
		}
		return -1, fmt.Errorf("unknown device")
	}

	getDevicePCIProximityGroupFunc = func(bdf string) (string, error) {
		// Group devices by their physical switch
		numaNode, _ := getDeviceNumaNodeIntFunc(bdf)
		if strings.HasPrefix(bdf, "0000:06:") || strings.HasPrefix(bdf, "0000:07:") ||
			strings.HasPrefix(bdf, "0000:08:") || strings.HasPrefix(bdf, "0000:09:") {
			return fmt.Sprintf("numa-%d-gpu-switch", numaNode), nil
		}
		if strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") ||
			strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") {
			return fmt.Sprintf("numa-%d-gpu-switch", numaNode), nil
		}
		if strings.HasPrefix(bdf, "0000:0a:") || strings.HasPrefix(bdf, "0000:0b:") {
			return fmt.Sprintf("numa-%d-ib-switch", numaNode), nil
		}
		if strings.HasPrefix(bdf, "0000:87:") || strings.HasPrefix(bdf, "0000:88:") {
			return fmt.Sprintf("numa-%d-ib-switch", numaNode), nil
		}
		return "", fmt.Errorf("unknown device")
	}

	getDevicePCIPathHierarchyFunc = func(bdf string) ([]string, error) {
		// Simulate devices behind physical PCIe switches
		// Key insight: Devices sharing a physical switch share the same parent path
		// The switch upstream port is the common parent

		// GPUs on NUMA 0 all share switch at 0000:17:00.0
		if strings.HasPrefix(bdf, "0000:06:") || strings.HasPrefix(bdf, "0000:07:") ||
			strings.HasPrefix(bdf, "0000:08:") || strings.HasPrefix(bdf, "0000:09:") {
			return []string{"0000:00:00.0", "0000:17:00.0", bdf}, nil
		}
		// IB NICs on NUMA 0 all share switch at 0000:20:00.0
		if strings.HasPrefix(bdf, "0000:0a:") || strings.HasPrefix(bdf, "0000:0b:") {
			return []string{"0000:00:00.0", "0000:20:00.0", bdf}, nil
		}
		// GPUs on NUMA 1 all share switch at 0000:85:00.0
		if strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") ||
			strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") {
			return []string{"0000:00:00.0", "0000:85:00.0", bdf}, nil
		}
		// IB NICs on NUMA 1 all share switch at 0000:90:00.0
		if strings.HasPrefix(bdf, "0000:87:") || strings.HasPrefix(bdf, "0000:88:") {
			return []string{"0000:00:00.0", "0000:90:00.0", bdf}, nil
		}
		return nil, fmt.Errorf("unknown device")
	}

	getDeviceIOMMUGroupInfoFunc = func(bdf string) (int, []string, error) {
		// Each device in its own IOMMU group for simplicity
		return 0, []string{bdf}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0"},
						{ID: "1"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					// NUMA 0 GPUs (4 devices sharing switch path)
					newTestPCIHostDevice("gpu0", "0000", "06"),
					newTestPCIHostDevice("gpu1", "0000", "07"),
					newTestPCIHostDevice("gpu2", "0000", "08"),
					newTestPCIHostDevice("gpu3", "0000", "09"),
					// NUMA 0 IB NICs (2 devices sharing switch path)
					newTestPCIHostDevice("ib0", "0000", "0a"),
					newTestPCIHostDevice("ib1", "0000", "0b"),
					// NUMA 1 GPUs (4 devices sharing switch path)
					newTestPCIHostDevice("gpu4", "0000", "83"),
					newTestPCIHostDevice("gpu5", "0000", "84"),
					newTestPCIHostDevice("gpu6", "0000", "85"),
					newTestPCIHostDevice("gpu7", "0000", "86"),
					// NUMA 1 IB NICs (2 devices sharing switch path)
					newTestPCIHostDevice("ib2", "0000", "87"),
					newTestPCIHostDevice("ib3", "0000", "88"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Verify PXB controllers for both NUMA nodes
	pxbCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
		}
	}
	if pxbCount != 2 {
		t.Fatalf("expected 2 PXB controllers (one per NUMA node), got %d", pxbCount)
	}

	// Verify PCIe switches were created
	// Should have 4 switches total: 2 for GPU groups (NUMA 0 & 1), 2 for IB groups (NUMA 0 & 1)
	upstreamCount := 0
	downstreamCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-switch-upstream-port" {
			upstreamCount++
		}
		if ctrl.Model == "pcie-switch-downstream-port" {
			downstreamCount++
		}
	}

	expectedSwitches := 4 // 2 GPU switches + 2 IB switches
	if upstreamCount != expectedSwitches {
		t.Fatalf("expected %d upstream ports (one per switch), got %d", expectedSwitches, upstreamCount)
	}

	// Each GPU switch needs 4 downstream ports, each IB switch needs 2
	// But implementation creates 4 ports minimum (NVIDIA standard)
	expectedDownstream := 16 // 4 switches × 4 ports each
	if downstreamCount != expectedDownstream {
		t.Fatalf("expected %d downstream ports (%d switches × 4 ports), got %d", expectedDownstream, expectedSwitches, downstreamCount)
	}

	// Verify all devices got Bus attribute assigned (not Controller)
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("device %d missing address", i)
		}
		if dev.Address.Bus == "" {
			t.Fatalf("device %d missing Bus attribute (NVIDIA pattern)", i)
		}
		if dev.Address.Controller != "" {
			t.Fatalf("device %d has Controller attribute (should use Bus), got %s", i, dev.Address.Controller)
		}
	}

	t.Logf("✓ HGX H200 topology verified:")
	t.Logf("  PXB controllers: %d (NUMA 0 + NUMA 1)", pxbCount)
	t.Logf("  PCIe switches: %d (2 GPU switches + 2 IB switches)", upstreamCount)
	t.Logf("  Downstream ports: %d (%d per switch)", downstreamCount, downstreamCount/upstreamCount)
	t.Logf("  Total devices: %d (8 GPUs + 4 IB NICs)", len(domain.Spec.Devices.HostDevices))
}

// TestSimpleServerDirectTopology tests simple server topology:
// - 4 GPUs (2 per NUMA node) directly on root ports
// - No physical PCIe switches - each device has unique path
// - Should NOT create virtual switches
func TestSimpleServerDirectTopology(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		if strings.HasPrefix(bdf, "0000:03:") || strings.HasPrefix(bdf, "0000:04:") {
			return 0, nil // 2 GPUs on NUMA 0
		}
		if strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") {
			return 1, nil // 2 GPUs on NUMA 1
		}
		return -1, fmt.Errorf("unknown device")
	}

	getDevicePCIProximityGroupFunc = func(bdf string) (string, error) {
		numaNode, _ := getDeviceNumaNodeIntFunc(bdf)
		return fmt.Sprintf("numa-%d", numaNode), nil
	}

	getDevicePCIPathHierarchyFunc = func(bdf string) ([]string, error) {
		// Each device directly on root complex - unique paths (no shared switches)
		if strings.HasPrefix(bdf, "0000:03:") {
			return []string{"0000:00:00.0", "0000:01:00.0", bdf}, nil
		}
		if strings.HasPrefix(bdf, "0000:04:") {
			return []string{"0000:00:00.0", "0000:02:00.0", bdf}, nil
		}
		if strings.HasPrefix(bdf, "0000:83:") {
			return []string{"0000:00:00.0", "0000:81:00.0", bdf}, nil
		}
		if strings.HasPrefix(bdf, "0000:84:") {
			return []string{"0000:00:00.0", "0000:82:00.0", bdf}, nil
		}
		return nil, fmt.Errorf("unknown device")
	}

	getDeviceIOMMUGroupInfoFunc = func(bdf string) (int, []string, error) {
		return 0, []string{bdf}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0"},
						{ID: "1"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					// NUMA 0 GPUs (each with unique path - direct attachment)
					newTestPCIHostDevice("gpu0", "0000", "03"),
					newTestPCIHostDevice("gpu1", "0000", "04"),
					// NUMA 1 GPUs (each with unique path - direct attachment)
					newTestPCIHostDevice("gpu2", "0000", "83"),
					newTestPCIHostDevice("gpu3", "0000", "84"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Verify PXB controllers for both NUMA nodes
	pxbCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
		}
	}
	if pxbCount != 2 {
		t.Fatalf("expected 2 PXB controllers (one per NUMA node), got %d", pxbCount)
	}

	// Verify NO PCIe switches were created (direct topology)
	upstreamCount := 0
	downstreamCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-switch-upstream-port" {
			upstreamCount++
		}
		if ctrl.Model == "pcie-switch-downstream-port" {
			downstreamCount++
		}
	}

	if upstreamCount != 0 {
		t.Fatalf("expected 0 upstream ports (direct topology, no switches), got %d", upstreamCount)
	}
	if downstreamCount != 0 {
		t.Fatalf("expected 0 downstream ports (direct topology, no switches), got %d", downstreamCount)
	}

	// Verify root ports were created (one per device)
	rootPortCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-root-port" && ctrl.Alias != nil {
			aliasName := ctrl.Alias.GetName()
			if strings.HasPrefix(aliasName, numaRootPortPrefix) {
				rootPortCount++
			}
		}
	}
	if rootPortCount != len(domain.Spec.Devices.HostDevices) {
		t.Fatalf("expected %d root ports (one per device), got %d", len(domain.Spec.Devices.HostDevices), rootPortCount)
	}

	// Verify all devices got Bus attribute assigned directly to root ports
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("device %d missing address", i)
		}
		if dev.Address.Bus == "" {
			t.Fatalf("device %d missing Bus attribute (NVIDIA pattern)", i)
		}
		if dev.Address.Controller != "" {
			t.Fatalf("device %d has Controller attribute (should use Bus), got %s", i, dev.Address.Controller)
		}
	}

	t.Logf("✓ Simple server topology verified:")
	t.Logf("  PXB controllers: %d (NUMA 0 + NUMA 1)", pxbCount)
	t.Logf("  PCIe switches: 0 (direct topology, no switches needed)")
	t.Logf("  Root ports: %d (one per device, direct attachment)", rootPortCount)
	t.Logf("  Total devices: %d (4 GPUs)", len(domain.Spec.Devices.HostDevices))
}

// TestMixedTopology tests mixed environment:
// - Some devices behind physical switches (get virtual switches)
// - Some devices directly on root ports (no virtual switches)
// - Tests that planner handles both patterns correctly in same VM
func TestMixedTopology(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		// NUMA 0: 4 GPUs behind switch + 1 IB NIC direct
		if strings.HasPrefix(bdf, "0000:06:") || strings.HasPrefix(bdf, "0000:07:") ||
			strings.HasPrefix(bdf, "0000:08:") || strings.HasPrefix(bdf, "0000:09:") ||
			strings.HasPrefix(bdf, "0000:0a:") {
			return 0, nil
		}
		// NUMA 1: 2 GPUs direct + 2 IB NICs behind switch
		if strings.HasPrefix(bdf, "0000:83:") || strings.HasPrefix(bdf, "0000:84:") ||
			strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") {
			return 1, nil
		}
		return -1, fmt.Errorf("unknown device")
	}

	getDevicePCIProximityGroupFunc = func(bdf string) (string, error) {
		numaNode, _ := getDeviceNumaNodeIntFunc(bdf)
		return fmt.Sprintf("numa-%d", numaNode), nil
	}

	getDevicePCIPathHierarchyFunc = func(bdf string) ([]string, error) {
		// NUMA 0: 4 GPUs share switch at 0000:17:00.0
		if strings.HasPrefix(bdf, "0000:06:") || strings.HasPrefix(bdf, "0000:07:") ||
			strings.HasPrefix(bdf, "0000:08:") || strings.HasPrefix(bdf, "0000:09:") {
			return []string{"0000:00:00.0", "0000:17:00.0", bdf}, nil
		}
		// NUMA 0: 1 IB NIC direct (unique path)
		if strings.HasPrefix(bdf, "0000:0a:") {
			return []string{"0000:00:00.0", "0000:20:00.0", bdf}, nil
		}
		// NUMA 1: 2 GPUs direct (unique paths)
		if strings.HasPrefix(bdf, "0000:83:") {
			return []string{"0000:00:00.0", "0000:81:00.0", bdf}, nil
		}
		if strings.HasPrefix(bdf, "0000:84:") {
			return []string{"0000:00:00.0", "0000:82:00.0", bdf}, nil
		}
		// NUMA 1: 2 IB NICs share switch at 0000:90:00.0
		if strings.HasPrefix(bdf, "0000:85:") || strings.HasPrefix(bdf, "0000:86:") {
			return []string{"0000:00:00.0", "0000:90:00.0", bdf}, nil
		}
		return nil, fmt.Errorf("unknown device")
	}

	getDeviceIOMMUGroupInfoFunc = func(bdf string) (int, []string, error) {
		return 0, []string{bdf}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0"},
						{ID: "1"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					// NUMA 0: 4 GPUs behind switch
					newTestPCIHostDevice("gpu0", "0000", "06"),
					newTestPCIHostDevice("gpu1", "0000", "07"),
					newTestPCIHostDevice("gpu2", "0000", "08"),
					newTestPCIHostDevice("gpu3", "0000", "09"),
					// NUMA 0: 1 IB NIC direct
					newTestPCIHostDevice("ib0", "0000", "0a"),
					// NUMA 1: 2 GPUs direct
					newTestPCIHostDevice("gpu4", "0000", "83"),
					newTestPCIHostDevice("gpu5", "0000", "84"),
					// NUMA 1: 2 IB NICs behind switch
					newTestPCIHostDevice("ib1", "0000", "85"),
					newTestPCIHostDevice("ib2", "0000", "86"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	// Verify PXB controllers for both NUMA nodes
	pxbCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-expander-bus" {
			pxbCount++
		}
	}
	if pxbCount != 2 {
		t.Fatalf("expected 2 PXB controllers (one per NUMA node), got %d", pxbCount)
	}

	// Verify PCIe switches: Should have 2 switches (NUMA 0 GPU switch + NUMA 1 IB switch)
	upstreamCount := 0
	downstreamCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-switch-upstream-port" {
			upstreamCount++
		}
		if ctrl.Model == "pcie-switch-downstream-port" {
			downstreamCount++
		}
	}

	expectedSwitches := 2 // NUMA 0 GPU switch (4 devices) + NUMA 1 IB switch (2 devices)
	if upstreamCount != expectedSwitches {
		t.Fatalf("expected %d upstream ports (mixed topology), got %d", expectedSwitches, upstreamCount)
	}

	// Each switch gets 4 downstream ports (NVIDIA standard)
	expectedDownstream := 8 // 2 switches × 4 ports
	if downstreamCount != expectedDownstream {
		t.Fatalf("expected %d downstream ports (%d switches × 4 ports), got %d", expectedDownstream, expectedSwitches, downstreamCount)
	}

	// Verify root ports: 2 for switches + 3 for direct devices = 5 total
	rootPortCount := 0
	for _, ctrl := range domain.Spec.Devices.Controllers {
		if ctrl.Model == "pcie-root-port" && ctrl.Alias != nil {
			aliasName := ctrl.Alias.GetName()
			if strings.HasPrefix(aliasName, numaRootPortPrefix) {
				rootPortCount++
			}
		}
	}
	expectedRootPorts := 5 // 2 for switches + 3 for direct devices
	if rootPortCount != expectedRootPorts {
		t.Fatalf("expected %d root ports (2 for switches + 3 direct), got %d", expectedRootPorts, rootPortCount)
	}

	// Verify all devices got Bus attribute assigned
	for i, dev := range domain.Spec.Devices.HostDevices {
		if dev.Address == nil {
			t.Fatalf("device %d missing address", i)
		}
		if dev.Address.Bus == "" {
			t.Fatalf("device %d missing Bus attribute (NVIDIA pattern)", i)
		}
		if dev.Address.Controller != "" {
			t.Fatalf("device %d has Controller attribute (should use Bus), got %s", i, dev.Address.Controller)
		}
	}

	t.Logf("✓ Mixed topology verified:")
	t.Logf("  PXB controllers: %d (NUMA 0 + NUMA 1)", pxbCount)
	t.Logf("  PCIe switches: %d (1 GPU switch on NUMA 0 + 1 IB switch on NUMA 1)", upstreamCount)
	t.Logf("  Downstream ports: %d (%d per switch)", downstreamCount, downstreamCount/upstreamCount)
	t.Logf("  Root ports: %d (2 for switches + 3 for direct devices)", rootPortCount)
	t.Logf("  Total devices: %d (6 GPUs + 3 NICs)", len(domain.Spec.Devices.HostDevices))
	t.Logf("  Distribution: NUMA 0 has 4 GPUs (switch) + 1 NIC (direct), NUMA 1 has 2 GPUs (direct) + 2 NICs (switch)")
}

func TestComputePCISwitchGroupKey(t *testing.T) {
	testCases := []struct {
		name     string
		path     []string
		bdf      string
		expected string
	}{
		{
			name:     "GPU0 behind switch on NUMA 0",
			path:     []string{"0000:00:01.1", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0"},
			bdf:      "0000:03:00.0",
			expected: "0000:00:01.1/0000:01:00.0",
		},
		{
			name:     "GPU1 behind same switch on NUMA 0",
			path:     []string{"0000:00:01.1", "0000:01:00.0", "0000:02:01.0", "0000:04:00.0"},
			bdf:      "0000:04:00.0",
			expected: "0000:00:01.1/0000:01:00.0",
		},
		{
			name:     "GPU2 behind same switch on NUMA 0",
			path:     []string{"0000:00:01.1", "0000:01:00.0", "0000:02:02.0", "0000:05:00.0"},
			bdf:      "0000:05:00.0",
			expected: "0000:00:01.1/0000:01:00.0",
		},
		{
			name:     "NIC0 behind same switch on NUMA 0",
			path:     []string{"0000:00:01.1", "0000:01:00.0", "0000:02:04.0", "0000:07:00.0"},
			bdf:      "0000:07:00.0",
			expected: "0000:00:01.1/0000:01:00.0",
		},
		{
			name:     "GPU on NUMA 1 behind different switch",
			path:     []string{"0000:80:01.1", "0000:81:00.0", "0000:82:00.0", "0000:83:00.0"},
			bdf:      "0000:83:00.0",
			expected: "0000:80:01.1/0000:81:00.0",
		},
		{
			name:     "Direct-attached NIC (no switch)",
			path:     []string{"0000:40:01.1", "0000:41:00.0"},
			bdf:      "0000:41:00.0",
			expected: "0000:40:01.1",
		},
		{
			name:     "Single element path",
			path:     []string{"0000:00:00.0"},
			bdf:      "0000:00:00.0",
			expected: "0000:00:00.0",
		},
		{
			name:     "Empty path",
			path:     []string{},
			bdf:      "0000:00:00.0",
			expected: "0000:00:00.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputePCISwitchGroupKey(tc.path, tc.bdf)
			if result != tc.expected {
				t.Errorf("ComputePCISwitchGroupKey(%v, %s) = %s, expected %s",
					tc.path, tc.bdf, result, tc.expected)
			}
		})
	}

	// Test that devices behind the same switch get the same key
	t.Run("devices behind same switch have same key", func(t *testing.T) {
		gpu0Path := []string{"0000:00:01.1", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0"}
		gpu1Path := []string{"0000:00:01.1", "0000:01:00.0", "0000:02:01.0", "0000:04:00.0"}
		nic0Path := []string{"0000:00:01.1", "0000:01:00.0", "0000:02:04.0", "0000:07:00.0"}

		key0 := ComputePCISwitchGroupKey(gpu0Path, "0000:03:00.0")
		key1 := ComputePCISwitchGroupKey(gpu1Path, "0000:04:00.0")
		keyNIC := ComputePCISwitchGroupKey(nic0Path, "0000:07:00.0")

		if key0 != key1 {
			t.Errorf("GPU0 and GPU1 should have same key, got %s and %s", key0, key1)
		}
		if key0 != keyNIC {
			t.Errorf("GPU0 and NIC0 should have same key, got %s and %s", key0, keyNIC)
		}
	})

	// Test that devices behind different switches get different keys
	t.Run("devices behind different switches have different keys", func(t *testing.T) {
		numa0Path := []string{"0000:00:01.1", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0"}
		numa1Path := []string{"0000:80:01.1", "0000:81:00.0", "0000:82:00.0", "0000:83:00.0"}

		key0 := ComputePCISwitchGroupKey(numa0Path, "0000:03:00.0")
		key1 := ComputePCISwitchGroupKey(numa1Path, "0000:83:00.0")

		if key0 == key1 {
			t.Errorf("Devices on different NUMA nodes should have different keys, both got %s", key0)
		}
	})
}

func TestApplyEGMMemoryDevices(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return os.Stat(os.DevNull) }

	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0000:03:00.0"},
				EGMSizeBytes: 56896 * 1024 * 1024,
				NUMANode:     0,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			MemoryBacking: &api.MemoryBacking{
				HugePages:  &api.HugePages{},
				Source:     &api.MemoryBackingSource{Type: "memfd"},
				Allocation: &api.MemoryAllocation{Mode: api.MemoryAllocationModeImmediate},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	egmDevices := make([]api.MemoryDevice, 0)
	for _, md := range domain.Spec.Devices.MemoryDevices {
		if md.Model == "egm" {
			egmDevices = append(egmDevices, md)
		}
	}
	if len(egmDevices) != 1 {
		t.Fatalf("expected 1 EGM memory device, got %d", len(egmDevices))
	}
	egm := egmDevices[0]
	if egm.Access != "shared" {
		t.Errorf("expected EGM access='shared', got %q", egm.Access)
	}
	if egm.Source == nil || egm.Source.Path != "/dev/egm4" {
		t.Errorf("expected EGM source path '/dev/egm4', got %v", egm.Source)
	}
	if egm.Target == nil {
		t.Fatal("expected EGM target, got nil")
	}
	if egm.Target.Node != "0" {
		t.Errorf("expected EGM target node '0', got %q", egm.Target.Node)
	}
	expectedSizeMiB := uint64(56896)
	if egm.Target.Size.Value != expectedSizeMiB || egm.Target.Size.Unit != "MiB" {
		t.Errorf("expected EGM target size %d MiB, got %d %s", expectedSizeMiB, egm.Target.Size.Value, egm.Target.Size.Unit)
	}
	if !strings.HasPrefix(egm.Target.PCIDev, api.UserAliasPrefix) {
		t.Errorf("expected EGM pciDev to start with %q, got %q", api.UserAliasPrefix, egm.Target.PCIDev)
	}
	xmlBytes, err := xml.Marshal(egm)
	if err != nil {
		t.Fatalf("failed to marshal EGM memory device: %v", err)
	}
	xmlString := string(xmlBytes)
	if strings.Contains(xmlString, "<requested") {
		t.Fatalf("did not expect requested field in EGM XML: %s", xmlString)
	}
	if strings.Contains(xmlString, "<current") {
		t.Fatalf("did not expect current field in EGM XML: %s", xmlString)
	}
	if strings.Contains(xmlString, "<block") {
		t.Fatalf("did not expect block field in EGM XML: %s", xmlString)
	}

	if domain.Spec.MemoryBacking == nil {
		t.Fatal("expected existing MemoryBacking to be preserved")
	}
	if domain.Spec.MemoryBacking.HugePages == nil {
		t.Fatal("expected hugepages configuration to be preserved")
	}
	if domain.Spec.MemoryBacking.Source == nil || domain.Spec.MemoryBacking.Source.Type != "memfd" {
		t.Errorf("expected MemoryBacking source type 'memfd' to be preserved, got %v", domain.Spec.MemoryBacking.Source)
	}
	if domain.Spec.MemoryBacking.Allocation == nil || domain.Spec.MemoryBacking.Allocation.Mode != api.MemoryAllocationModeImmediate {
		t.Errorf("expected MemoryBacking allocation mode 'immediate' to be preserved, got %v", domain.Spec.MemoryBacking.Allocation)
	}

	// Verify GPU hostdev has the alias set
	found := false
	for i := range domain.Spec.Devices.HostDevices {
		dev := &domain.Spec.Devices.HostDevices[i]
		if dev.Alias != nil && dev.Alias.IsUserDefined() && strings.HasPrefix(dev.Alias.GetName(), egmHostdevAliasPrefix) {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected GPU hostdev to have a user-defined alias with hostdev prefix")
	}
}

func TestApplyEGMMemoryDevicesNoEGMDevicesOnHost(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return nil, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	err := ApplyNUMAHostDeviceTopology(vmi, domain)
	if err == nil {
		t.Fatal("expected EGM host discovery failure to return an error")
	}
	if !strings.Contains(err.Error(), "no /sys/class/egm devices were found") {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, md := range domain.Spec.Devices.MemoryDevices {
		if md.Model == "egm" {
			t.Fatal("expected no EGM memory devices when host has no EGM devices")
		}
	}
}

func TestApplyEGMMemoryDevicesFailsWhenEGMDeviceMissingInPod(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0000:03:00.0"},
				EGMSizeBytes: 56896 * 1024 * 1024,
				NUMANode:     0,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	err := ApplyNUMAHostDeviceTopology(vmi, domain)
	if err == nil {
		t.Fatal("expected missing /dev/egmN device to return an error")
	}
	if !strings.Contains(err.Error(), "not present in the virt-launcher pod") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyEGMMemoryDevicesDisabledWhenEGMFlagFalse(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		t.Fatal("discoverEGMDevices should not be called when EGM is disabled")
		return nil, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	for _, md := range domain.Spec.Devices.MemoryDevices {
		if md.Model == "egm" {
			t.Fatal("expected no EGM memory devices when EGM flag is not set")
		}
	}
}

func TestApplyEGMMemoryDevicesRejectsPartialSharedEGMGroup(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return os.Stat(os.DevNull) }
	stubPCIPath("0008:01:00.0", []string{"0008:00:00.0", "0008:01:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0008:01:00.0", "0009:01:00.0"},
				EGMSizeBytes: 56896 * 1024 * 1024,
				NUMANode:     0,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0008", "0x01"),
				},
			},
		},
	}

	err := ApplyNUMAHostDeviceTopology(vmi, domain)
	if err == nil {
		t.Fatal("expected partial shared EGM group selection to be rejected")
	}
	if !strings.Contains(err.Error(), "egm requires all GPUs associated with discovered host EGM devices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyEGMMemoryDevicesRejectsPartialDiscoveredEGMBoundary(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return os.Stat(os.DevNull) }

	stubPCIPath("0008:01:00.0", []string{"0008:00:00.0", "0008:01:00.0"})
	stubPCIPath("0009:01:00.0", []string{"0009:00:00.0", "0009:01:00.0"})

	totalEGMSizeBytes := uint64(56896) * 1024 * 1024
	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0008:01:00.0", "0009:01:00.0"},
				EGMSizeBytes: totalEGMSizeBytes,
				NUMANode:     0,
			},
			{
				DevPath:      "/dev/egm5",
				GPUBDFs:      []string{"0018:01:00.0", "0019:01:00.0"},
				EGMSizeBytes: totalEGMSizeBytes,
				NUMANode:     1,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0008", "0x01"),
					newTestPCIHostDevice("gpu1", "0x0009", "0x01"),
				},
			},
		},
	}

	err := ApplyNUMAHostDeviceTopology(vmi, domain)
	if err == nil {
		t.Fatal("expected partial discovered EGM boundary selection to be rejected")
	}
	if !strings.Contains(err.Error(), "egm requires all GPUs associated with discovered host EGM devices") ||
		!strings.Contains(err.Error(), "selected 0/2 GPUs sharing /dev/egm5") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestApplyEGMMemoryDevicesMultiSocket verifies that on a 2-socket GB200 system
// with 4 GPUs and 2 EGM devices (/dev/egm4 → 2 GPUs on socket 0,
// /dev/egm5 → 2 GPUs on socket 1), each GPU gets its own <memory model='egm'>
// element with size = totalEGM / numGPUsPerDevice, and the NUMA node is the
// correct per-socket guest node.
func TestApplyEGMMemoryDevicesMultiSocket(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}

	numaNodeByDomain := map[string]int{
		"0008": 0, "0009": 0,
		"0018": 1, "0019": 1,
	}
	getDeviceNumaNodeIntFunc = func(bdf string) (int, error) {
		parts := strings.SplitN(bdf, ":", 2)
		if n, ok := numaNodeByDomain[parts[0]]; ok {
			return n, nil
		}
		return 0, nil
	}
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return largeMMIOPXBIsolationThreshold, nil
	}
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return os.Stat(os.DevNull) }

	proximityByDomain := map[string]string{
		"0008": "numa-0", "0009": "numa-0",
		"0018": "numa-1", "0019": "numa-1",
	}
	getDevicePCIProximityGroupFunc = func(bdf string) (string, error) {
		parts := strings.SplitN(bdf, ":", 2)
		if g, ok := proximityByDomain[parts[0]]; ok {
			return g, nil
		}
		return "numa-0", nil
	}

	stubPCIPath("0008:01:00.0", []string{"0008:00:00.0", "0008:01:00.0"})
	stubPCIPath("0009:01:00.0", []string{"0009:00:00.0", "0009:01:00.0"})
	stubPCIPath("0018:01:00.0", []string{"0018:00:00.0", "0018:01:00.0"})
	stubPCIPath("0019:01:00.0", []string{"0019:00:00.0", "0019:01:00.0"})

	totalEGMSizeBytes := uint64(56896) * 1024 * 1024

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0008:01:00.0", "0009:01:00.0"},
				EGMSizeBytes: totalEGMSizeBytes,
				NUMANode:     0,
			},
			{
				DevPath:      "/dev/egm5",
				GPUBDFs:      []string{"0018:01:00.0", "0019:01:00.0"},
				EGMSizeBytes: totalEGMSizeBytes,
				NUMANode:     1,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-71", Memory: 268369920, Unit: "KiB"},
						{ID: "1", CPUs: "72-143", Memory: 268928000, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0008", "0x01"),
					newTestPCIHostDevice("gpu1", "0x0009", "0x01"),
					newTestPCIHostDevice("gpu2", "0x0018", "0x01"),
					newTestPCIHostDevice("gpu3", "0x0019", "0x01"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	egmDevs := make([]api.MemoryDevice, 0)
	for _, md := range domain.Spec.Devices.MemoryDevices {
		if md.Model == "egm" {
			egmDevs = append(egmDevs, md)
		}
	}
	if len(egmDevs) != 4 {
		t.Fatalf("expected 4 EGM memory devices (one per GPU), got %d", len(egmDevs))
	}

	expectedPerGPUSizeMiB := uint64(56896 / 2)

	egm4Count, egm5Count := 0, 0
	for _, e := range egmDevs {
		if e.Access != "shared" {
			t.Errorf("expected EGM access='shared', got %q", e.Access)
		}
		if e.Target == nil {
			t.Fatal("expected EGM target, got nil")
			continue
		}
		if e.Target.Size.Value != expectedPerGPUSizeMiB || e.Target.Size.Unit != "MiB" {
			t.Errorf("expected per-GPU EGM size %d MiB, got %d %s",
				expectedPerGPUSizeMiB, e.Target.Size.Value, e.Target.Size.Unit)
		}

		switch {
		case e.Source != nil && e.Source.Path == "/dev/egm4":
			egm4Count++
			if e.Target.Node != "0" {
				t.Errorf("expected /dev/egm4 GPU on guest NUMA 0, got %q", e.Target.Node)
			}
		case e.Source != nil && e.Source.Path == "/dev/egm5":
			egm5Count++
			if e.Target.Node != "1" {
				t.Errorf("expected /dev/egm5 GPU on guest NUMA 1, got %q", e.Target.Node)
			}
		default:
			t.Errorf("unexpected EGM source path: %v", e.Source)
		}
	}
	if egm4Count != 2 {
		t.Errorf("expected 2 EGM entries for /dev/egm4 (socket 0), got %d", egm4Count)
	}
	if egm5Count != 2 {
		t.Errorf("expected 2 EGM entries for /dev/egm5 (socket 1), got %d", egm5Count)
	}

	aliasCount := 0
	for i := range domain.Spec.Devices.HostDevices {
		dev := &domain.Spec.Devices.HostDevices[i]
		if dev.Alias != nil && dev.Alias.IsUserDefined() && strings.HasPrefix(dev.Alias.GetName(), egmHostdevAliasPrefix) {
			aliasCount++
		}
	}
	if aliasCount != 4 {
		t.Errorf("expected 4 GPU hostdevs with EGM aliases, got %d", aliasCount)
	}

	cell0 := mustNUMACellByID(t, domain, "0")
	if got := mustNUMADistanceValue(t, cell0, "1"); got != graceNUMADistanceRemoteNode {
		t.Fatalf("expected cell 0 -> 1 distance %d, got %d", graceNUMADistanceRemoteNode, got)
	}
	if got := mustNUMADistanceValue(t, cell0, "2"); got != graceNUMADistanceLocalToGPU {
		t.Fatalf("expected cell 0 -> 2 distance %d, got %d", graceNUMADistanceLocalToGPU, got)
	}
	if got := mustNUMADistanceValue(t, cell0, "18"); got != graceNUMADistanceRemoteToGPU {
		t.Fatalf("expected cell 0 -> 18 distance %d, got %d", graceNUMADistanceRemoteToGPU, got)
	}

	cell2 := mustNUMACellByID(t, domain, "2")
	if got := mustNUMADistanceValue(t, cell2, "9"); got != graceNUMADistanceSameGPUGroup {
		t.Fatalf("expected cell 2 -> 9 distance %d, got %d", graceNUMADistanceSameGPUGroup, got)
	}
	if got := mustNUMADistanceValue(t, cell2, "10"); got != graceNUMADistanceRemoteNode {
		t.Fatalf("expected cell 2 -> 10 distance %d, got %d", graceNUMADistanceRemoteNode, got)
	}
	if got := mustNUMADistanceValue(t, cell2, "0"); got != graceNUMADistanceLocalToGPU {
		t.Fatalf("expected cell 2 -> 0 distance %d, got %d", graceNUMADistanceLocalToGPU, got)
	}
}

func TestApplyEGMMemoryDevicesUsesExplicitEGMAssociationEvenWithoutDedicatedPXB(t *testing.T) {
	defer restoreNUMAHelpers()

	formatPCIAddressFunc = func(addr *api.Address) (string, error) {
		domain := strings.TrimPrefix(addr.Domain, "0x")
		bus := strings.TrimPrefix(addr.Bus, "0x")
		slot := strings.TrimPrefix(addr.Slot, "0x")
		function := strings.TrimPrefix(addr.Function, "0x")
		return fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function), nil
	}
	getDeviceNumaNodeIntFunc = func(string) (int, error) { return 0, nil }
	isIOMMUFDDeviceAvailableFunc = func() bool { return true }
	getDevicePCITotalMMIOSizeFunc = func(string) (uint64, error) {
		return 0, nil
	}
	getDevicePCIProximityGroupFunc = func(string) (string, error) { return "numa-0", nil }
	statEGMDevicePathFunc = func(string) (os.FileInfo, error) { return os.Stat(os.DevNull) }

	stubPCIPath("0000:03:00.0", []string{"0000:00:01.0", "0000:03:00.0"})

	discoverEGMDevicesFunc = func() ([]hardware.EGMDeviceInfo, error) {
		return []hardware.EGMDeviceInfo{
			{
				DevPath:      "/dev/egm4",
				GPUBDFs:      []string{"0000:03:00.0"},
				EGMSizeBytes: 56896 * 1024 * 1024,
				NUMANode:     0,
			},
		}, nil
	}

	vmi := &v1.VirtualMachineInstance{
		Spec: v1.VirtualMachineInstanceSpec{
			Architecture: arm64Architecture,
			Domain: v1.DomainSpec{
				CPU: &v1.CPU{
					NUMA: &v1.NUMA{
						GuestMappingPassthrough: &v1.NUMAGuestMappingPassthrough{},
					},
				},
			},
		},
	}
	vmi.Annotations = map[string]string{
		v1.GraceVirtualizationAnnotation: `{"smmuv3": true, "egm": true}`,
	}

	domain := &api.Domain{
		Spec: api.DomainSpec{
			CPU: api.CPU{
				NUMA: &api.NUMA{
					Cells: []api.NUMACell{
						{ID: "0", CPUs: "0-7", Memory: 8388608, Unit: "KiB"},
					},
				},
			},
			Devices: api.Devices{
				HostDevices: []api.HostDevice{
					newTestPCIHostDevice("gpu0", "0x0000", "0x03"),
				},
			},
		},
	}

	mustApplyNUMAHostDeviceTopology(t, vmi, domain)

	egmCount := 0
	for _, md := range domain.Spec.Devices.MemoryDevices {
		if md.Model == "egm" {
			egmCount++
		}
	}
	if egmCount != 1 {
		t.Fatalf("expected 1 EGM memory device when explicit host EGM association exists, got %d", egmCount)
	}
}
