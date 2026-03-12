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

package admitters

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/util"
)

func filterNUMAHugepagesRequirementForGraceEGM(causes []metav1.StatusCause, field *k8sfield.Path, annotations map[string]string) []metav1.StatusCause {
	if len(annotations) == 0 {
		return causes
	}

	rawConfig, exists := annotations[v1.GraceVirtualizationAnnotation]
	if !exists || strings.TrimSpace(rawConfig) == "" {
		return causes
	}

	cfg, err := util.ParseGraceVirtualizationConfigStrict(rawConfig)
	if err != nil || !util.GraceFieldEnabled(cfg.EGM) {
		return causes
	}

	hugepagesField := field.Child("domain", "memory", "hugepages").String()
	numaField := field.Child("domain", "cpu", "numa", "guestMappingPassthrough").String()
	filtered := causes[:0]
	for _, cause := range causes {
		if cause.Field == numaField &&
			strings.Contains(cause.Message, hugepagesField) &&
			strings.Contains(cause.Message, "NUMA topology strategy") {
			continue
		}
		filtered = append(filtered, cause)
	}
	return filtered
}
