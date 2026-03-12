package hostdevice

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/util"
	"kubevirt.io/kubevirt/pkg/util/hardware"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

const (
	rootBusDomain           = "0x0000"
	defaultPXBSlot          = 0x0a
	maxPXBSlot              = 0x1a // Leave 0x1b-0x1f for root hotplug ports
	maxRootBusSlot          = 0x1f
	maxRootPortSlot         = 0x1f
	pxbBusNumberBase        = 0x20 // Base bus number for PXB hierarchies (32, 64, 96, etc.)
	pxbBusNumberSpacing     = 0x20 // Space between PXB bus numbers (32 buses per hierarchy)
	maxPCIBusNumber         = 0xfe
	unifiedNUMAGroup        = 0
	rootBusControllerMarker = -1
	rootPortsPerSlot        = 8 // Number of root port functions per PXB slot

	numaPXBAliasPrefix  = "numa-pxb"
	numaRootPortPrefix  = "numa-rp"
	rootPortAliasLength = 12

	rootHotplugSlotStart           = 0x1b
	HotplugRootPortAliasPrefix     = "hotplug-rp-"
	numaHotplugAliasDiscriminator  = "numa-"
	NUMAHotplugRootPortAliasPrefix = HotplugRootPortAliasPrefix + numaHotplugAliasDiscriminator

	arm64Architecture        = "arm64"
	defaultHostDeviceIOMMUFD = "yes"

	ioResourceMemFlag = 0x00000200
	// Full-GPU passthrough on NVIDIA Grace Hopper and Grace Blackwell platforms
	// uses very large PCI memory BARs. For example, observed GB200 GPU PFs expose
	// a 256 GiB BAR2, and Hopper-class GPUs are documented as requiring similarly
	// large MMIO apertures.
	//
	// The 128 GiB value is intentionally a detection threshold, not an exact GPU
	// BAR size or hardware ABI. Devices above this threshold are treated as
	// large-MMIO GPU-like devices that need a dedicated PXB hierarchy and Grace
	// ACPI GI nodeset mapping. Longer term, this should be driven by explicit
	// host-device metadata such as resource type, vendor, and PCI class, with BAR
	// size used only as a fallback signal.
	largeMMIOPXBIsolationThresholdGiB = uint64(128)
	largeMMIOPXBIsolationThreshold    = largeMMIOPXBIsolationThresholdGiB << 30
	// NVIDIA Grace GPU passthrough requires dedicated guest ACPI Generic
	// Initiator NUMA cells per GPU. These are zero-memory guest cells referenced
	// by the hostdev <acpi nodeset='...'> element; they are not the host's HBM
	// memory NUMA nodes. Grace I/O virtualization contract uses eight GI cells
	// per passed-through GPU for driver/MIG memory onlining in the guest.
	//
	// Treat this as a current virtualization contract, not a hardware ABI for the
	// host NUMA topology. Longer term, the GI count should be discovered from host
	// sysfs/ACPI topology per GPU.
	graceGINodesPerGPU = 8
	graceGINodeSetBase = 1

	egmHostdevAliasPrefix = "hostdev"
)

var (
	formatPCIAddressFunc           = hardware.FormatPCIAddress
	getDeviceNumaNodeIntFunc       = hardware.GetDeviceNumaNodeInt
	getMdevParentPCIAddressFunc    = hardware.GetMdevParentPCIAddress
	getDevicePCIProximityGroupFunc = hardware.GetDevicePCITopologyGroup
	getDevicePCIPathHierarchyFunc  = hardware.GetDevicePCIPathHierarchy
	getDeviceIOMMUGroupInfoFunc    = hardware.GetDeviceIOMMUGroupInfo
	getDevicePCITotalMMIOSizeFunc  = getDevicePCITotalMMIOSize
	isIOMMUFDDeviceAvailableFunc   = isIOMMUFDDeviceAvailable
	discoverEGMDevicesFunc         = hardware.DiscoverEGMDevices
	statEGMDevicePathFunc          = os.Stat
)

// NormalizeHotplugRootPortAlias strips any user-alias prefixes (like "ua-") that libvirt may add.
func NormalizeHotplugRootPortAlias(alias string) string {
	normalized := strings.TrimPrefix(alias, api.UserAliasPrefix)
	for strings.HasPrefix(normalized, api.UserAliasPrefix) {
		normalized = strings.TrimPrefix(normalized, api.UserAliasPrefix)
	}
	return normalized
}

// IsHotplugRootPortAlias reports whether the provided alias denotes a planner-managed root port.
func IsHotplugRootPortAlias(alias string) bool {
	if alias == "" {
		return false
	}
	return strings.HasPrefix(NormalizeHotplugRootPortAlias(alias), HotplugRootPortAliasPrefix)
}

// IsNUMARootPortAlias reports whether the provided alias denotes a NUMA planner root port.
func IsNUMARootPortAlias(alias string) bool {
	if alias == "" {
		return false
	}
	return strings.HasPrefix(NormalizeHotplugRootPortAlias(alias), numaRootPortPrefix)
}

type numaPCIPlanner struct {
	domain              *api.Domain
	nextControllerIndex int
	nextChassis         int
	nextPXBSlot         int
	nextPCIBus          int
	usedBusSlots        map[busSlotKey]map[int]struct{}
	usedNUMAChassis     map[int]struct{}
	usedPCIBuses        map[int]struct{}
	pxbs                map[pxbKey]*pxbInfo
	rootPorts           map[deviceGroupKey]*rootPortInfo
	pcieSwitches        map[deviceGroupKey]*pcieSwitchInfo // Track switches per device group
	existingPXBs        map[string]*pxbInfo
	existingRootPorts   map[string]*rootPortInfo
	nextRootHotplugSlot int
	nextRootHotplugPort int
	// When enabled, NUMA PXB controllers are emitted without user aliases so libvirt
	// keeps the default PCI controller IDs (pci.<index>), which are required by
	// libvirt-native smmuv3 driver pciBus wiring.
	useDefaultPXBIDs bool
}

type pxbInfo struct {
	index            int
	busNr            int
	nextPortSlot     int
	nextPortFunction int // Track function number for multifunction ports
	portsCreated     int // Count of ports created for this PXB
}

type rootPortInfo struct {
	controllerIndex int
	downstreamBus   int
	pcieSwitch      *pcieSwitchInfo // Optional: PCIe switch attached to this root port
}

// pcieSwitchInfo tracks a PCIe switch hierarchy (upstream + downstream ports)
type pcieSwitchInfo struct {
	upstreamPortIndex  int                       // Controller index of upstream port
	upstreamBus        int                       // Bus number of upstream port
	downstreamPorts    []*pcieDownstreamPortInfo // Downstream ports (typically 6)
	nextDownstreamSlot int                       // Next available slot for downstream port
	nextDownstreamBus  int                       // Next available bus for downstream port
	devicesAssigned    int                       // Number of devices assigned to this switch
}

// pcieDownstreamPortInfo tracks a single downstream port on a PCIe switch
type pcieDownstreamPortInfo struct {
	controllerIndex int  // Controller index of this downstream port
	downstreamBus   int  // Bus number below this downstream port
	deviceAssigned  bool // Whether a device is assigned to this port
}

type busSlotKey struct {
	controller int
	bus        int
}

type deviceNUMAInfo struct {
	dev             *api.HostDevice
	hostNUMANode    int
	guestNUMANode   int
	bdf             string
	topologyGroup   string
	path            []string
	pathKey         string
	iommuGroup      int
	iommuPeers      []string
	mappingAccurate bool
	dedicatedPXB    bool // true for large-MMIO devices (GPU class)
	isolatedPXB     bool // true when the device should get its own PXB hierarchy
}

type deviceGroupKey struct {
	guestNUMANode int
	hostNUMANode  int
	pathKey       string
	pxbGroup      string
}

type pxbKey struct {
	guestNUMANode int
	hostNUMANode  int
	pxbGroup      string
}

func pxbAlias(hostNUMA, guestNUMA int) string {
	return fmt.Sprintf("%s-%d-%d", numaPXBAliasPrefix, hostNUMA, guestNUMA)
}

func pxbAliasForGroup(hostNUMA, guestNUMA int, pxbGroup string) string {
	if pxbGroup == "" {
		return pxbAlias(hostNUMA, guestNUMA)
	}
	sum := sha1.Sum([]byte(pxbGroup))
	return fmt.Sprintf("%s-%d-%d-%x", numaPXBAliasPrefix, hostNUMA, guestNUMA, sum[:4])
}

func rootPortAlias(key deviceGroupKey) string {
	data := fmt.Sprintf("%d-%d-%s", key.hostNUMANode, key.guestNUMANode, key.pathKey)
	sum := sha1.Sum([]byte(data))
	return fmt.Sprintf("%s-%x", numaRootPortPrefix, sum[:rootPortAliasLength/2])
}

// ComputePCISwitchGroupKey generates a grouping key for devices based on their shared PCIe switch topology.
// Devices behind the same PCIe switch (sharing the same upstream bridge) will get the same key,
// allowing them to be grouped together and attached to a virtual PCIe switch that mirrors the physical topology.
//
// Physical PCIe Switch Topology:
//
//	Root Port (0000:00:01.1)
//	  └─ PCIe Switch Upstream Port (0000:01:00.0) ← THE SWITCH
//	       ├─ Downstream Port 0 (0000:02:00.0) → Device (0000:03:00.0) GPU0
//	       ├─ Downstream Port 1 (0000:02:01.0) → Device (0000:04:00.0) GPU1
//	       ├─ Downstream Port 2 (0000:02:02.0) → Device (0000:05:00.0) GPU2
//	       └─ Downstream Port 3 (0000:02:03.0) → Device (0000:06:00.0) GPU3
//
// In the PCI path hierarchy:
//   - path[0]: Root complex port (e.g., "0000:00:01.1")
//   - path[1]: PCIe Switch Upstream Port (e.g., "0000:01:00.0") ← Group by this!
//   - path[2]: PCIe Switch Downstream Port (e.g., "0000:02:00.0", "0000:02:01.0", etc.)
//   - path[3]: Endpoint Device (e.g., "0000:03:00.0")
//
// All devices with the same path[0..1] share the same physical PCIe switch.
func ComputePCISwitchGroupKey(path []string, bdf string) string {
	if len(path) == 0 {
		return bdf
	}

	// Single element path: device directly on root bus
	if len(path) == 1 {
		return bdf
	}

	// Two element path: device directly behind root port (no switch)
	// e.g., ["0000:00:01.1", "0000:41:00.0"] for direct-attached NIC
	if len(path) == 2 {
		return strings.Join(path[:1], "/")
	}

	// For 3+ element paths, we have a switch topology:
	// path = ["root_port", "switch_upstream", "switch_downstream", "device"]
	//
	// Group by path[1] (the switch upstream port) to identify devices behind same switch
	// This correctly groups:
	//   - GPU0: ["0000:00:01.1", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0"]
	//   - GPU1: ["0000:00:01.1", "0000:01:00.0", "0000:02:01.0", "0000:04:00.0"]
	// Both group to: "0000:00:01.1/0000:01:00.0"
	if len(path) >= 3 {
		// Return path up to and including the switch upstream port
		return strings.Join(path[:2], "/")
	}

	// Fallback: use full path excluding device
	return strings.Join(path[:len(path)-1], "/")
}

func ApplyNUMAHostDeviceTopology(vmi *v1.VirtualMachineInstance, domain *api.Domain) error {
	if vmi == nil || domain == nil {
		return nil
	}
	graceCfg := util.GetGraceVirtualizationConfig(vmi)
	graceHostDevicesEnabled := isGraceHostDevicesEnabledForArch(vmi, graceCfg)
	egmEnabled := graceCfg != nil && util.GraceFieldEnabled(graceCfg.EGM)
	numaPassthroughEnabled := vmi.Spec.Domain.CPU != nil &&
		vmi.Spec.Domain.CPU.NUMA != nil &&
		vmi.Spec.Domain.CPU.NUMA.GuestMappingPassthrough != nil

	log.Log.V(1).Info("evaluating host device NUMA topology passthrough")
	if !numaPassthroughEnabled && !egmEnabled {
		log.Log.V(1).Info("NUMA host device topology not applied: guestMappingPassthrough disabled")
		return nil
	}

	numaMappingAvailable := domain.Spec.CPU.NUMA != nil && len(domain.Spec.CPU.NUMA.Cells) > 0
	if !numaMappingAvailable {
		log.Log.V(1).Info("NUMA host device topology not applied: guest NUMA mapping unavailable")
		if egmEnabled {
			return fmt.Errorf("egm requested but guest NUMA topology was not prepared before host-device placement")
		}
		return nil
	}

	hostDevices := domain.Spec.Devices.HostDevices
	if len(hostDevices) == 0 {
		log.Log.V(1).Info("NUMA host device topology not applied: no host devices present")
		if egmEnabled {
			return fmt.Errorf("egm requires passthrough host devices to be present in the domain")
		}
		return nil
	}

	log.Log.V(1).Infof("NUMA host device topology: processing %d host devices", len(hostDevices))

	planner := newNUMAPCIPlanner(domain)
	graceSMMUv3Enabled := graceHostDevicesEnabled && graceCfg != nil && util.GraceFieldEnabled(graceCfg.SMMUv3)
	// Accelerated Grace SMMUv3 and hostdev iommufd wiring require /dev/iommu inside virt-launcher.
	// If it is absent, keep the domain XML valid by generating non-accelerated SMMUv3 and omitting hostdev iommufd.
	iommuFDDeviceAvailable := isIOMMUFDDeviceAvailableFunc()
	planner.useDefaultPXBIDs = graceSMMUv3Enabled
	if graceSMMUv3Enabled && !iommuFDDeviceAvailable {
		log.Log.Warning("Grace smmuv3 requested but /dev/iommu is unavailable; falling back to smmuv3 accel=off and disabling iommufd hostdev backend")
	}
	devicesWithNUMA := make([]deviceNUMAInfo, 0, len(hostDevices))
	hostNUMANodes := make(map[int]struct{})
	guestNUMANodes := getGuestNUMANodes(domain)
	hostToGuestNUMA := getHostToGuestNUMAMap(domain)
	unmappedHostNodes := make(map[int]struct{})
	allMappingsAccurate := true

	for i := range hostDevices {
		dev := &domain.Spec.Devices.HostDevices[i]
		if dev.Type != api.HostDevicePCI && dev.Type != api.HostDeviceMDev {
			log.Log.V(1).Infof("skipping host device %d with unsupported type %s", i, dev.Type)
			continue
		}
		bdf, err := resolveHostDevicePCIAddress(dev)
		if err != nil {
			log.Log.V(1).Reason(err).Info("unable to resolve host device PCI address for NUMA planning, skipping")
			continue
		}
		numaNode, err := getDeviceNumaNodeIntFunc(bdf)
		if err != nil || numaNode < 0 {
			if err != nil {
				log.Log.V(1).Reason(err).Infof("skipping host device %s - failed to detect NUMA node", bdf)
			} else {
				log.Log.V(1).Infof("skipping host device %s - NUMA node not available", bdf)
			}
			continue
		}
		topologyGroup, err := getDevicePCIProximityGroupFunc(bdf)
		if err != nil {
			log.Log.V(1).Reason(err).Infof("using default topology grouping for host device %s (host NUMA %d)", bdf, numaNode)
			topologyGroup = fmt.Sprintf("numa-%d", numaNode)
		}

		path, err := getDevicePCIPathHierarchyFunc(bdf)
		if err != nil {
			log.Log.V(1).Reason(err).Infof("skipping host device %s - unable to derive PCI hierarchy", bdf)
			continue
		}
		pathKey := ComputePCISwitchGroupKey(path, bdf)

		iommuGroup, iommuPeers, err := getDeviceIOMMUGroupInfoFunc(bdf)
		if err != nil {
			log.Log.Reason(err).Warningf("unable to detect IOMMU group for host device %s", bdf)
		}

		guestNode, mappingAccurate := mapHostToGuestNUMANode(numaNode, guestNUMANodes, hostToGuestNUMA)
		if !mappingAccurate {
			unmappedHostNodes[numaNode] = struct{}{}
			allMappingsAccurate = false
			log.Log.V(1).Infof("host device %s (host NUMA %d) cannot be matched to a distinct guest NUMA node", bdf, numaNode)
		} else {
			log.Log.V(1).Infof("host device %s aligned with guest NUMA node %d", bdf, guestNode)
		}

		largeMMIO := shouldUseDedicatedPXBForDevice(deviceNUMAInfo{
			dev: dev,
			bdf: bdf,
		}, graceHostDevicesEnabled)

		devicesWithNUMA = append(devicesWithNUMA, deviceNUMAInfo{
			dev:             dev,
			hostNUMANode:    numaNode,
			guestNUMANode:   guestNode,
			bdf:             bdf,
			topologyGroup:   topologyGroup,
			path:            path,
			pathKey:         pathKey,
			iommuGroup:      iommuGroup,
			iommuPeers:      iommuPeers,
			mappingAccurate: mappingAccurate,
			dedicatedPXB:    largeMMIO,
			isolatedPXB:     largeMMIO,
		})
		hostNUMANodes[numaNode] = struct{}{}
	}

	if len(devicesWithNUMA) == 0 {
		log.Log.V(1).Info("NUMA host device topology not applied: no devices with NUMA affinity detected")
		if egmEnabled {
			return fmt.Errorf("egm requested but no passthrough devices with NUMA affinity were detected")
		}
		return nil
	}

	if !allMappingsAccurate {
		var nodes []string
		for node := range unmappedHostNodes {
			nodes = append(nodes, strconv.Itoa(node))
		}
		slices.Sort(nodes)
		log.Log.Infof("NUMA host device topology not applied: guest NUMA topology lacks representation for host NUMA nodes %s", strings.Join(nodes, ","))
		if egmEnabled {
			return fmt.Errorf("egm requested but guest NUMA topology lacks representation for host NUMA nodes %s", strings.Join(nodes, ","))
		}
		return nil
	}

	if domain.Spec.CPU.NUMA != nil && len(domain.Spec.CPU.NUMA.Cells) == 1 && len(hostNUMANodes) > 1 {
		var hostNodes []string
		for node := range hostNUMANodes {
			hostNodes = append(hostNodes, strconv.Itoa(node))
		}
		slices.Sort(hostNodes)
		log.Log.Infof("NUMA host device topology not applied: guest exposes a single NUMA cell but devices span host NUMA nodes %s", strings.Join(hostNodes, ","))
		if egmEnabled {
			return fmt.Errorf("egm requested but guest exposes a single NUMA cell while devices span host NUMA nodes %s", strings.Join(hostNodes, ","))
		}
		return nil
	}

	// In Grace mixed GPU+NIC topologies, isolate every passthrough device onto its own PXB hierarchy once large-MMIO and small-MMIO devices
	// are present together. This prevents shared hierarchies from becoming a contention point.
	if shouldForceIsolatedPXBsForGraceMixedTopology(graceSMMUv3Enabled, devicesWithNUMA) {
		log.Log.V(1).Info("NUMA PCI Planner: mixed Grace topology detected (large+small MMIO), forcing isolated PXB hierarchies for all passthrough devices")
		for i := range devicesWithNUMA {
			devicesWithNUMA[i].isolatedPXB = true
		}
	}

	collapseNUMA := shouldCollapseHostDeviceNUMA(vmi, domain, hostNUMANodes)
	if collapseNUMA {
		log.Log.V(1).Info("collapsing host device NUMA groups to a single guest NUMA node")
	}

	grouped := make(map[deviceGroupKey][]deviceNUMAInfo)
	for _, info := range devicesWithNUMA {
		groupKey := deviceGroupKey{
			hostNUMANode:  info.hostNUMANode,
			guestNUMANode: info.guestNUMANode,
			pathKey:       info.pathKey,
		}
		if collapseNUMA {
			log.Log.V(1).Infof("host device %s (host NUMA %d, path %q) collapsed to guest NUMA group %d", info.bdf, info.hostNUMANode, info.pathKey, unifiedNUMAGroup)
			groupKey.hostNUMANode = unifiedNUMAGroup
			groupKey.guestNUMANode = unifiedNUMAGroup
		} else {
			log.Log.V(1).Infof("host device %s grouped to host NUMA %d (guest NUMA %d) path %q", info.bdf, groupKey.hostNUMANode, groupKey.guestNUMANode, groupKey.pathKey)
		}
		if info.isolatedPXB {
			groupKey.pxbGroup = info.pathKey
		}
		grouped[groupKey] = append(grouped[groupKey], info)
	}

	groupKeys := make([]deviceGroupKey, 0, len(grouped))
	for key := range grouped {
		groupKeys = append(groupKeys, key)
	}
	slices.SortFunc(groupKeys, func(a, b deviceGroupKey) int {
		if a.guestNUMANode != b.guestNUMANode {
			if a.guestNUMANode < b.guestNUMANode {
				return -1
			}
			return 1
		}
		if a.hostNUMANode != b.hostNUMANode {
			if a.hostNUMANode < b.hostNUMANode {
				return -1
			}
			return 1
		}
		if a.pathKey == b.pathKey {
			if a.pxbGroup == b.pxbGroup {
				return 0
			}
			if a.pxbGroup < b.pxbGroup {
				return -1
			}
			return 1
		}
		if a.pathKey < b.pathKey {
			return -1
		}
		return 1
	})

	requestedDevices := make(map[string]struct{}, len(devicesWithNUMA))
	for _, info := range devicesWithNUMA {
		requestedDevices[info.bdf] = struct{}{}
	}

	// Pre-reserve PXB bus numbers to prevent conflicts with sequential allocation
	// This must be done before any root port creation to avoid bus number collisions
	planner.reservePXBBusNumbers(groupKeys)

	// Add a default root port for general-purpose device assignment
	planner.addDefaultRootPort()
	graceGINodeSetAssignments := buildGraceGINodeSetAssignments(domain, devicesWithNUMA, graceHostDevicesEnabled)

	for _, key := range groupKeys {
		infos := grouped[key]
		if len(infos) == 0 {
			continue
		}

		pathLabel := infos[0].pathKey
		if len(infos[0].path) > 0 {
			pathLabel = strings.Join(infos[0].path, " -> ")
		}

		log.Log.V(1).Infof("setting up PCI expander bus for host NUMA %d (guest NUMA %d) path %q with %d devices",
			key.hostNUMANode, key.guestNUMANode, pathLabel, len(infos))

		pxb, err := planner.ensurePXBForGroup(key.hostNUMANode, key.guestNUMANode, key.pxbGroup)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to create PCI expander bus for host NUMA %d (guest NUMA %d)", key.hostNUMANode, key.guestNUMANode)
			continue
		}

		// Determine if we need a PCIe switch for this group
		// Physical topology patterns:
		//   - Multiple devices behind same PCIe switch (e.g., HGX with 4-6 devices per switch)
		//     → len(infos) > 1 → Create virtual PCIe switch to mirror physical topology
		//   - Single device directly on root port (no physical switch)
		//     → len(infos) == 1 → No virtual switch, direct root port attachment
		needsSwitch := len(infos) > 1

		if needsSwitch {
			log.Log.V(1).Infof("Device group has %d devices sharing path %q - will create PCIe switch (mirroring physical topology)",
				len(infos), pathLabel)
		} else {
			log.Log.V(1).Infof("Device group has 1 device on path %q - will attach directly to root port (no switch needed)",
				pathLabel)
		}

		rootPort, err := planner.ensureRootPortForGroupWithSwitch(pxb, key, needsSwitch, len(infos))
		if err != nil {
			log.Log.Reason(err).Error("failed to allocate root port for host device path")
			continue
		}

		for _, info := range infos {
			if info.iommuGroup >= 0 {
				for _, peer := range info.iommuPeers {
					if peer == info.bdf {
						continue
					}
					if _, present := requestedDevices[peer]; !present {
						log.Log.Warningf("host device %s belongs to IOMMU group %d with peer %s that is not attached to the VMI", info.bdf, info.iommuGroup, peer)
					}
				}
			}

			assignHostDeviceToRootPort(info.dev, rootPort)
			if graceHostDevicesEnabled {
				enableIOMMUFD := iommuFDDeviceAvailable && (info.dedicatedPXB || graceSMMUv3Enabled)
				applyGraceHostDeviceSettings(info.dev, key.guestNUMANode, graceGINodeSetAssignments[info.dev], enableIOMMUFD)
			}
			log.Log.V(1).Infof("assigned host device %s to host NUMA %d (guest NUMA %d) via controller %d", info.bdf, key.hostNUMANode, key.guestNUMANode, rootPort.controllerIndex)
		}
	}

	applyGraceSMMUv3IOMMUTopology(domain, graceCfg, graceHostDevicesEnabled, iommuFDDeviceAvailable)

	if egmEnabled && graceHostDevicesEnabled {
		if err := applyEGMMemoryDevices(domain, devicesWithNUMA); err != nil {
			return err
		}
	}
	return nil
}

func newNUMAPCIPlanner(domain *api.Domain) *numaPCIPlanner {
	planner := &numaPCIPlanner{
		domain:              domain,
		nextPXBSlot:         defaultPXBSlot,
		nextPCIBus:          pxbBusNumberBase,
		usedBusSlots:        map[busSlotKey]map[int]struct{}{},
		usedNUMAChassis:     map[int]struct{}{},
		usedPCIBuses:        map[int]struct{}{},
		pxbs:                map[pxbKey]*pxbInfo{},
		rootPorts:           map[deviceGroupKey]*rootPortInfo{},
		pcieSwitches:        map[deviceGroupKey]*pcieSwitchInfo{},
		existingPXBs:        map[string]*pxbInfo{},
		existingRootPorts:   map[string]*rootPortInfo{},
		nextChassis:         1,
		nextRootHotplugSlot: rootHotplugSlotStart,
		nextRootHotplugPort: 0,
	}

	maxIndex := -1
	hasPCIeRoot := false
	for i := range domain.Spec.Devices.Controllers {
		ctrl := domain.Spec.Devices.Controllers[i]
		planner.reserveRootSlotIfNeeded(ctrl.Address)
		idx, err := strconv.Atoi(ctrl.Index)
		if err == nil && idx > maxIndex {
			maxIndex = idx
		}
		// Check if pcie-root (index 0) is already explicitly defined
		if idx == 0 && ctrl.Model == "pcie-root" {
			hasPCIeRoot = true
		}
		if ctrl.Address != nil {
			parentIdx := parseControllerIndex(ctrl.Address.Controller, rootBusControllerMarker)
			if busVal, err := parseBusNumber(ctrl.Address.Bus); err == nil {
				planner.reservePCIBus(busVal)
				if ctrl.Address.Slot != "" {
					if slotVal, err := strconv.ParseInt(strings.TrimPrefix(ctrl.Address.Slot, "0x"), 16, 32); err == nil {
						planner.markBusSlot(parentIdx, busVal, int(slotVal))
						if parentIdx == rootBusControllerMarker && busVal == 0 &&
							int(slotVal) >= planner.nextPXBSlot && int(slotVal) <= maxPXBSlot {
							planner.nextPXBSlot = int(slotVal) + 1
						}
					}
				}
			}
		}
		// Track existing chassis numbers to avoid conflicts
		if ctrl.Target != nil && ctrl.Target.Chassis != "" {
			if chassisVal, err := strconv.Atoi(ctrl.Target.Chassis); err == nil {
				planner.usedNUMAChassis[chassisVal] = struct{}{}
				if chassisVal >= planner.nextChassis {
					planner.nextChassis = chassisVal + 1
				}
			}
		}
		if ctrl.Target != nil && ctrl.Target.BusNr != "" {
			if busNr, err := parseBusNumber(ctrl.Target.BusNr); err == nil {
				planner.reservePCIBus(busNr)
			}
		}

		if ctrl.Alias != nil && err == nil {
			aliasName := ctrl.Alias.GetName()
			if ctrl.Model == "pcie-expander-bus" && strings.HasPrefix(aliasName, numaPXBAliasPrefix) {
				if ctrl.Target != nil && ctrl.Target.BusNr != "" {
					if busNr, err2 := parseBusNumber(ctrl.Target.BusNr); err2 == nil {
						planner.existingPXBs[aliasName] = &pxbInfo{
							index:            idx,
							busNr:            busNr,
							nextPortSlot:     0,
							nextPortFunction: 0,
							portsCreated:     0,
						}
					}
				}
			}
			if ctrl.Model == "pcie-root-port" && strings.HasPrefix(aliasName, numaRootPortPrefix) {
				downstreamBus := -1
				if ctrl.Target != nil && ctrl.Target.BusNr != "" {
					if busNr, err2 := parseBusNumber(ctrl.Target.BusNr); err2 == nil {
						downstreamBus = busNr
						planner.reservePCIBus(busNr)
					}
				}
				if ctrl.Address != nil {
					parentIdx := parseControllerIndex(ctrl.Address.Controller, idx)
					if ctrl.Address.Bus != "" {
						if busVal, err2 := parseBusNumber(ctrl.Address.Bus); err2 == nil {
							if ctrl.Address.Slot != "" {
								if slotVal, err3 := strconv.ParseInt(strings.TrimPrefix(ctrl.Address.Slot, "0x"), 16, 32); err3 == nil {
									planner.markBusSlot(parentIdx, busVal, int(slotVal))
								}
							}
						}
					} else if ctrl.Address.Slot != "" {
						if slotVal, err2 := strconv.ParseInt(strings.TrimPrefix(ctrl.Address.Slot, "0x"), 16, 32); err2 == nil {
							planner.markBusSlot(parentIdx, 0, int(slotVal))
						}
					}
				}
				if downstreamBus >= 0 {
					planner.existingRootPorts[aliasName] = &rootPortInfo{
						controllerIndex: idx,
						downstreamBus:   downstreamBus,
					}
				}
			} else if ctrl.Model == "pcie-root-port" && IsHotplugRootPortAlias(aliasName) {
				if ctrl.Target != nil && ctrl.Target.BusNr != "" {
					if busNr, err2 := parseBusNumber(ctrl.Target.BusNr); err2 == nil {
						planner.reservePCIBus(busNr)
					}
				}
				if ctrl.Address != nil && ctrl.Address.Slot != "" {
					if slotVal, err2 := strconv.ParseInt(strings.TrimPrefix(ctrl.Address.Slot, "0x"), 16, 32); err2 == nil {
						slot := int(slotVal)
						planner.markBusSlot(rootBusControllerMarker, 0, slot)
						if slot >= planner.nextRootHotplugSlot {
							planner.nextRootHotplugSlot = slot + 1
						}
					}
				}
				if ctrl.Target != nil && ctrl.Target.Port != "" {
					if portVal, err2 := strconv.ParseInt(strings.TrimPrefix(ctrl.Target.Port, "0x"), 16, 32); err2 == nil {
						port := int(portVal)
						if port >= planner.nextRootHotplugPort {
							planner.nextRootHotplugPort = port + 1
						}
					}
				}
			}
		}
	}

	for i := range domain.Spec.Devices.HostDevices {
		dev := &domain.Spec.Devices.HostDevices[i]
		planner.reserveRootSlotIfNeeded(dev.Address)
		if dev.Address != nil && dev.Address.Bus != "" {
			if busVal, err := parseBusNumber(dev.Address.Bus); err == nil {
				planner.reservePCIBus(busVal)
				if dev.Address.Slot != "" {
					if slotVal, err := strconv.ParseInt(strings.TrimPrefix(dev.Address.Slot, "0x"), 16, 32); err == nil {
						parentIdx := parseControllerIndex(dev.Address.Controller, rootBusControllerMarker)
						planner.markBusSlot(parentIdx, busVal, int(slotVal))
					}
				}
			}
		}
	}

	// Also check other device types that might use root bus addresses
	for i := range domain.Spec.Devices.Interfaces {
		intf := &domain.Spec.Devices.Interfaces[i]
		planner.reserveRootSlotIfNeeded(intf.Address)
		if intf.Address != nil && intf.Address.Bus != "" {
			if busVal, err := parseBusNumber(intf.Address.Bus); err == nil {
				if intf.Address.Slot != "" {
					if slotVal, err := strconv.ParseInt(strings.TrimPrefix(intf.Address.Slot, "0x"), 16, 32); err == nil {
						parentIdx := parseControllerIndex(intf.Address.Controller, rootBusControllerMarker)
						planner.markBusSlot(parentIdx, busVal, int(slotVal))
					}
				}
			}
		}
	}

	for alias, info := range planner.existingPXBs {
		info.nextPortSlot = planner.nextFreeSlot(info.index, info.busNr)
		info.portsCreated = len(planner.usedBusSlots[busSlotKey{controller: info.index, bus: info.busNr}])
		planner.existingPXBs[alias] = info
	}

	planner.nextControllerIndex = maxIndex + 1
	if planner.nextPXBSlot > maxPXBSlot {
		planner.nextPXBSlot = maxPXBSlot
	}
	if planner.nextPCIBus < pxbBusNumberBase {
		planner.nextPCIBus = pxbBusNumberBase
	}

	// Reserve slot 0x00 for pcie-root controller
	planner.markBusSlot(rootBusControllerMarker, 0, 0x00)

	// Note: We do NOT reserve Q35 built-in device slots (0x1b, 0x1f, etc.) because:
	// 1. These are implicit devices that libvirt/QEMU create automatically
	// 2. Libvirt already knows about them and will avoid them during address allocation
	// 3. Explicitly marking them here can interfere with hotplug port allocation (0x10-0x1f range)
	// 4. Our planner only needs to track slots for controllers we explicitly create

	// Ensure pcie-root controller (index 0) is explicitly defined when needed
	// This is required when other controllers explicitly reference it with controller="0"
	if !hasPCIeRoot && maxIndex >= 0 {
		log.Log.V(1).Info("NUMA PCI Planner: Adding explicit pcie-root controller (index 0)")
		pcieRoot := api.Controller{
			Type:  "pci",
			Index: "0",
			Model: "pcie-root",
		}
		// Prepend to ensure index 0 comes first
		domain.Spec.Devices.Controllers = append([]api.Controller{pcieRoot}, domain.Spec.Devices.Controllers...)
	}

	log.Log.V(1).Infof("NUMA PCI Planner: PXB slots available: 0x%02x-0x%02x, reserved root slots: 0x01,0x1b-0x1f",
		defaultPXBSlot, maxPXBSlot)

	return planner
}

func (p *numaPCIPlanner) ensurePXB(hostNUMA, guestNUMA int) (*pxbInfo, error) {
	return p.ensurePXBForGroup(hostNUMA, guestNUMA, "")
}

func (p *numaPCIPlanner) ensurePXBForGroup(hostNUMA, guestNUMA int, pxbGroup string) (*pxbInfo, error) {
	key := pxbKey{
		hostNUMANode:  hostNUMA,
		guestNUMANode: guestNUMA,
		pxbGroup:      pxbGroup,
	}
	if existing, ok := p.pxbs[key]; ok {
		return existing, nil
	}
	alias := pxbAliasForGroup(hostNUMA, guestNUMA, pxbGroup)
	if existing, ok := p.existingPXBs[alias]; ok {
		info := &pxbInfo{
			index:            existing.index,
			busNr:            existing.busNr,
			nextPortSlot:     p.nextFreeSlot(existing.index, existing.busNr),
			nextPortFunction: 0,
			portsCreated:     len(p.usedBusSlots[busSlotKey{controller: existing.index, bus: existing.busNr}]),
		}
		p.pxbs[key] = info
		return info, nil
	}
	index := p.nextControllerIndex
	p.nextControllerIndex++

	slot := p.allocateRootSlot()
	if slot < 0 {
		return nil, fmt.Errorf("no PXB slots available on root bus (slots 0x%02x-0x%02x reserved for hotplug)", rootHotplugSlotStart, maxRootBusSlot)
	}

	baseBusNr := pxbBusNumberBase + (hostNUMA * pxbBusNumberSpacing)
	if baseBusNr > maxPCIBusNumber {
		return nil, fmt.Errorf("calculated PXB base bus number 0x%02x exceeds maximum 0x%02x for host NUMA %d", baseBusNr, maxPCIBusNumber, hostNUMA)
	}

	// Keep historical bus placement for shared-per-NUMA PXBs, but allocate distinct buses
	// for dedicated high-MMIO groups to prevent large BAR endpoints from competing for the
	// same host bridge aperture.
	busNr := baseBusNr
	if pxbGroup != "" {
		dedicatedBusNr, err := p.allocateDedicatedPXBBusNumber(baseBusNr)
		if err != nil {
			return nil, err
		}
		busNr = dedicatedBusNr
	}

	p.reservePCIBus(busNr)

	nodeVal := guestNUMA
	controller := api.Controller{
		Type:  "pci",
		Index: strconv.Itoa(index),
		Model: "pcie-expander-bus",
		ModelInfo: &api.ControllerModel{
			Name: "pxb-pcie",
		},
		Target: &api.ControllerTarget{
			BusNr: strconv.Itoa(busNr),
			Node:  &nodeVal,
		},
		Address: &api.Address{
			Type:     api.AddressPCI,
			Domain:   rootBusDomain,
			Bus:      "0x00",
			Slot:     fmt.Sprintf("0x%02x", slot),
			Function: "0x0",
		},
	}
	if !p.useDefaultPXBIDs {
		controller.Alias = api.NewUserDefinedAlias(alias)
	}

	p.domain.Spec.Devices.Controllers = append(p.domain.Spec.Devices.Controllers, controller)

	info := &pxbInfo{
		index:            index,
		busNr:            busNr,
		nextPortSlot:     0,
		nextPortFunction: 0,
		portsCreated:     0,
	}
	p.pxbs[key] = info
	p.existingPXBs[alias] = info
	log.Log.V(1).Infof("NUMA PCI Planner: Created PXB controller index=%d, busNr=%d (0x%02x), slot=0x%02x", index, busNr, busNr, slot)
	return info, nil
}

func (p *numaPCIPlanner) allocateDedicatedPXBBusNumber(baseBusNr int) (int, error) {
	lower := baseBusNr + 1
	// Keep one bus of headroom so firmware/qemu auto-assigned downstream buses for
	// the first root port on this PXB do not collide with the next NUMA base bus.
	upper := baseBusNr + pxbBusNumberSpacing - 2
	if upper > maxPCIBusNumber {
		upper = maxPCIBusNumber
	}
	// Allocate every other bus (odd offset from base). This avoids adjacent PXB bus
	// numbers like 0x41/0x42, which can conflict with libvirt/qemu auto secondary-bus
	// assignment for root ports behind the first PXB.
	for bus := lower; bus <= upper; bus += 2 {
		if _, used := p.usedPCIBuses[bus]; !used {
			return bus, nil
		}
	}
	return 0, fmt.Errorf("no free dedicated PXB bus available in range 0x%02x-0x%02x", lower, upper)
}

func (p *numaPCIPlanner) ensureRootPortForGroup(pxb *pxbInfo, key deviceGroupKey) (*rootPortInfo, error) {
	return p.ensureRootPortForGroupWithSwitch(pxb, key, false, 1)
}

// ensureRootPortForGroupWithSwitch creates or retrieves a root port for the given device group,
// optionally creating a PCIe switch if multiple devices share the same path (mirroring HGX topology)
func (p *numaPCIPlanner) ensureRootPortForGroupWithSwitch(pxb *pxbInfo, key deviceGroupKey, needsSwitch bool, numDevices int) (*rootPortInfo, error) {
	if existing, ok := p.rootPorts[key]; ok {
		return existing, nil
	}
	alias := rootPortAlias(key)
	if existing, ok := p.existingRootPorts[alias]; ok {
		p.rootPorts[key] = existing
		return existing, nil
	}
	rootPort, err := p.addRootPort(pxb, alias)
	if err != nil {
		return nil, err
	}

	// If this group has multiple devices sharing the same PCIe path (like physical HGX topology),
	// create a PCIe switch to mirror the physical topology
	if needsSwitch && numDevices > 1 {
		log.Log.V(1).Infof("Creating PCIe switch for device group with %d devices (mirroring HGX topology)", numDevices)

		// HGX switches typically have 4 downstream ports per NVIDIA documentation
		// Use a power of 2 that fits: 2, 4, 8, 16, 32
		numPorts := 4 // NVIDIA standard for HGX
		if numDevices > 4 {
			// Round up to next power of 2
			numPorts = 8
		}
		if numDevices > 8 {
			numPorts = 16
		}
		if numDevices > 16 {
			numPorts = 32
		}

		switchInfo, err := p.addPCIeSwitch(rootPort, numPorts)
		if err != nil {
			log.Log.Reason(err).Warningf("Failed to create PCIe switch for group, falling back to direct attachment")
		} else {
			rootPort.pcieSwitch = switchInfo
			p.pcieSwitches[key] = switchInfo
			log.Log.V(1).Infof("Created PCIe switch with %d downstream ports for device group", numPorts)
		}
	}

	p.rootPorts[key] = rootPort
	p.existingRootPorts[alias] = rootPort
	return rootPort, nil
}

func (p *numaPCIPlanner) addRootPort(info *pxbInfo, alias string) (*rootPortInfo, error) {
	index := p.nextControllerIndex
	p.nextControllerIndex++

	// For multifunction ports, always use slot 0x00, incrementing function
	slot := 0x00
	function := info.nextPortFunction

	// Validate we haven't exceeded multifunction limits
	if function >= rootPortsPerSlot {
		return nil, fmt.Errorf("no more root port functions available on PXB (max %d per slot)", rootPortsPerSlot)
	}

	// Find next available chassis number
	chassis := p.allocateChassis()
	if chassis < 0 {
		return nil, fmt.Errorf("no more chassis numbers available")
	}

	// Use function number as port value for target
	portHex := fmt.Sprintf("0x%x", function)

	info.nextPortFunction++
	info.portsCreated++

	downstreamBus, err := p.allocatePCIBusNumber()
	if err != nil {
		return nil, err
	}

	controller := api.Controller{
		Type:  "pci",
		Index: strconv.Itoa(index),
		Model: "pcie-root-port",
		ModelInfo: &api.ControllerModel{
			Name: "pcie-root-port",
		},
		Target: &api.ControllerTarget{
			Chassis: strconv.Itoa(chassis),
			Port:    portHex,
			// Don't set BusNr for pcie-root-port - libvirt doesn't support it
			// Only pci-expander-bus/pcie-expander-bus controllers use busNr
		},
		Alias: api.NewUserDefinedAlias(alias),
		Address: &api.Address{
			Type:   api.AddressPCI,
			Domain: rootBusDomain,
			// Set bus to the PXB controller INDEX (not busNr!) - this tells libvirt to place the root port on the PXB
			Bus:      strconv.Itoa(info.index),
			Slot:     fmt.Sprintf("0x%02x", slot),
			Function: fmt.Sprintf("0x%x", function),
		},
	}

	// Enable multifunction for function 0 to allow additional functions on this slot
	if function == 0 {
		controller.Address.MultiFunction = "on"
	}

	p.domain.Spec.Devices.Controllers = append(p.domain.Spec.Devices.Controllers, controller)

	log.Log.V(1).Infof("NUMA PCI Planner: Created root port index=%d, parent=index=%d(busNr=%d), slot=0x%02x, function=0x%x",
		index, info.index, info.busNr, slot, function)

	return &rootPortInfo{
		controllerIndex: index,
		downstreamBus:   downstreamBus,
	}, nil
}

// addPCIeSwitch creates a complete PCIe switch hierarchy (upstream + downstream ports)
// attached to the specified root port. This mirrors physical HGX topology where devices
// behind a PCIe switch share the same upstream connection.
func (p *numaPCIPlanner) addPCIeSwitch(rootPort *rootPortInfo, numDownstreamPorts int) (*pcieSwitchInfo, error) {
	if numDownstreamPorts < 1 || numDownstreamPorts > 32 {
		return nil, fmt.Errorf("invalid number of downstream ports: %d (must be 1-32)", numDownstreamPorts)
	}

	// Create upstream port attached to the root port
	upstreamIndex := p.nextControllerIndex
	p.nextControllerIndex++

	upstreamBus, err := p.allocatePCIBusNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate bus number for PCIe switch upstream port: %v", err)
	}

	upstreamChassis := p.allocateChassis()
	if upstreamChassis < 0 {
		return nil, fmt.Errorf("failed to allocate chassis for PCIe switch upstream port")
	}

	// PCIe switch upstream port - connects to root port
	// Use x3130-upstream which is the QEMU device model for PCIe switch upstream port
	// Note: upstream ports don't use chassis/port in target (only downstream ports do)
	upstreamController := api.Controller{
		Type:  "pci",
		Index: strconv.Itoa(upstreamIndex),
		Model: "pcie-switch-upstream-port",
		ModelInfo: &api.ControllerModel{
			Name: "x3130-upstream",
		},
		// No Target for upstream port - it doesn't support chassis/port attributes
		Alias: api.NewUserDefinedAlias(fmt.Sprintf("numa-switch-up-%d", upstreamIndex)),
		Address: &api.Address{
			Type:     api.AddressPCI,
			Domain:   rootBusDomain,
			Bus:      strconv.Itoa(rootPort.controllerIndex), // Attach to root port by controller index
			Slot:     "0x00",
			Function: "0x0",
		},
	}

	p.domain.Spec.Devices.Controllers = append(p.domain.Spec.Devices.Controllers, upstreamController)

	log.Log.V(1).Infof("NUMA PCI Planner: Created PCIe switch upstream port index=%d, attached to root port controller=%d",
		upstreamIndex, rootPort.controllerIndex)

	// Create downstream ports
	switchInfo := &pcieSwitchInfo{
		upstreamPortIndex:  upstreamIndex,
		upstreamBus:        upstreamBus,
		downstreamPorts:    make([]*pcieDownstreamPortInfo, 0, numDownstreamPorts),
		nextDownstreamSlot: 0,
		nextDownstreamBus:  upstreamBus + 1,
	}

	for i := 0; i < numDownstreamPorts; i++ {
		downstreamPort, err := p.addPCIeDownstreamPort(switchInfo, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create downstream port %d: %v", i, err)
		}
		switchInfo.downstreamPorts = append(switchInfo.downstreamPorts, downstreamPort)
	}

	log.Log.V(1).Infof("NUMA PCI Planner: Created PCIe switch with upstream port index=%d and %d downstream ports",
		upstreamIndex, numDownstreamPorts)

	return switchInfo, nil
}

// addPCIeDownstreamPort creates a single PCIe switch downstream port
func (p *numaPCIPlanner) addPCIeDownstreamPort(switchInfo *pcieSwitchInfo, portNum int) (*pcieDownstreamPortInfo, error) {
	downstreamIndex := p.nextControllerIndex
	p.nextControllerIndex++

	downstreamBus, err := p.allocatePCIBusNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate bus number for downstream port %d: %v", portNum, err)
	}

	downstreamChassis := p.allocateChassis()
	if downstreamChassis < 0 {
		return nil, fmt.Errorf("failed to allocate chassis for downstream port %d", portNum)
	}

	// PCIe switch downstream port - connects to upstream port
	// Use xio3130-downstream which is the QEMU device model for PCIe switch downstream port
	downstreamController := api.Controller{
		Type:  "pci",
		Index: strconv.Itoa(downstreamIndex),
		Model: "pcie-switch-downstream-port",
		ModelInfo: &api.ControllerModel{
			Name: "xio3130-downstream",
		},
		Target: &api.ControllerTarget{
			Chassis: strconv.Itoa(downstreamChassis),
			Port:    fmt.Sprintf("0x%x", portNum),
		},
		Alias: api.NewUserDefinedAlias(fmt.Sprintf("numa-switch-down-%d-%d", switchInfo.upstreamPortIndex, portNum)),
		Address: &api.Address{
			Type:     api.AddressPCI,
			Domain:   rootBusDomain,
			Bus:      strconv.Itoa(switchInfo.upstreamPortIndex), // Attach to upstream port by controller index
			Slot:     fmt.Sprintf("0x%02x", portNum),
			Function: "0x0",
		},
	}

	p.domain.Spec.Devices.Controllers = append(p.domain.Spec.Devices.Controllers, downstreamController)

	log.Log.V(1).Infof("NUMA PCI Planner: Created PCIe switch downstream port index=%d, port=%d, chassis=%d, attached to upstream controller=%d",
		downstreamIndex, portNum, downstreamChassis, switchInfo.upstreamPortIndex)

	return &pcieDownstreamPortInfo{
		controllerIndex: downstreamIndex,
		downstreamBus:   downstreamBus,
		deviceAssigned:  false,
	}, nil
}

func (p *numaPCIPlanner) allocateRootSlot() int {
	for slot := p.nextPXBSlot; slot <= maxPXBSlot; slot++ {
		if !p.isBusSlotUsed(rootBusControllerMarker, 0, slot) {
			p.markBusSlot(rootBusControllerMarker, 0, slot)
			p.nextPXBSlot = slot + 1
			return slot
		}
	}
	return -1
}

func (p *numaPCIPlanner) addDefaultRootPort() {
	// Add an explicit pcie-root-port on the root bus at slot 0x01
	// Libvirt requires this for topology validation
	index := p.nextControllerIndex
	p.nextControllerIndex++

	chassis := p.allocateChassis()
	if chassis < 0 {
		log.Log.Warning("Unable to allocate chassis for default root port")
		return
	}

	// Don't allocate a bus number - libvirt will auto-assign for default root port

	controller := api.Controller{
		Type:  "pci",
		Index: strconv.Itoa(index),
		Model: "pcie-root-port",
		ModelInfo: &api.ControllerModel{
			Name: "pcie-root-port",
		},
		Target: &api.ControllerTarget{
			Chassis: strconv.Itoa(chassis),
			Port:    "0x0",
			// Don't set BusNr for default root port - let libvirt auto-assign
		},
		Alias: api.NewUserDefinedAlias("default-root-port"),
		Address: &api.Address{
			Type:       api.AddressPCI,
			Domain:     rootBusDomain,
			Bus:        "0x00",
			Slot:       "0x01",
			Function:   "0x0",
			Controller: "0", // Explicitly connect to pcie-root (index 0)
		},
	}

	p.domain.Spec.Devices.Controllers = append(p.domain.Spec.Devices.Controllers, controller)
	p.markBusSlot(rootBusControllerMarker, 0, 0x01)
	// Don't reserve a bus number - libvirt will auto-assign

	log.Log.V(1).Infof("NUMA PCI Planner: Created default root port index=%d at slot 0x01", index)
}

func (p *numaPCIPlanner) allocateChassis() int {
	for chassis := p.nextChassis; chassis <= 255; chassis++ {
		if _, used := p.usedNUMAChassis[chassis]; !used {
			p.usedNUMAChassis[chassis] = struct{}{}
			p.nextChassis = chassis + 1
			return chassis
		}
	}
	return -1
}

func (p *numaPCIPlanner) reservePXBBusNumbers(groupKeys []deviceGroupKey) {
	// Pre-reserve bus numbers that will be used by PXBs to prevent conflicts
	// with sequentially allocated bus numbers for other controllers.
	//
	// Also reserve an initial downstream window (base+1, base+2, ...) for the
	// shared-per-NUMA PXB root ports. Without this, dedicated large-MMIO PXBs can
	// consume those low buses (for example 0x41/0x43 on NUMA base 0x40), which can
	// collide with firmware/libvirt secondary-bus assignment on the shared PXB.
	numaNodes := make(map[int]struct{})
	sharedGroupsPerNUMA := make(map[int]int)
	for _, key := range groupKeys {
		numaNodes[key.hostNUMANode] = struct{}{}
		if key.pxbGroup == "" {
			sharedGroupsPerNUMA[key.hostNUMANode]++
		}
	}

	for numaNode := range numaNodes {
		busNr := pxbBusNumberBase + (numaNode * pxbBusNumberSpacing)
		if busNr <= maxPCIBusNumber {
			p.reservePCIBus(busNr)
			log.Log.V(1).Infof("NUMA PCI Planner: Pre-reserved PXB bus 0x%02x for NUMA node %d", busNr, numaNode)
		}

		sharedGroups := sharedGroupsPerNUMA[numaNode]
		for offset := 1; offset <= sharedGroups; offset++ {
			downstreamBus := busNr + offset
			if downstreamBus > maxPCIBusNumber {
				break
			}
			p.reservePCIBus(downstreamBus)
		}
		if sharedGroups > 0 {
			lastReserved := busNr + sharedGroups
			if lastReserved > maxPCIBusNumber {
				lastReserved = maxPCIBusNumber
			}
			log.Log.V(1).Infof(
				"NUMA PCI Planner: Reserved shared-PXB downstream bus window 0x%02x-0x%02x for NUMA node %d",
				busNr+1, lastReserved, numaNode,
			)
		}
	}
}

func (p *numaPCIPlanner) reservePCIBus(bus int) {
	if bus < 0 || bus > maxPCIBusNumber {
		return
	}
	if _, used := p.usedPCIBuses[bus]; used {
		return
	}
	p.usedPCIBuses[bus] = struct{}{}
	// Do NOT advance nextPCIBus here - we want sequential allocation to skip
	// reserved buses but not jump past them. allocatePCIBusNumber() will skip
	// over reserved buses naturally by checking usedPCIBuses.
}

func (p *numaPCIPlanner) allocatePCIBusNumber() (int, error) {
	if p.nextPCIBus < pxbBusNumberBase {
		p.nextPCIBus = pxbBusNumberBase
	}
	for bus := p.nextPCIBus; bus <= maxPCIBusNumber; bus++ {
		if _, used := p.usedPCIBuses[bus]; !used {
			p.usedPCIBuses[bus] = struct{}{}
			p.nextPCIBus = bus + 1
			return bus, nil
		}
	}
	return -1, fmt.Errorf("no PCI bus numbers available")
}

func (p *numaPCIPlanner) allocatePXBRootPortSlot(info *pxbInfo) (int, error) {
	start := info.nextPortSlot
	if start < 0 {
		start = 0
	}
	for slot := start; slot <= maxRootPortSlot; slot++ {
		if p.isBusSlotUsed(info.index, info.busNr, slot) {
			continue
		}
		p.markBusSlot(info.index, info.busNr, slot)
		info.nextPortSlot = slot + 1
		return slot, nil
	}
	return -1, fmt.Errorf("no more slots available on NUMA expander bus")
}

func (p *numaPCIPlanner) markBusSlot(controller, bus, slot int) {
	if bus < 0 || slot < 0 {
		return
	}
	key := busSlotKey{controller: controller, bus: bus}
	m, ok := p.usedBusSlots[key]
	if !ok {
		m = make(map[int]struct{})
		p.usedBusSlots[key] = m
	}
	m[slot] = struct{}{}
}

func (p *numaPCIPlanner) isBusSlotUsed(controller, bus, slot int) bool {
	if slot < 0 {
		return true
	}
	key := busSlotKey{controller: controller, bus: bus}
	if m, ok := p.usedBusSlots[key]; ok {
		_, used := m[slot]
		return used
	}
	return false
}

func (p *numaPCIPlanner) nextFreeSlot(controller, bus int) int {
	for slot := 0; slot <= maxRootPortSlot; slot++ {
		if !p.isBusSlotUsed(controller, bus, slot) {
			return slot
		}
	}
	return maxRootPortSlot + 1
}

func parseBusNumber(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("empty bus value")
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		val, err := strconv.ParseInt(trimmed[2:], 16, 32)
		if err != nil {
			return 0, err
		}
		return int(val), nil
	}
	if val, err := strconv.ParseInt(trimmed, 10, 32); err == nil {
		return int(val), nil
	}
	if val, err := strconv.ParseInt(trimmed, 16, 32); err == nil {
		return int(val), nil
	}
	return 0, fmt.Errorf("invalid bus value %q", value)
}

func parseControllerIndex(value string, defaultVal int) int {
	if strings.TrimSpace(value) == "" {
		return defaultVal
	}
	idx, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return idx
}

// assignHostDeviceToRootPort wires a host device onto the downstream bus of a NUMA-aware root port.
func assignHostDeviceToRootPort(dev *api.HostDevice, port *rootPortInfo) {
	if dev.Address == nil {
		dev.Address = &api.Address{}
	}
	dev.Address.Type = api.AddressPCI
	dev.Address.Domain = rootBusDomain

	// If the root port has a PCIe switch, assign device to next available downstream port
	if port.pcieSwitch != nil {
		// Find next available downstream port
		for _, downstreamPort := range port.pcieSwitch.downstreamPorts {
			if !downstreamPort.deviceAssigned {
				// Use Bus attribute with the downstream port's index (NVIDIA docs pattern)
				// Example from NVIDIA: <address type='pci' domain='0x0000' bus='22' slot='0x00' function='0x0'/>
				// where bus='22' is the index of the downstream port controller
				dev.Address.Bus = strconv.Itoa(downstreamPort.controllerIndex)
				dev.Address.Slot = "0x00"
				dev.Address.Function = "0x0"
				downstreamPort.deviceAssigned = true
				port.pcieSwitch.devicesAssigned++
				log.Log.V(1).Infof("Assigned device to PCIe switch downstream port index=%d (total devices on switch: %d/%d)",
					downstreamPort.controllerIndex, port.pcieSwitch.devicesAssigned, len(port.pcieSwitch.downstreamPorts))
				return
			}
		}
		log.Log.Warningf("No available downstream ports on PCIe switch, falling back to direct root port attachment")
	}

	// No switch or switch full: attach directly to root port
	// Use Bus attribute with the root port's index (NVIDIA docs pattern)
	// Example: <address type='pci' domain='0x0000' bus='17' slot='0x00' function='0x0'/>
	// where bus='17' is the index of the root port controller
	dev.Address.Bus = strconv.Itoa(port.controllerIndex)
	dev.Address.Slot = "0x00"
	dev.Address.Function = "0x0"
}

func isGraceHostDevicesEnabledForArch(vmi *v1.VirtualMachineInstance, cfg *util.GraceVirtualizationConfig) bool {
	if cfg == nil {
		return false
	}
	arch := strings.TrimSpace(vmi.Spec.Architecture)
	return arch == "" || strings.EqualFold(arch, arm64Architecture)
}

func applyGraceSMMUv3IOMMUTopology(domain *api.Domain, cfg *util.GraceVirtualizationConfig, graceHostDevicesEnabled bool, iommuFDDeviceAvailable bool) {
	if domain == nil {
		return
	}
	// Without the Grace annotation this planner must not own SMMUv3 cleanup; preserve
	// any domain XML or hook-injected QEMU args from other flows.
	if !graceHostDevicesEnabled || cfg == nil {
		return
	}

	// Cleanup any legacy raw qemu SMMUv3 args from previous implementation revisions.
	removeGraceSMMUv3QEMUArgs(domain)

	filtered := make([]api.IOMMU, 0, len(domain.Spec.Devices.IOMMUs))
	for i := range domain.Spec.Devices.IOMMUs {
		if strings.EqualFold(domain.Spec.Devices.IOMMUs[i].Model, "smmuv3") {
			continue
		}
		filtered = append(filtered, domain.Spec.Devices.IOMMUs[i])
	}
	domain.Spec.Devices.IOMMUs = filtered

	if !util.GraceFieldEnabled(cfg.SMMUv3) {
		return
	}

	pciBuses := collectNUMAPXBPciBuses(domain)
	cmdqvRequested := util.GraceFieldEnabled(cfg.VCMDQ)
	for _, pciBus := range pciBuses {
		accel := "off"
		if iommuFDDeviceAvailable {
			accel = "on"
		}
		driver := &api.IOMMUDriver{
			PCIBus: pciBus,
			Accel:  accel,
		}
		if accel == "on" {
			driver.ATS = "on"
			driver.RIL = "off"
			driver.PASID = "on"
			driver.OAS = "48"
			if cmdqvRequested {
				driver.CMDQV = "on"
			}
		} else {
			// Non-accelerated SMMUv3 must not advertise ATS/PASID/CMDQV/OAS.
			// ril='on' matches the libvirt/QEMU shape used without iommufd.
			driver.RIL = "on"
		}
		domain.Spec.Devices.IOMMUs = append(domain.Spec.Devices.IOMMUs, api.IOMMU{
			Model:  "smmuv3",
			Driver: driver,
		})
	}
}

func removeGraceSMMUv3QEMUArgs(domain *api.Domain) {
	if domain == nil || domain.Spec.QEMUCmd == nil || len(domain.Spec.QEMUCmd.QEMUArg) == 0 {
		return
	}

	filtered := make([]api.Arg, 0, len(domain.Spec.QEMUCmd.QEMUArg))
	for i := 0; i < len(domain.Spec.QEMUCmd.QEMUArg); i++ {
		arg := domain.Spec.QEMUCmd.QEMUArg[i]
		if arg.Value == "-device" && i+1 < len(domain.Spec.QEMUCmd.QEMUArg) &&
			strings.Contains(domain.Spec.QEMUCmd.QEMUArg[i+1].Value, "arm-smmuv3") {
			i++
			continue
		}
		if strings.Contains(arg.Value, "arm-smmuv3") {
			continue
		}
		filtered = append(filtered, arg)
	}

	if len(filtered) == 0 {
		domain.Spec.QEMUCmd = nil
		return
	}
	domain.Spec.QEMUCmd.QEMUArg = filtered
}

// applyEGMMemoryDevices discovers EGM character devices on the host and adds
// <memory model='egm' access='shared'> elements to the domain XML for each
// passthrough GPU that has an associated EGM region.
//
// On multi-socket systems (e.g. GB200) a single EGM char device (/dev/egmN)
// may serve multiple GPUs on the same socket. Each GPU still gets its own
// <memory model='egm'> element (QEMU needs per-GPU acpi-egm-memory objects),
// but the Grace EGM contract requires a VM to own the complete discovered EGM
// allocation boundary: every GPU associated with every host EGM char device.
// Per-GPU EGM size is therefore derived from the full host EGM group, not from
// the VM-local subset of selected GPUs. If a future allocator exposes a narrower
// compute-tray grouping explicitly, this check can be refined to that boundary.
//
// The EGM backing size comes from /sys/class/egm/egm*/egm_size. Do not derive
// this value from lsmem, MemTotal, or generic host RAM accounting: those report
// Linux memory state, while egm_size is the size QEMU maps through /dev/egmN.
//
// The function also assigns stable user-defined aliases to the GPU hostdevs
// so that the <memory> pciDev attribute can reference them.
func applyEGMMemoryDevices(domain *api.Domain, devicesWithNUMA []deviceNUMAInfo) error {
	const bytesPerMiB = 1024 * 1024

	egmDevices, err := discoverEGMDevicesFunc()
	if err != nil {
		return fmt.Errorf("egm requested but failed to discover host EGM devices: %w", err)
	}
	if len(egmDevices) == 0 {
		return fmt.Errorf("egm requested but no /sys/class/egm devices were found on the host")
	}

	gpuDevices := make([]deviceNUMAInfo, 0, len(devicesWithNUMA))
	for _, info := range devicesWithNUMA {
		if info.dev == nil || info.dev.Type != api.HostDevicePCI {
			continue
		}
		if hardware.FindEGMDeviceForGPU(info.bdf, egmDevices) == nil {
			continue
		}
		gpuDevices = append(gpuDevices, info)
	}

	if len(gpuDevices) == 0 {
		return fmt.Errorf("egm requested but no GPU host devices with EGM association were found")
	}

	slices.SortFunc(gpuDevices, func(a, b deviceNUMAInfo) int {
		if a.bdf < b.bdf {
			return -1
		}
		if a.bdf > b.bdf {
			return 1
		}
		return 0
	})

	selectedGPUsPerEGMPath := make(map[string]uint64)
	hostGPUsPerEGMPath := make(map[string]uint64)
	perGPUSizeMiBByPath := make(map[string]uint64)
	for _, gpuInfo := range gpuDevices {
		egm := hardware.FindEGMDeviceForGPU(gpuInfo.bdf, egmDevices)
		if egm == nil {
			return fmt.Errorf("egm requested for GPU %s but no matching host EGM device was found", gpuInfo.bdf)
		}
		selectedGPUsPerEGMPath[egm.DevPath]++
	}

	for _, egm := range egmDevices {
		selectedGPUs := selectedGPUsPerEGMPath[egm.DevPath]
		hostGPUs := uint64(len(egm.GPUBDFs))
		if hostGPUs == 0 {
			return fmt.Errorf("egm device %s has no associated GPUs in host sysfs", egm.DevPath)
		}
		if selectedGPUs != hostGPUs {
			return fmt.Errorf("egm requires all GPUs associated with discovered host EGM devices to be assigned to the VM; selected %d/%d GPUs sharing %s",
				selectedGPUs, hostGPUs, egm.DevPath)
		}
		if egm.EGMSizeBytes%hostGPUs != 0 {
			return fmt.Errorf("egm device %s reports %d bytes across %d GPUs; size must divide evenly per GPU",
				egm.DevPath, egm.EGMSizeBytes, hostGPUs)
		}

		perGPUSizeBytes := egm.EGMSizeBytes / hostGPUs
		if perGPUSizeBytes%bytesPerMiB != 0 {
			return fmt.Errorf("egm device %s reports %d bytes across %d GPUs; per-GPU size must align to MiB",
				egm.DevPath, egm.EGMSizeBytes, hostGPUs)
		}

		perGPUSizeMiB := perGPUSizeBytes / bytesPerMiB
		if perGPUSizeMiB == 0 {
			return fmt.Errorf("egm device %s has zero per-GPU EGM size (total %d bytes, %d GPUs)",
				egm.DevPath, egm.EGMSizeBytes, hostGPUs)
		}

		hostGPUsPerEGMPath[egm.DevPath] = hostGPUs
		perGPUSizeMiBByPath[egm.DevPath] = perGPUSizeMiB
	}

	hostdevIndex := 0
	for i := range gpuDevices {
		gpuInfo := &gpuDevices[i]
		hostdevAlias := fmt.Sprintf("%s%d", egmHostdevAliasPrefix, hostdevIndex)

		gpuInfo.dev.Alias = api.NewUserDefinedAlias(hostdevAlias)

		egm := hardware.FindEGMDeviceForGPU(gpuInfo.bdf, egmDevices)
		if egm == nil {
			return fmt.Errorf("egm requested for GPU %s but no matching host EGM device was found", gpuInfo.bdf)
		}
		if _, err := statEGMDevicePathFunc(egm.DevPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("egm device %s is not present in the virt-launcher pod; expose /dev/egmN to the pod before starting the VM", egm.DevPath)
			}
			return fmt.Errorf("egm device %s is not accessible in the virt-launcher pod: %w", egm.DevPath, err)
		}

		numGPUs := hostGPUsPerEGMPath[egm.DevPath]
		perGPUSizeMiB := perGPUSizeMiBByPath[egm.DevPath]
		if perGPUSizeMiB == 0 {
			return fmt.Errorf("egm device %s has no validated per-GPU size for GPU %s", egm.DevPath, gpuInfo.bdf)
		}

		domain.Spec.Devices.MemoryDevices = append(domain.Spec.Devices.MemoryDevices, api.MemoryDevice{
			Model:  "egm",
			Access: "shared",
			Source: &api.MemoryDeviceSource{Path: egm.DevPath},
			Target: &api.MemoryTarget{
				Size:   api.Memory{Value: perGPUSizeMiB, Unit: "MiB"},
				Node:   strconv.Itoa(gpuInfo.guestNUMANode),
				PCIDev: api.UserAliasPrefix + hostdevAlias,
			},
		})

		log.Log.Infof("EGM: added memory device %s (%d MiB, 1/%d host GPUs) for GPU %s (alias %s, guest NUMA %d)",
			egm.DevPath, perGPUSizeMiB, numGPUs, gpuInfo.bdf, hostdevAlias, gpuInfo.guestNUMANode)

		hostdevIndex++
	}

	return nil
}

func collectNUMAPXBPciBuses(domain *api.Domain) []string {
	type pxbRef struct {
		busNr int
		index int
	}
	pxbs := make([]pxbRef, 0)
	for i := range domain.Spec.Devices.Controllers {
		ctrl := domain.Spec.Devices.Controllers[i]
		if ctrl.Model != "pcie-expander-bus" || ctrl.Target == nil {
			continue
		}
		if strings.TrimSpace(ctrl.Target.BusNr) == "" {
			continue
		}
		busNr, err := parseBusNumber(ctrl.Target.BusNr)
		if err != nil {
			continue
		}
		index, err := strconv.Atoi(ctrl.Index)
		if err != nil || index <= 0 {
			continue
		}
		pxbs = append(pxbs, pxbRef{
			busNr: busNr,
			index: index,
		})
	}

	slices.SortFunc(pxbs, func(a, b pxbRef) int {
		if a.busNr == b.busNr {
			if a.index == b.index {
				return 0
			}
			if a.index < b.index {
				return -1
			}
			return 1
		}
		if a.busNr < b.busNr {
			return -1
		}
		return 1
	})

	buses := make([]string, 0, len(pxbs))
	for i := range pxbs {
		// Libvirt native smmuv3 wiring uses <iommu><driver pciBus='N'/></iommu>,
		// where N maps to the pcie-expander-bus controller index.
		buses = append(buses, strconv.Itoa(pxbs[i].index))
	}
	return buses
}

func applyGraceHostDeviceSettings(dev *api.HostDevice, guestNUMANode int, giNodeSet string, enableIOMMUFD bool) {
	if dev == nil || dev.Type != api.HostDevicePCI {
		return
	}

	if enableIOMMUFD {
		if dev.Driver == nil {
			dev.Driver = &api.HostDeviceDriver{}
		}
		dev.Driver.IOMMUFD = defaultHostDeviceIOMMUFD
	} else {
		dev.Driver = nil
	}

	nodeSet := strings.TrimSpace(giNodeSet)
	if nodeSet == "" {
		nodeSet = strconv.Itoa(guestNUMANode)
	}
	dev.ACPI = &api.HostDeviceACPI{
		NodeSet: nodeSet,
	}
}

func buildGraceGINodeSetAssignments(domain *api.Domain, devicesWithNUMA []deviceNUMAInfo, graceHostDevicesEnabled bool) map[*api.HostDevice]string {
	assignments := make(map[*api.HostDevice]string)
	if !graceHostDevicesEnabled || domain == nil {
		return assignments
	}

	giDevices := make([]deviceNUMAInfo, 0, len(devicesWithNUMA))
	for _, info := range devicesWithNUMA {
		if info.dev == nil || info.dev.Type != api.HostDevicePCI {
			continue
		}
		// Current Grace GPU passthrough guidance uses dedicated zero-memory GI
		// NUMA nodes per large-MMIO GPU. We treat dedicated PXB devices as GI
		// consumers so each GPU receives a unique guest ACPI nodeset.
		if !info.dedicatedPXB {
			continue
		}
		giDevices = append(giDevices, info)
	}
	if len(giDevices) == 0 {
		return assignments
	}

	slices.SortFunc(giDevices, func(a, b deviceNUMAInfo) int {
		if a.guestNUMANode != b.guestNUMANode {
			if a.guestNUMANode < b.guestNUMANode {
				return -1
			}
			return 1
		}
		if a.bdf == b.bdf {
			return 0
		}
		if a.bdf < b.bdf {
			return -1
		}
		return 1
	})

	nextNode := nextAvailableGuestNUMACellID(domain)
	if nextNode < graceGINodeSetBase {
		nextNode = graceGINodeSetBase
	}
	for _, info := range giDevices {
		start := nextNode
		end := start + graceGINodesPerGPU - 1
		ensureGuestNUMACellRange(domain, start, end)
		assignments[info.dev] = fmt.Sprintf("%d-%d", start, end)
		log.Log.V(1).Infof("Grace GI: assigned host device %s to dedicated guest NUMA GI nodes %d-%d", info.bdf, start, end)
		nextNode = end + 1
	}
	return assignments
}

func nextAvailableGuestNUMACellID(domain *api.Domain) int {
	if domain == nil || domain.Spec.CPU.NUMA == nil {
		return 0
	}
	maxID := -1
	for _, cell := range domain.Spec.CPU.NUMA.Cells {
		id, err := strconv.Atoi(strings.TrimSpace(cell.ID))
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func ensureGuestNUMACellRange(domain *api.Domain, start, end int) {
	if domain == nil || start > end {
		return
	}
	if domain.Spec.CPU.NUMA == nil {
		domain.Spec.CPU.NUMA = &api.NUMA{}
	}
	existing := make(map[int]struct{}, len(domain.Spec.CPU.NUMA.Cells))
	for _, cell := range domain.Spec.CPU.NUMA.Cells {
		id, err := strconv.Atoi(strings.TrimSpace(cell.ID))
		if err != nil {
			continue
		}
		existing[id] = struct{}{}
	}
	for id := start; id <= end; id++ {
		if _, ok := existing[id]; ok {
			continue
		}
		domain.Spec.CPU.NUMA.Cells = append(domain.Spec.CPU.NUMA.Cells, api.NUMACell{
			ID:     strconv.Itoa(id),
			Memory: 0,
			Unit:   "KiB",
		})
	}
	slices.SortFunc(domain.Spec.CPU.NUMA.Cells, func(a, b api.NUMACell) int {
		ai, errA := strconv.Atoi(strings.TrimSpace(a.ID))
		bi, errB := strconv.Atoi(strings.TrimSpace(b.ID))
		switch {
		case errA == nil && errB == nil:
			if ai == bi {
				return 0
			}
			if ai < bi {
				return -1
			}
			return 1
		case errA == nil:
			return -1
		case errB == nil:
			return 1
		default:
			if a.ID == b.ID {
				return 0
			}
			if a.ID < b.ID {
				return -1
			}
			return 1
		}
	})
}

func shouldUseDedicatedPXBForDevice(info deviceNUMAInfo, graceHostDevicesEnabled bool) bool {
	if !graceHostDevicesEnabled || info.dev == nil || info.dev.Type != api.HostDevicePCI {
		return false
	}
	totalMMIO, err := getDevicePCITotalMMIOSizeFunc(info.bdf)
	if err != nil {
		log.Log.V(1).Reason(err).Infof("unable to inspect PCI MMIO footprint for host device %s", info.bdf)
		return false
	}
	if totalMMIO < largeMMIOPXBIsolationThreshold {
		return false
	}
	log.Log.V(1).Infof("host device %s exposes %d GiB MMIO aperture; allocating dedicated PXB hierarchy",
		info.bdf, totalMMIO>>30)
	return true
}

func shouldForceIsolatedPXBsForGraceMixedTopology(graceSMMUv3Enabled bool, devices []deviceNUMAInfo) bool {
	if !graceSMMUv3Enabled {
		return false
	}
	var hasLargeMMIO bool
	var hasSmallMMIO bool
	for i := range devices {
		if devices[i].dedicatedPXB {
			hasLargeMMIO = true
		} else {
			hasSmallMMIO = true
		}
		if hasLargeMMIO && hasSmallMMIO {
			return true
		}
	}
	return false
}

func getDevicePCITotalMMIOSize(bdf string) (uint64, error) {
	resourcePath := filepath.Join("/sys/bus/pci/devices", bdf, "resource")
	f, err := os.Open(resourcePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var total uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		start, err := strconv.ParseUint(fields[0], 0, 64)
		if err != nil {
			return 0, err
		}
		end, err := strconv.ParseUint(fields[1], 0, 64)
		if err != nil {
			return 0, err
		}
		flags, err := strconv.ParseUint(fields[2], 0, 64)
		if err != nil {
			return 0, err
		}
		if flags&ioResourceMemFlag == 0 || (start == 0 && end == 0) || end < start {
			continue
		}
		total += end - start + 1
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return total, nil
}

func isIOMMUFDDeviceAvailable() bool {
	_, err := os.Stat("/dev/iommu")
	return err == nil
}

func resolveHostDevicePCIAddress(dev *api.HostDevice) (string, error) {
	if dev == nil || dev.Source.Address == nil {
		return "", fmt.Errorf("host device address missing")
	}

	addr := dev.Source.Address
	if addr.UUID != "" {
		return getMdevParentPCIAddressFunc(addr.UUID)
	}
	if addr.Domain != "" && addr.Bus != "" && addr.Slot != "" && addr.Function != "" {
		return formatPCIAddressFunc(addr)
	}
	return "", fmt.Errorf("unsupported host device address format")
}

// shouldCollapseHostDeviceNUMA returns true when PCI host devices should all be mapped to a single
// guest NUMA node. This happens when CPU NUMA passthrough is enabled (already validated by the caller),
// the converter produced only one guest NUMA cell, and every host device being considered already
// resides on the unified group (node 0). In all other cases we preserve the original host NUMA
// placement to avoid breaking locality expectations from VFIO / IOMMU mappings.
func shouldCollapseHostDeviceNUMA(vmi *v1.VirtualMachineInstance, domain *api.Domain, hostNUMANodes map[int]struct{}) bool {
	if vmi == nil || domain == nil {
		return false
	}

	if domain.Spec.CPU.NUMA == nil {
		return false
	}

	if len(domain.Spec.CPU.NUMA.Cells) > 1 {
		return false
	}

	if len(hostNUMANodes) != 1 {
		return false
	}

	_, onlyNodeZero := hostNUMANodes[unifiedNUMAGroup]
	return onlyNodeZero
}

func getGuestNUMANodes(domain *api.Domain) map[int]struct{} {
	result := make(map[int]struct{})
	if domain == nil || domain.Spec.CPU.NUMA == nil {
		return result
	}
	for _, cell := range domain.Spec.CPU.NUMA.Cells {
		if cell.ID == "" {
			continue
		}
		if id, err := strconv.Atoi(cell.ID); err == nil {
			result[id] = struct{}{}
		}
	}
	return result
}

func mapHostToGuestNUMANode(hostNode int, guestNodes map[int]struct{}, hostToGuest map[int]int) (int, bool) {
	if mapped, ok := hostToGuest[hostNode]; ok {
		return mapped, true
	}
	if len(guestNodes) == 0 {
		return hostNode, false
	}
	if _, ok := guestNodes[hostNode]; ok {
		return hostNode, true
	}
	// Fallback to the lowest guest NUMA node to keep the domain XML valid
	guestList := make([]int, 0, len(guestNodes))
	for node := range guestNodes {
		guestList = append(guestList, node)
	}
	slices.Sort(guestList)
	return guestList[0], false
}

func getHostToGuestNUMAMap(domain *api.Domain) map[int]int {
	result := make(map[int]int)
	if domain == nil || domain.Spec.NUMATune == nil {
		return result
	}
	for _, node := range domain.Spec.NUMATune.MemNodes {
		if strings.TrimSpace(node.NodeSet) == "" {
			continue
		}
		values, err := hardware.ParseCPUSetLine(node.NodeSet, 1024)
		if err != nil {
			log.Log.V(1).Reason(err).Infof("unable to parse NUMA memnode nodeset %q", node.NodeSet)
			continue
		}
		for _, host := range values {
			result[host] = int(node.CellID)
		}
	}
	return result
}

func (p *numaPCIPlanner) reserveRootSlotIfNeeded(addr *api.Address) {
	if !isRootBusAddress(addr) || addr == nil || strings.TrimSpace(addr.Slot) == "" {
		return
	}
	slot, err := parsePCISlot(addr.Slot)
	if err != nil {
		return
	}
	p.markBusSlot(rootBusControllerMarker, 0, slot)
	if slot >= defaultPXBSlot && slot <= maxPXBSlot && slot >= p.nextPXBSlot {
		p.nextPXBSlot = slot + 1
	}
}

func isRootBusAddress(addr *api.Address) bool {
	if addr == nil {
		return false
	}
	if !isZeroOrUnsetPCIField(addr.Domain) {
		return false
	}
	bus := strings.TrimSpace(addr.Bus)
	if bus == "" {
		return true
	}
	busVal, err := parseBusNumber(bus)
	if err != nil {
		return false
	}
	return busVal == 0
}

func parsePCISlot(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return -1, fmt.Errorf("empty slot value")
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		val, err := strconv.ParseInt(trimmed[2:], 16, 32)
		return int(val), err
	}
	if val, err := strconv.ParseInt(trimmed, 16, 32); err == nil {
		return int(val), nil
	}
	val, err := strconv.ParseInt(trimmed, 10, 32)
	return int(val), err
}

func isZeroOrUnsetPCIField(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		trimmed = trimmed[2:]
		if trimmed == "" {
			return true
		}
		val, err := strconv.ParseInt(trimmed, 16, 64)
		return err == nil && val == 0
	}
	if val, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return val == 0
	}
	if val, err := strconv.ParseInt(trimmed, 16, 64); err == nil {
		return val == 0
	}
	return false
}
