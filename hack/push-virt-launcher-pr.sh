#!/usr/bin/env bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright the KubeVirt Authors.
#

set -eo pipefail

DOCKER_PREFIX=${DOCKER_PREFIX:-"quay.io/kubevirt/kubevirt-aie"}
KUBEVIRT_CENTOS_STREAM_VERSION=${KUBEVIRT_CENTOS_STREAM_VERSION:-"10"}

branch=$(git rev-parse --abbrev-ref HEAD)
branch=$(printf '%s' "${branch}" | sed 's/[^a-zA-Z0-9_.-]/-/g' | cut -c 1-100)
short_sha=$(git rev-parse --short=10 HEAD)
docker_tag="wip-${branch}-${short_sha}"
DOCKER_TAG=${DOCKER_TAG:-"${docker_tag}"}

echo "Building and pushing virt-launcher manifest"
echo "  DOCKER_PREFIX: ${DOCKER_PREFIX}"
echo "  DOCKER_TAG:    ${DOCKER_TAG}"
echo "  KUBEVIRT_CENTOS_STREAM_VERSION: ${KUBEVIRT_CENTOS_STREAM_VERSION}"
echo ""
echo "  Image: ${DOCKER_PREFIX}/virt-launcher:${DOCKER_TAG}"
echo ""

export KUBEVIRT_CENTOS_STREAM_VERSION
export DOCKER_PREFIX
export DOCKER_TAG
export BUILD_ARCH="amd64,crossbuild-aarch64"
export PUSH_TARGETS="virt-launcher"

make bazel-push-images
