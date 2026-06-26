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
# sync-nv-rpms-update-workspace.awk
#
# Called by hack/sync-nv-rpms.sh to update the WORKSPACE file in a single
# pass. It performs two operations:
#
#   1. Remove existing el10nv rpm() blocks for the sub-packages we manage.
#   2. Insert new el10nv rpm() blocks immediately after the last el10 entry
#      for each sub-package.
#
# Required variables (passed via -v):
#
#   insert_map_file   - Path to a file mapping bazel names to insertion
#                       content files.  Each line has the format:
#                           <last_el10_name>|<path_to_insertion_file>
#                       When we finish printing the rpm() block whose name
#                       matches <last_el10_name>, we append the contents of
#                       the corresponding insertion file.
#
#   subpkg_list_file  - Path to a file listing the sub-package prefixes we
#                       manage (one per line, e.g. "libvirt-libs").  Any
#                       existing rpm() block whose name starts with one of
#                       these prefixes and contains ".el10nv." is considered
#                       stale and is removed.
#

# ---------------------------------------------------------------------------
# Initialisation: load the insertion map and sub-package list into memory.
# ---------------------------------------------------------------------------
BEGIN {
    # Build insert_after[name] = filepath from the map file.
    while ((getline line < insert_map_file) > 0) {
        idx = index(line, "|")
        if (idx > 0) {
            iname = substr(line, 1, idx - 1)
            ifile = substr(line, idx + 1)
            insert_after[iname] = ifile
        }
    }
    close(insert_map_file)

    # Load sub-package prefixes into the sp[] array.
    n_sp = 0
    while ((getline line < subpkg_list_file) > 0) {
        sp[++n_sp] = line
    }
    close(subpkg_list_file)

    in_block = 0
    out_n = 0
}

# ---------------------------------------------------------------------------
# Helper: buffer a line for output.
# ---------------------------------------------------------------------------
function buf(line) { out[++out_n] = line }

# ---------------------------------------------------------------------------
# Helper: read all lines from fpath and buffer them.
# ---------------------------------------------------------------------------
function insert_from_file(fpath,    l) {
    while ((getline l < fpath) > 0) buf(l)
    close(fpath)
}

# ---------------------------------------------------------------------------
# Detect the start of an rpm() block and begin accumulating its lines.
# We need to collect the entire block before deciding whether to keep,
# remove, or augment it.
# ---------------------------------------------------------------------------
/^rpm\(/ {
    in_block = 1
    delete blk
    blk_n = 0
    blk[++blk_n] = $0
    next
}

# ---------------------------------------------------------------------------
# Accumulate lines inside an rpm() block until we hit the closing ")".
# Once the block is complete we:
#   - Extract the "name" field.
#   - Decide whether this block should be removed (skip=1) because it is an
#     el10nv entry for a sub-package we manage.
#   - If kept, buffer the block and optionally append new entries after it.
# ---------------------------------------------------------------------------
in_block {
    blk[++blk_n] = $0
    if ($0 == ")") {
        in_block = 0

        # Extract the name = "..." value from the block.
        name = ""
        for (i = 1; i <= blk_n; i++) {
            if (blk[i] ~ /name = "/) {
                name = blk[i]
                sub(/.*name = "/, "", name)
                sub(/".*/, "", name)
                break
            }
        }

        # Check whether this el10nv block belongs to a managed sub-package.
        skip = 0
        if (name ~ /\.el10nv\./) {
            for (i = 1; i <= n_sp; i++) {
                pat = "^" sp[i] "-[0-9]"
                if (name ~ pat) {
                    skip = 1
                    break
                }
            }
        }

        if (skip) {
            # Drop the stale el10nv block and any trailing blank line that
            # preceded it in the output buffer.
            if (out_n > 0 && out[out_n] == "") out_n--
        } else {
            # Keep this block.
            for (i = 1; i <= blk_n; i++) buf(blk[i])

            # If this is the anchor block (last el10 entry for a
            # sub-package), insert the new el10nv entries after it.
            if (name in insert_after) {
                insert_from_file(insert_after[name])
            }
        }
    }
    next
}

# ---------------------------------------------------------------------------
# Lines outside any rpm() block are passed through unchanged.
# ---------------------------------------------------------------------------
{ buf($0) }

# ---------------------------------------------------------------------------
# Flush the output buffer.
# ---------------------------------------------------------------------------
END {
    for (i = 1; i <= out_n; i++) print out[i]
}
