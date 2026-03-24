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

package virtwrap

// IOMMUFD FD Receiving: virt-launcher side
//
// This file implements the virt-launcher (unprivileged) side of IOMMUFD file
// descriptor passing. The privileged virt-handler device plugin pre-opens and
// configures /dev/iommu, then passes the FD over a Unix domain socket using
// SCM_RIGHTS.
//
// virt-launcher connects to a fixed socket path, receives the FD, and will
// eventually pass it to libvirt via virDomainFDAssociate(domain, "iommu", 1, &fd, 0).
//
// The fixed socket path (/var/run/kubevirt/iommufd.sock) is bind-mounted by
// kubelet from the host-side socket created by the device plugin. virt-launcher
// simply checks if this path exists — no annotation or environment variable is needed.
//
// This pattern is designed to be reusable: external device plugins (e.g., NVIDIA
// kubevirt-gpu-device-plugin) can create their own socket at the same fixed path
// and virt-launcher will receive the FD identically.

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/sys/unix"

	"kubevirt.io/client-go/log"
)

const (
	// IOMMUFDSocketPath is the fixed path inside the virt-launcher pod where
	// the IOMMUFD socket is expected. This is bind-mounted by kubelet from the
	// host socket created by the device plugin during Allocate().
	IOMMUFDSocketPath = "/var/run/kubevirt/iommufd.sock"
)

// ReceiveIOMMUFD connects to the IOMMUFD Unix domain socket and receives the
// pre-configured IOMMUFD file descriptor via SCM_RIGHTS.
//
// The device plugin (virt-handler side) opens /dev/iommu, configures
// IOMMU_OPTION_RLIMIT_MODE, and sends the FD over this socket. This function
// is the receiving end.
//
// The returned FD is ready for use with libvirt's virDomainFDAssociate.
// The caller is responsible for eventually closing the FD.
//
// Parameters:
//   - socketPath: path to the Unix domain socket (typically IOMMUFDSocketPath)
//
// Returns:
//   - int: the received IOMMUFD file descriptor
//   - error: any error during connection or FD reception
func ReceiveIOMMUFD(socketPath string) (int, error) {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		return -1, fmt.Errorf("failed to connect to IOMMUFD socket %s: %w", socketPath, err)
	}
	defer conn.Close()

	log.Log.Infof("Connected to IOMMUFD socket, sleeping 5 seconds before read")
	time.Sleep(5 * time.Second)
	// Receive the FD via SCM_RIGHTS
	// The payload is a single byte; the FD is in the ancillary (out-of-band) data
	buf := make([]byte, 32)
	// Allocate generous space for SCM_RIGHTS control message
	// Need enough space for cmsghdr + fd array + alignment padding
	oob := make([]byte, 256)

	log.Log.Infof("About to ReadMsgUnix from IOMMUFD socket, oob buffer size: %d", len(oob))
	n, oobn, flags, _, err := conn.ReadMsgUnix(buf, oob)
	// 1. Log the exact state
	fmt.Printf("Read: n=%d bytes, oobn=%d, flags=%d, data=%v\n", n, oobn, flags, buf[:n])
	if err != nil {
		return -1, fmt.Errorf("failed to read from IOMMUFD socket: %w", err)
	}
	// 2. Check for truncation
	if flags&unix.MSG_CTRUNC != 0 {
		return -1, fmt.Errorf("OOB data was truncated by the kernel")
	}

	log.Log.Infof("ReadMsgUnix result: n=%d bytes, oobn=%d bytes, flags=%d, err=%v", n, oobn, flags, err)
	if err != nil {
		return -1, fmt.Errorf("failed to receive IOMMUFD FD from %s: %w", socketPath, err)
	}

	if oobn == 0 {
		return -1, fmt.Errorf("no oobn received from IOMMUFD socket (read %d regular bytes, flags=%d)", n, flags)
	}
	// Parse the ancillary data to extract the file descriptor
	scms, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, fmt.Errorf("failed to parse socket control message: %w", err)
	}

	if len(scms) == 0 {
		return -1, fmt.Errorf("no scms received from IOMMUFD socket")
	}

	fds, err := unix.ParseUnixRights(&scms[0])
	if err != nil {
		return -1, fmt.Errorf("failed to parse unix rights: %w", err)
	}

	if len(fds) == 0 {
		return -1, fmt.Errorf("no file descriptors received from IOMMUFD socket")
	}

	log.Log.V(3).Infof("Received IOMMUFD file descriptor %d from %s", fds[0], socketPath)
	return fds[0], nil
}
