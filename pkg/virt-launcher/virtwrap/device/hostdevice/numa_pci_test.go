package hostdevice

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/util/hardware"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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

	ApplyNUMAHostDeviceTopology(vmi, domain)

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
