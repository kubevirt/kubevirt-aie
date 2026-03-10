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

package util

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	v1 "kubevirt.io/api/core/v1"
)

// GraceVirtualizationConfig is the canonical representation of the
// alpha.kubevirt.io/graceVirtualization annotation JSON payload.
type GraceVirtualizationConfig struct {
	SMMUv3 *bool `json:"smmuv3,omitempty"`
	VCMDQ  *bool `json:"vcmdq,omitempty"`
	EGM    *bool `json:"egm,omitempty"`
}

// ParseGraceVirtualizationConfigStrict parses raw JSON with unknown-field
// rejection and trailing-content detection. Use this in admission webhooks
// where malformed input must be surfaced to the user.
func ParseGraceVirtualizationConfigStrict(raw string) (*GraceVirtualizationConfig, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()

	cfg := &GraceVirtualizationConfig{}
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("invalid trailing content")
	}

	return cfg, nil
}

// GetGraceVirtualizationConfig extracts the config from a VMI's annotations.
// Returns nil when the annotation is absent, empty, or unparseable. Runtime
// code should use this lenient variant so malformed input does not crash
// virt-launcher.
func GetGraceVirtualizationConfig(vmi *v1.VirtualMachineInstance) *GraceVirtualizationConfig {
	if vmi == nil || len(vmi.Annotations) == 0 {
		return nil
	}
	raw, exists := vmi.Annotations[v1.GraceVirtualizationAnnotation]
	if !exists || strings.TrimSpace(raw) == "" {
		return nil
	}
	cfg := &GraceVirtualizationConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil
	}
	return cfg
}

// GraceFieldEnabled returns true only when the pointer is non-nil and true.
func GraceFieldEnabled(value *bool) bool {
	return value != nil && *value
}
