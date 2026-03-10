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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("Grace virtualization annotation config", func() {
	It("strictly parses explicit feature values", func() {
		cfg, err := ParseGraceVirtualizationConfigStrict(`{"smmuv3":true,"vcmdq":true,"egm":false}`)

		Expect(err).ToNot(HaveOccurred())
		Expect(GraceFieldEnabled(cfg.SMMUv3)).To(BeTrue())
		Expect(GraceFieldEnabled(cfg.VCMDQ)).To(BeTrue())
		Expect(GraceFieldEnabled(cfg.EGM)).To(BeFalse())
	})

	It("strictly parses the baseline opt-in payload", func() {
		cfg, err := ParseGraceVirtualizationConfigStrict(`{}`)

		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.SMMUv3).To(BeNil())
		Expect(cfg.VCMDQ).To(BeNil())
		Expect(cfg.EGM).To(BeNil())
	})

	It("rejects unknown fields in strict mode", func() {
		_, err := ParseGraceVirtualizationConfigStrict(`{"hostDevices":true}`)

		Expect(err).To(HaveOccurred())
	})

	It("rejects trailing content in strict mode", func() {
		_, err := ParseGraceVirtualizationConfigStrict(`{} {}`)

		Expect(err).To(HaveOccurred())
	})

	It("returns nil when runtime annotation is absent, empty, or unparseable", func() {
		Expect(GetGraceVirtualizationConfig(nil)).To(BeNil())
		Expect(GetGraceVirtualizationConfig(&v1.VirtualMachineInstance{})).To(BeNil())

		vmi := &v1.VirtualMachineInstance{}
		vmi.Annotations = map[string]string{v1.GraceVirtualizationAnnotation: ""}
		Expect(GetGraceVirtualizationConfig(vmi)).To(BeNil())

		vmi.Annotations = map[string]string{v1.GraceVirtualizationAnnotation: "{"}
		Expect(GetGraceVirtualizationConfig(vmi)).To(BeNil())
	})

	It("extracts the runtime annotation config", func() {
		vmi := &v1.VirtualMachineInstance{}
		vmi.Annotations = map[string]string{
			v1.GraceVirtualizationAnnotation: `{"smmuv3":true}`,
		}

		cfg := GetGraceVirtualizationConfig(vmi)

		Expect(cfg).ToNot(BeNil())
		Expect(GraceFieldEnabled(cfg.SMMUv3)).To(BeTrue())
		Expect(GraceFieldEnabled(cfg.VCMDQ)).To(BeFalse())
		Expect(GraceFieldEnabled(cfg.EGM)).To(BeFalse())
	})
})
