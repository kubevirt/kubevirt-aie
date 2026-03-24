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

package device_manager

// IOMMUFD FD Passing: Privileged Device Plugin → Unprivileged virt-launcher
//
// This file implements the host-side (virt-handler) half of IOMMUFD file descriptor
// passing for GPU/PCI device passthrough on systems with IOMMUFD support.
//
// Background:
// On modern kernels (6.2+), IOMMUFD replaces the legacy VFIO container model.
// libvirt requires an IOMMUFD file descriptor that has been pre-configured with
// IOMMU_OPTION_RLIMIT_MODE to enable process-level resource accounting. In KubeVirt,
// the virt-launcher pod runs unprivileged and cannot open /dev/iommu itself.
// The privileged virt-handler device plugin opens and configures the FD, then passes
// it to virt-launcher via SCM_RIGHTS over a Unix domain socket.
//
// Flow:
// 1. Device plugin Allocate() detects IOMMUFD support
// 2. openAndConfigureIOMMUFD() opens /dev/iommu + sets RLIMIT_MODE (replicating
//    libvirt's virIOMMUFDOpenDevice + virIOMMUFDSetRLimitMode)
// 3. createIOMMUFDSocket() creates a one-shot Unix socket that transfers the FD
// 4. The socket file is bind-mounted into the virt-launcher pod at a fixed path
// 5. virt-launcher connects, receives the FD via SCM_RIGHTS, and later passes it
//    to libvirt via virDomainFDAssociate
//
// This pattern is designed to be reusable by external device plugins (e.g., the
// NVIDIA kubevirt-gpu-device-plugin) that need to pass pre-configured FDs to
// virt-launcher pods. The socket protocol is simple: one connection, one FD,
// one byte of payload, then cleanup.

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/safepath"
	"kubevirt.io/kubevirt/pkg/virt-handler/selinux"
)

const (
	// iommuFDSocketDir is the host side directory for IOMMUFD sockets.
	// This directory is under virt-lib-dir which is already mounted by the
	// virt-handler DaemonSet. Each socket file will have a  unique name per allocation.
	iommuFDSocketDir = "/var/run/kubevirt/fd-sockets"

	// IOMMUFDContainerSocketPath is the fixed path inside virt-launcher pods
	// where the IOMMUFD socket is mounted. virt-launcher checks for this path
	// at startup — no annotation is needed.
	IOMMUFDContainerSocketPath = "/var/run/kubevirt/iommufd.sock"

	// IOMMU_OPTION is the ioctl number for the IOMMUFD OPTION command.
	// Defined in Linux uAPI as _IO(IOMMUFD_TYPE, IOMMUFD_CMD_OPTION)
	// where IOMMUFD_TYPE = ';' (0x3B) and IOMMUFD_CMD_OPTION = 0x87.
	//
	//nolint:stylecheck,revive
	IOMMU_OPTION = 0x3B87
)

// iommuOption mirrors the kernel's struct iommu_option from
// include/uapi/linux/iommufd.h. This is the same layout that libvirt uses
// in virIOMMUFDSetRLimitMode.
type iommuOption struct {
	Size     uint32
	OptionID uint32
	Op       uint16
	Reserved uint16
	ObjectID uint32
	Val64    uint64
}

// openAndConfigureIOMMUFD opens /dev/iommu and configures it for use with
// libvirt by enabling IOMMU_OPTION_RLIMIT_MODE. This replicates the behavior
// of libvirt's virIOMMUFDOpenDevice + virIOMMUFDSetRLimitMode.
//
// IOMMU_OPTION_RLIMIT_MODE enables process-level resource accounting for
// IOMMU operations, which is required for proper memory pinning limits
// when doing device passthrough.
//
// The returned file descriptor must be passed to virt-launcher and eventually
// to libvirt via virDomainFDAssociate. The caller is responsible for closing
// the FD if createIOMMUFDSocket is not called.
func openAndConfigureIOMMUFD() (int, error) {
	fd, err := unix.Open("/dev/iommu", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to open /dev/iommu: %w", err)
	}

	// Set IOMMU_OPTION_RLIMIT_MODE = true
	// This enables per-process RLIMIT_MEMLOCK accounting for IOMMU mappings,
	// matching what libvirt expects when managing IOMMUFD-backed devices.
	option := iommuOption{
		Size:     uint32(unsafe.Sizeof(iommuOption{})), // size as used by libvirt
		OptionID: 0,                                    // IOMMU_OPTION_RLIMIT_MODE
		Op:       0,                                    // IOMMU_OPTION_OP_SET
		Val64:    1,                                    // true — enable rlimit mode
	}

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(IOMMU_OPTION),
		uintptr(unsafe.Pointer(&option)),
	)
	if errno != 0 {
		unix.Close(fd)
		return -1, fmt.Errorf("IOMMU_OPTION ioctl failed: %v", errno)
	}

	log.DefaultLogger().V(3).Infof("Opened and configured IOMMUFD (fd=%d, rlimit_mode=true)", fd)
	return fd, nil
}

// createIOMMUFDSocket creates a one-shot Unix domain socket that will transfer
// the IOMMUFD file descriptor to a connecting virt-launcher pod via SCM_RIGHTS.
//
// The socket file is created at /var/lib/kubevirt/fd-sockets/iommufd-{uniqueID}.sock
// and is intended to be bind-mounted into the virt-launcher container at the fixed
// path /var/run/kubevirt/iommufd.sock.
//
// A background goroutine is started that:
//   - Accepts exactly one connection
//   - Sends the IOMMUFD FD via SCM_RIGHTS (unix.UnixRights + WriteMsgUnix)
//   - Closes the FD and removes the socket file
//
// This function returns immediately after creating the listener. The goroutine
// handles the actual FD transfer asynchronously.
//
// Parameters:
//   - iommuFD: the configured IOMMUFD file descriptor from openAndConfigureIOMMUFD
//   - uniqueID: a unique identifier (UUID) for the socket filename
//
// Returns:
//   - hostSocketPath: the full path to the socket file on the host
//   - err: any error during socket creation
func createIOMMUFDSocket(iommuFD int, uniqueID string) (string, error) {
	if err := os.MkdirAll(iommuFDSocketDir, 0766); err != nil {
		return "", fmt.Errorf("failed to create socket directory %s: %w", iommuFDSocketDir, err)
	}

	if se, exists, err := selinux.NewSELinux(); err == nil && exists {
		safeIommuSocketDir, err := safepath.NewPathNoFollow(iommuFDSocketDir)
		if err != nil {
			return "", err
		}
		if err := selinux.RelabelFilesUnprivileged(se.IsPermissive(), safeIommuSocketDir); err != nil {
			return "", fmt.Errorf("failed to relabel iommu fd socket directory %s: %w", iommuFDSocketDir, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to detect SELinux: %w", err)
	}

	hostSocketPath := filepath.Join(iommuFDSocketDir, fmt.Sprintf("iommufd-%s.sock", uniqueID))

	// Remove stale socket file if it exists
	os.Remove(hostSocketPath)

	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: hostSocketPath, Net: "unix"})
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", hostSocketPath, err)
	}

	// Allow virt-launcher to connect
	if err := os.Chmod(hostSocketPath, 0666); err != nil {
		listener.Close()
		os.Remove(hostSocketPath)
		return "", fmt.Errorf("failed to chmod socket %s: %w", hostSocketPath, err)
	}

	if se, exists, err := selinux.NewSELinux(); err == nil && exists {
		safeHostSocketPath, err := safepath.NewPathNoFollow(hostSocketPath)
		if err != nil {
			return "", err
		}
		if err := selinux.RelabelFilesUnprivileged(se.IsPermissive(), safeHostSocketPath); err != nil {
			listener.Close()
			return "", fmt.Errorf("failed to relabel iommu fd socket %s: %w", hostSocketPath, err)
		}
	}

	log.DefaultLogger().V(3).Infof("IOMMUFD socket created at %s, waiting for virt-launcher connection", hostSocketPath)

	// One-shot goroutine: accept one connection, send FD, clean up
	go func() {
		defer listener.Close()
		defer os.Remove(hostSocketPath)
		defer unix.Close(iommuFD)

		conn, err := listener.AcceptUnix()
		if err != nil {
			log.DefaultLogger().Errorf("IOMMUFD socket accept failed: %v", err)
			return
		}
		log.DefaultLogger().V(3).Info("Accepted")
		defer conn.Close()

		// Send the IOMMUFD FD via SCM_RIGHTS
		rights := unix.UnixRights(iommuFD)
		log.DefaultLogger().V(3).Infof("IOMMUFD rights: %s", rights)
		wb, woob, err := conn.WriteMsgUnix([]byte{0}, rights, nil)
		if err != nil {
			log.DefaultLogger().Errorf("IOMMUFD socket WriteMsgUnix failed: %v", err)
			return
		}
		log.DefaultLogger().V(3).Infof("IOMMUFD socket WriteMsgUnix succeeded: written b: %d, written oob: %d", wb, woob)
		time.Sleep(500 * time.Second)
		log.DefaultLogger().V(3).Infof("IOMMUFD FD sent to virt-launcher via %s", hostSocketPath)
	}()

	return hostSocketPath, nil
}
