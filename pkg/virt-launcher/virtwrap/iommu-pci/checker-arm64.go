//go:build arm64

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

/*
#include <stdint.h>
uint64_t read_id_aa64mmfr0() {
    uint64_t id;
    // #nosec G103
    __asm__ volatile ("mrs %0, ID_AA64MMFR0_EL1" : "=r" (id));
    return id;
}
*/
import "C"

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"

	"kubevirt.io/client-go/log"
)

// ARM64 IOMMU and PCI Configuration Checker
//
// This file provides utilities to check IOMMU capabilities and configure PCI devices
// on ARM64 systems with NVIDIA GPU virtualization support. It is based on the work
// from: https://github.com/vladikr/iommu-pci-checker
//
// The code handles two device access interfaces:
//  1. Modern IOMMUFD interface (Linux 6.2+, preferred)
//  2. Legacy VFIO group-based interface (fallback for older kernels)
//
// Key capabilities checked:
//  - ATS (Address Translation Services): Allows devices to cache address translations
//  - PASID (Process Address Space ID): Enables multiple address spaces per device
//  - IOMMUFD: Modern kernel interface for IOMMU device management
//  - SMMUv3: ARM System Memory Management Unit version 3
//  - OAS (Output Address Size): CPU physical address width capability
//
// These checks are critical for:
//  - Determining if GPU passthrough will work correctly
//  - Calculating required PCI memory hole sizes for device BARs
//  - Configuring NUMA topology for optimal IOMMU performance
//  - Ensuring proper address translation and memory protection
//
// This functionality is ARM64-specific. See checker.go for stub implementations
// used on other architectures (x86_64, s390x).

const (
	// Legacy VFIO (Virtual Function I/O) ioctl constants
	// VFIO provides a framework for exposing direct device access to userspace
	VFIO_GROUP_GET_DEVICE_FD = 0x3B6A // Get file descriptor for a specific device
	VFIO_CHECK_EXTENSION     = 0x3B65 // Check if a VFIO extension is supported
	VFIO_SET_IOMMU           = 0x3B66 // Set the IOMMU type for the container
	VFIO_GROUP_SET_CONTAINER = 0x3B68 // Associate a group with a container
	VFIO_TYPE1v2_IOMMU       = 3      // Type1v2 IOMMU backend

	// Native IOMMUFD Constants (Mainline Linux kernel interface)
	// IOMMUFD is the modern replacement for legacy VFIO group-based device access
	VFIO_DEVICE_BIND_IOMMUFD = 0x3B76 // _IO(VFIO_TYPE, VFIO_BASE + 18)
)

// vfioDeviceBindIommufd16 is the standard 16-byte struct for binding a device to IOMMUFD
// This is the base struct used in mainline Linux kernels
type vfioDeviceBindIommufd16 struct {
	Argsz    uint32 // Size of this structure
	Flags    uint32 // Flags for the binding operation
	Iommufd  int32  // File descriptor of the IOMMU device
	OutDevid uint32 // Output: device ID assigned by IOMMUFD
}

// vfioDeviceBindIommufd24 is an extended 24-byte struct for binding with additional features
// This variant includes support for token-based authentication
type vfioDeviceBindIommufd24 struct {
	Argsz        uint32 // Size of this structure
	Flags        uint32 // Flags for the binding operation
	Iommufd      int32  // File descriptor of the IOMMU device
	OutDevid     uint32 // Output: device ID assigned by IOMMUFD
	TokenUuidPtr uint64 // Pointer to UUID token for secure binding
}

// bindNoArgsz32 is a variant without the argsz field, using 32-bit device ID
// Used for compatibility with certain kernel versions
type bindNoArgsz32 struct {
	Flags    uint32
	Iommufd  int32
	OutDevid uint32
	_        uint32 // padding if needed
}

// bindNoArgsz64 is a variant without the argsz field, using 64-bit device ID
// Used for compatibility with certain kernel versions
type bindNoArgsz64 struct {
	Flags    uint32
	Iommufd  int32
	OutDevid uint64
}

// parseConfigHybrid reads and parses the PCI extended configuration space for a device
// to determine IOMMU-related capabilities. It uses a hybrid strategy that tries
// modern IOMMUFD first, then falls back to legacy VFIO.
//
// Returns:
//   - atsSupported: whether Address Translation Services capability is present
//   - atsEnabled: whether ATS is currently enabled
//   - pasidSupported: whether Process Address Space ID capability is present
//   - ssidSize: the SubStream ID size for PASID
//   - err: any error encountered during reading or parsing
func parseConfigHybrid(bdf string) (iommufdEnabled, atsSupported, atsEnabled, pasidSupported bool, ssidSize int, err error) {
	native, data, err := getConfigHybridData(bdf)
	if err != nil {
		return false, false, false, false, 0, err
	}
	// Parse PCI Extended Capabilities from the configuration space
	ats, atsEn, pasid, ssid, err := parseCapabilities(data)
	return native, ats, atsEn, pasid, ssid, err
}

// getConfigHybridData retrieves the PCI extended configuration space data for a device
// using a two-tier strategy:
//  1. Try Native IOMMUFD (modern, preferred method for kernels 6.2+)
//  2. If Native fails/unavailable, fall back to Legacy VFIO (group-based access)
//
// This hybrid approach ensures compatibility across different kernel versions and
// IOMMU configurations.
func getConfigHybridData(bdf string) (bool, []byte, error) {
	var data []byte
	var err error

	// Strategy 1: Try Native IOMMUFD (Preferred for modern kernels)
	cdevPath := getCdevPath(bdf)
	if cdevPath != "" {
		log.Log.V(6).Infof("Attempting Native IOMMUFD (%s)...", cdevPath)
		data, err = readConfigNative(cdevPath)
		if err == nil {
			log.Log.V(6).Info("[Native] Success.")
			return true, data, nil
		}
	}
	log.Log.V(6).Infof("[Native] Failed: %v", err)

	// Strategy 2: Fallback to Legacy VFIO (for older kernels or unavailable IOMMUFD)
	log.Log.V(6).Info("Attempting Legacy VFIO Group...")
	data, err = readConfigLegacy(bdf)
	if err != nil {
		return false, nil, fmt.Errorf("both Native and Legacy methods failed: %v", err)
	}

	log.Log.V(6).Info("[Legacy] Success.")
	return false, data, nil
}

// parseCapabilities walks the PCI Extended Capability linked list to find
// ATS (Address Translation Services) and PASID (Process Address Space ID) capabilities.
//
// The PCI Extended Configuration Space starts at offset 0x100 and uses a linked list
// structure where each capability has a header containing the capability ID and
// pointer to the next capability.
//
// Capability IDs:
//   - 0x0f: ATS (Address Translation Services) - enables device-initiated address translation
//   - 0x1b: PASID (Process Address Space ID) - enables multiple address spaces per device
//
// Returns the presence and configuration of these IOMMU-related capabilities.
func parseCapabilities(data []byte) (ats, atsEn, pasid bool, ssid int, err error) {
	if len(data) < 0x1000 {
		return false, false, false, 0, fmt.Errorf("config space too small: %d bytes", len(data))
	}

	const (
		atsCAPId   = 0x0f
		pasidCAPId = 0x1b
	)
	// Extended capabilities start at offset 0x100
	offset := uint32(0x100)
	// Walk the capability linked list
	for offset != 0 && int(offset) < len(data) {
		// Each capability header is 4 bytes: [15:0] = Cap ID, [31:20] = Next pointer
		header := binary.LittleEndian.Uint32(data[offset : offset+4])
		capID := uint16(header & 0xffff)
		next := uint32((header >> 20) & 0xfff)

		if capID == atsCAPId { // ATS Capability
			ats = true
			// ATS Control register is at offset +6 from capability start
			// Bit 15 indicates if ATS is enabled
			ctrl := binary.LittleEndian.Uint16(data[offset+6 : offset+8])
			atsEn = (ctrl & 0x8000) != 0
		} else if capID == pasidCAPId { // PASID Capability
			pasid = true
			// PASID Capability register at offset +4 contains SSID size at bits [12:8]
			capReg := binary.LittleEndian.Uint16(data[offset+4 : offset+6])
			ssid = int((capReg >> 8) & 0x1f)
		}

		// Next pointer less than 0x100 indicates end of capability list
		if next < 0x100 {
			break
		}
		offset = next
	}
	if !pasid {
		log.Log.V(6).Info("PASID capability missing.")
	}
	return
}

// -------------------------------------------------------------------------
// NATIVE STRATEGY
// Uses the modern IOMMUFD interface (available in Linux kernel 6.2+)
// This provides direct device access without the legacy group-based model
// -------------------------------------------------------------------------

// readConfigNative reads PCI configuration space using the modern IOMMUFD interface.
// This is the preferred method for kernels 6.2+ as it provides more direct and
// efficient access to devices.
//
// Steps:
//  1. Open /dev/iommu to get an IOMMU file descriptor
//  2. Open the device's character device in /dev/vfio/devices/
//  3. Bind the device to the IOMMUFD using adaptive struct sizing
//  4. Read the extended configuration space (region 7)
func readConfigNative(cdevPath string) ([]byte, error) {
	// Open the IOMMU device
	iommuFd, err := unix.Open("/dev/iommu", unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/iommu failed: %v", err)
	}
	defer unix.Close(iommuFd)

	// Open the device's character device node
	devFd, err := unix.Open(cdevPath, unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open cdev failed: %v", err)
	}
	defer unix.Close(devFd)

	// Bind device to IOMMUFD with adaptive struct sizing for compatibility
	if err := adaptiveBind(devFd, iommuFd); err != nil {
		return nil, err
	}

	// Read the extended configuration space
	return readConfigFromFD(devFd)
}

// adaptiveBind attempts to bind a device to IOMMUFD using adaptive struct sizing.
// The VFIO_DEVICE_BIND_IOMMUFD ioctl struct size varies across kernel versions:
//   - Mainline kernels (6.2+): 16-byte struct (base)
//   - Extended kernels: 24-byte struct (with TokenUuidPtr for secure binding)
//
// This function tries the 16-byte struct first, and falls back to the 24-byte
// variant if the kernel returns EINVAL, E2BIG, or ENOTTY errors.
func adaptiveBind(devFd, iommuFd int) error {
	// Attempt 1: Try standard 16-byte struct (works on most recent mainline kernels)
	var args16 vfioDeviceBindIommufd16
	args16.Argsz = uint32(unsafe.Sizeof(args16)) // 16
	args16.Flags = 0
	args16.Iommufd = int32(iommuFd)
	args16.OutDevid = 0

	_, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(devFd),
		uintptr(VFIO_DEVICE_BIND_IOMMUFD),
		uintptr(unsafe.Pointer(&args16)),
	)
	if e1 == 0 {
		log.Log.V(6).Infof("[Bind] Success with mainline 16-byte (dev_id: %d)", args16.OutDevid)
		return nil
	}
	log.Log.V(6).Infof("[Bind] 16-byte failed (%v), trying 24-byte extension...", e1)

	// Attempt 2: Try extended 24-byte struct if the 16-byte version failed
	// This handles kernels that require the extended struct with TokenUuidPtr field
	if e1 == unix.EINVAL || e1 == unix.E2BIG || e1 == unix.ENOTTY {
		var args24 vfioDeviceBindIommufd24
		args24.Argsz = uint32(unsafe.Sizeof(args24)) // 24
		args24.Flags = 0
		args24.Iommufd = int32(iommuFd)
		args24.OutDevid = 0
		args24.TokenUuidPtr = 0 // No token needed for basic binding

		_, _, e2 := unix.Syscall(
			unix.SYS_IOCTL,
			uintptr(devFd),
			uintptr(VFIO_DEVICE_BIND_IOMMUFD),
			uintptr(unsafe.Pointer(&args24)),
		)
		if e2 == 0 {
			log.Log.V(6).Infof("[Bind] Success with mainline 24-byte extension (dev_id: %d).", args24.OutDevid)
			return nil
		}
		return fmt.Errorf("both 16-byte (%v) and 24-byte (%v) failed", e1, e2)
	}

	return fmt.Errorf("16-byte bind failed (no fallback): %v", e1)
}

// -------------------------------------------------------------------------
// LEGACY STRATEGY
// Uses the legacy VFIO group-based interface (pre-6.2 kernels)
// This method is more complex as it requires group and container management
// -------------------------------------------------------------------------

// readConfigLegacy reads PCI configuration space using the legacy VFIO group interface.
// This is the fallback method for kernels before 6.2 or when IOMMUFD is unavailable.
//
// The legacy VFIO model requires:
//  1. Finding the device's IOMMU group from sysfs
//  2. Opening the VFIO container (/dev/vfio/vfio)
//  3. Opening the group file descriptor (/dev/vfio/<group>)
//  4. Associating the group with the container
//  5. Setting the IOMMU type on the container
//  6. Getting the device file descriptor from the group
//  7. Reading the configuration space from the device FD
func readConfigLegacy(bdf string) ([]byte, error) {
	// Step 1: Determine which IOMMU group this device belongs to
	groupLink := fmt.Sprintf("/sys/bus/pci/devices/%s/iommu_group", bdf)
	target, err := os.Readlink(groupLink)
	if err != nil {
		return nil, fmt.Errorf("readlink iommu_group failed: %v", err)
	}
	group := filepath.Base(target)

	// Step 2: Open the VFIO container
	containerFd, err := unix.Open("/dev/vfio/vfio", unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/vfio/vfio failed: %v", err)
	}
	defer unix.Close(containerFd)

	// Step 3: Check if Type1v2 IOMMU extension is supported (best effort, ignore errors)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(containerFd), VFIO_CHECK_EXTENSION, uintptr(VFIO_TYPE1v2_IOMMU))
	if errno != 0 {
		return nil, fmt.Errorf("CHECK_EXTENSION failed: %v", errno)
	}

	// Step 4: Open the IOMMU group file descriptor
	groupPath := filepath.Join("/dev/vfio", group)
	groupFd, err := unix.Open(groupPath, unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open group %s failed: %v", groupPath, err)
	}
	defer unix.Close(groupFd)

	// Step 5: Associate the group with the container
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, uintptr(groupFd), VFIO_GROUP_SET_CONTAINER, uintptr(unsafe.Pointer(&containerFd)))
	if errno != 0 {
		return nil, fmt.Errorf("SET_CONTAINER failed: %v", errno)
	}

	// Step 6: Set the IOMMU type to Type1v2
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, uintptr(containerFd), VFIO_SET_IOMMU, uintptr(VFIO_TYPE1v2_IOMMU))
	if errno != 0 {
		return nil, fmt.Errorf("SET_IOMMU failed: %v", errno)
	}

	// Step 7: Get the device file descriptor from the group
	devFd, err := ioctlGetDeviceFd(groupFd, bdf)
	if err != nil {
		return nil, fmt.Errorf("GET_DEVICE_FD failed: %v", err)
	}
	defer unix.Close(devFd)

	// Step 8: Read the configuration space
	return readConfigFromFD(devFd)
}

// readConfigFromFD reads the PCI extended configuration space from a device file descriptor.
// It uses pread with offset (7<<40) to access VFIO region 7, which contains the
// extended configuration space (offsets 0x100-0xFFF).
func readConfigFromFD(devFd int) ([]byte, error) {
	data := make([]byte, 4096)
	// Region 7 contains the extended PCI configuration space
	n, err := unix.Pread(devFd, data, 7<<40)
	if err != nil {
		return nil, fmt.Errorf("pread failed: %v", err)
	}
	return data[:n], nil
}

// getCdevPath finds the character device path for a PCI device in the modern IOMMUFD interface.
// It searches in /sys/bus/pci/devices/<bdf>/vfio-dev/ for entries starting with "vfio"
// and returns the corresponding path in /dev/vfio/devices/.
func getCdevPath(bdf string) string {
	vfioDevDir := fmt.Sprintf("/sys/bus/pci/devices/%s/vfio-dev", bdf)
	entries, err := os.ReadDir(vfioDevDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "vfio") {
			return filepath.Join("/dev/vfio/devices", entry.Name())
		}
	}
	return ""
}

// ioctlGetDeviceFd gets a device file descriptor from a VFIO group using the legacy interface.
// It performs the VFIO_GROUP_GET_DEVICE_FD ioctl to retrieve a FD for the specified device.
func ioctlGetDeviceFd(groupFd int, deviceName string) (int, error) {
	cStr, err := unix.BytePtrFromString(deviceName)
	if err != nil {
		return 0, err
	}
	fd, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(groupFd), uintptr(VFIO_GROUP_GET_DEVICE_FD), uintptr(unsafe.Pointer(cStr)))
	if errno != 0 {
		return 0, errno
	}
	return int(fd), nil
}

// isSMMUv3Enabled checks if ARM SMMUv3 (System Memory Management Unit v3) is enabled.
// It searches /sys/class/iommu for symlinks containing "arm-smmu-v3" to determine
// if the ARM IOMMU is available on this system.
func isSMMUv3Enabled() (bool, error) {
	dir := "/sys/class/iommu"
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, f := range files {
		link := filepath.Join(dir, f.Name())
		target, err := os.Readlink(link)
		if err == nil && strings.Contains(target, "arm-smmu-v3") {
			return true, nil
		}
	}
	return false, nil
}

// getOASFromCPU determines the Output Address Size (OAS) supported by the ARM64 CPU.
// It reads the ID_AA64MMFR0_EL1 system register to extract the PARange field,
// which indicates the physical address width.
//
// PARange values map to OAS as follows:
//
//	0: 32 bits, 1: 36 bits, 2: 40 bits, 3: 44 bits,
//	4: 48 bits, 5: 52 bits, 6: 56 bits, 7: 64 bits
func getOASFromCPU() (int, error) {
	id := readIDAA64MMFR0()
	// PARange is in bits [3:0] of ID_AA64MMFR0_EL1
	parange := id & 0xf
	oasMap := map[uint64]int{0: 32, 1: 36, 2: 40, 3: 44, 4: 48, 5: 52, 6: 56, 7: 64}
	if oas, ok := oasMap[parange]; ok {
		// Sanity check: values below 40 bits are unlikely on modern systems
		if oas < 40 {
			return 0, fmt.Errorf("unlikely low PARange value: %d (oas=%d)", parange, oas)
		}
		return oas, nil
	}
	return 0, fmt.Errorf("unknown PARange value: %d", parange)
}

// readIDAA64MMFR0 reads the ARM64 ID_AA64MMFR0_EL1 system register using inline assembly.
// This register provides information about the memory model and address translation features.
func readIDAA64MMFR0() uint64 { return uint64(C.read_id_aa64mmfr0()) }

// CalculatePCIHole64Size computes the total size of 64-bit PCI memory regions for a device.
// It reads the /sys/bus/pci/devices/<bdf>/resource file which lists all BARs (Base Address Registers)
// and their properties.
//
// Each line in the resource file contains:
//   - start address (hex)
//   - end address (hex)
//   - flags (hex)
//
// This function only counts 64-bit prefetchable memory regions:
//   - flags & 0x200 must be set (64-bit region)
//   - flags & 0xf must equal 0xc (prefetchable memory)
//
// Returns the total size in bytes.
func CalculatePCIHole64Size(bdf string) (uint64, error) {
	var totalSize uint64
	resourcePath := fmt.Sprintf("/sys/bus/pci/devices/%s/resource", bdf)
	file, err := os.Open(resourcePath)
	if err != nil {
		return 0, fmt.Errorf("open resource failed: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			continue
		}
		// Remove "0x" prefix for parsing
		fields[0] = strings.TrimPrefix(fields[0], "0x")
		fields[1] = strings.TrimPrefix(fields[1], "0x")
		fields[2] = strings.TrimPrefix(fields[2], "0x")

		start, _ := strconv.ParseUint(fields[0], 16, 64)
		end, _ := strconv.ParseUint(fields[1], 16, 64)
		flags, _ := strconv.ParseUint(fields[2], 16, 64)

		// Skip invalid or empty regions
		if start == 0 || end == 0 || start > end {
			continue
		}

		// Only include 64-bit prefetchable memory regions
		// 0x200: 64-bit region flag
		// 0xc: prefetchable memory type
		if (flags&0x00000200) != 0x00000200 || (flags&0xf) != 0xc {
			continue
		}

		totalSize += end - start + 1
	}
	return totalSize, nil
}

// CalculateTotalPCIHole64Size computes the total PCI hole size with margin and rounds up to
// the next power of 2. This ensures proper alignment for memory mapping.
//
// Parameters:
//   - bdfComputedSize: total size in bytes from all devices
//   - marginKiB: safety margin in KiB to add
//
// Returns the final hole size in KiB, rounded up to the nearest power of 2.
func CalculateTotalPCIHole64Size(bdfComputedSize uint64, marginKiB uint64) uint64 {
	// Convert bytes to KiB and add margin
	totalSizeKiB := bdfComputedSize / 1024
	totalSizeKiB += marginKiB
	mod := bdfComputedSize % 1024

	if totalSizeKiB == 0 && mod == 0 {
		return 0
	}

	// Round up to next power of 2 for proper memory alignment
	holeSize := uint64(1)
	for holeSize < totalSizeKiB {
		holeSize <<= 1
	}

	// Round up if there were a missing margin
	if holeSize == totalSizeKiB && mod > 0 {
		holeSize <<= 1
	}

	return holeSize
}

// InferExtraNUMANodes determines the NUMA node configuration needed for proper IOMMU setup.
// For devices with SMMUv3 and PASID support, additional virtual NUMA nodes may be required
// to satisfy memory topology constraints.
//
// The function handles two scenarios:
//  1. Main node has CPUs: Return all memory-less nodes as extra nodes
//  2. Main node has no CPUs: Find consecutive memory-less nodes starting from mainNode+1
//
// This is important because IOMMU page tables need to be allocated on specific NUMA nodes
// to ensure proper memory affinity and performance.
//
// Returns:
//   - extraNodes: list of additional NUMA node IDs needed
//   - mainNode: the primary NUMA node for the device
//   - err: any error encountered reading NUMA information
func InferExtraNUMANodes(bdf string) (extraNodes []int, mainNode int, err error) {
	// Read the device's primary NUMA node from sysfs
	numaPath := fmt.Sprintf("/sys/bus/pci/devices/%s/numa_node", bdf)
	numaData, err := os.ReadFile(numaPath)
	if err != nil {
		return nil, -1, err
	}
	mainNode, err = strconv.Atoi(strings.TrimSpace(string(numaData)))
	if err != nil {
		return nil, -1, err
	}
	if mainNode < 0 {
		return nil, mainNode, fmt.Errorf("invalid main NUMA node: %d", mainNode)
	}

	// Get system NUMA topology information
	onlineNodes, _ := getOnlineNodes()
	allMemoryLess, _ := getMemoryLessNoCPUNodes(onlineNodes)
	hasCPUs, _ := nodeHasCPUs(mainNode)

	if hasCPUs {
		// Scenario 1: Main node has CPUs - use all memory-less nodes as extras
		sort.Ints(allMemoryLess)
		return allMemoryLess, mainNode, nil
	} else {
		// Scenario 2: Main node has no CPUs - find consecutive memory-less nodes
		extraNodes = []int{}
		candidate := mainNode + 1
		maxExtra := 16 // Limit to prevent excessive node allocation

		for len(extraNodes) < maxExtra {
			// Stop if we've reached a non-existent node
			if !contains(onlineNodes, candidate) {
				break
			}
			// Stop if we've reached a node with memory or CPUs
			isMemoryLess, _ := isMemoryLessNoCPU(candidate)
			if !isMemoryLess {
				break
			}
			extraNodes = append(extraNodes, candidate)
			candidate++
		}
		return extraNodes, mainNode, nil
	}
}

// getOnlineNodes parses /sys/devices/system/node/online to get a list of online NUMA nodes.
// The file format can be:
//   - Single nodes: "0,2,4"
//   - Ranges: "0-3,8-11"
//   - Mixed: "0-3,5,8-11"
//
// Returns a sorted list of all online node IDs.
func getOnlineNodes() ([]int, error) {
	data, err := os.ReadFile("/sys/devices/system/node/online")
	if err != nil {
		return nil, err
	}
	var nodes []int
	// Parse comma-separated list of nodes and ranges
	for _, part := range strings.Split(strings.TrimSpace(string(data)), ",") {
		if strings.Contains(part, "-") {
			// Handle range format (e.g., "0-3")
			rangeParts := strings.Split(part, "-")
			start, _ := strconv.Atoi(rangeParts[0])
			end, _ := strconv.Atoi(rangeParts[1])
			for i := start; i <= end; i++ {
				nodes = append(nodes, i)
			}
		} else {
			// Handle single node (e.g., "5")
			node, _ := strconv.Atoi(part)
			nodes = append(nodes, node)
		}
	}
	sort.Ints(nodes)
	return nodes, nil
}

// getMemoryLessNoCPUNodes filters the given list of nodes to return only those that have
// no memory and no CPUs. These are special NUMA nodes used for devices or other purposes.
func getMemoryLessNoCPUNodes(nodes []int) ([]int, error) {
	var memoryLess []int
	for _, node := range nodes {
		is, _ := isMemoryLessNoCPU(node)
		if is {
			memoryLess = append(memoryLess, node)
		}
	}
	return memoryLess, nil
}

// isMemoryLessNoCPU checks if a NUMA node has zero memory and no CPUs assigned.
// Such nodes are typically used for I/O devices or as placeholder nodes in the topology.
//
// Returns true if the node has MemTotal=0 and an empty/none cpulist.
func isMemoryLessNoCPU(node int) (bool, error) {
	// Check if node has zero memory
	memData, err := os.ReadFile(fmt.Sprintf("/sys/devices/system/node/node%d/meminfo", node))
	if err != nil {
		return false, err
	}
	memTotalRe := regexp.MustCompile(`MemTotal:\s+(\d+)\s+kB`)
	match := memTotalRe.FindStringSubmatch(string(memData))
	if len(match) != 2 || match[1] != "0" {
		return false, nil
	}

	// Check if node has no CPUs
	cpuData, err := os.ReadFile(fmt.Sprintf("/sys/devices/system/node/node%d/cpulist", node))
	if err != nil {
		return false, err
	}
	cpuStr := strings.TrimSpace(string(cpuData))
	if cpuStr != "" && cpuStr != "none" {
		return false, nil
	}

	return true, nil
}

// nodeHasCPUs checks if a NUMA node has any CPUs assigned to it.
// Returns true if the cpulist is non-empty and not "none".
func nodeHasCPUs(node int) (bool, error) {
	cpuData, err := os.ReadFile(fmt.Sprintf("/sys/devices/system/node/node%d/cpulist", node))
	if err != nil {
		return false, err
	}
	cpuStr := strings.TrimSpace(string(cpuData))
	return cpuStr != "" && cpuStr != "none", nil
}

// contains checks if a slice contains a specific integer value.
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
