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
# Copyright 2026 Red Hat, Inc.
#
# Merges upstream kubevirt changes into the current branch.
#
# On success the working tree contains a merge commit ready to push.
# On merge conflict the script prints conflict details and exits 2.
#
# Usage:
#   hack/sync-upstream.sh [--upstream-remote <remote>] \
#                         [--upstream-branch <branch>] \
#                         [--target-branch <branch>]
#

set -e
set -o pipefail

UPSTREAM_REMOTE="upstream"
UPSTREAM_BRANCH="release-1.8"
TARGET_BRANCH="release-1.8-aie-nv"

while [[ $# -gt 0 ]]; do
    case "$1" in
    --upstream-remote)
        UPSTREAM_REMOTE="$2"
        shift 2
        ;;
    --upstream-branch)
        UPSTREAM_BRANCH="$2"
        shift 2
        ;;
    --target-branch)
        TARGET_BRANCH="$2"
        shift 2
        ;;
    *)
        echo "Unknown argument: $1" >&2
        exit 1
        ;;
    esac
done

echo "Fetching ${UPSTREAM_REMOTE}/${UPSTREAM_BRANCH}..."
git fetch "${UPSTREAM_REMOTE}" "${UPSTREAM_BRANCH}"

upstream_head=$(git rev-parse --short "${UPSTREAM_REMOTE}/${UPSTREAM_BRANCH}")
echo "Upstream HEAD: ${upstream_head}"

echo "Attempting merge of ${UPSTREAM_REMOTE}/${UPSTREAM_BRANCH} into ${TARGET_BRANCH}..."
if ! git merge --no-edit "${UPSTREAM_REMOTE}/${UPSTREAM_BRANCH}"; then
    echo ""
    echo "Merge conflict detected."
    echo "Conflicting files:"
    git diff --name-only --diff-filter=U
    git merge --abort
    exit 2
fi

echo "Merge successful (upstream HEAD: ${upstream_head})."
