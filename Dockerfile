#
# Copyright 2021-2025 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Builder stage for compiling go source code
FROM golang:1.24 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Copy the go source
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lustre-csi-driver ./pkg/hpelustreplugin

ENTRYPOINT ["/bin/sh"]

# The final application stage, minimal image with compiled binary copied in
# We're using openSUSE's bci-base image since it has the mount binary, and
# is what we've built the Cray /sbin/mount.lustre user-space tool for.
FROM registry.suse.com/bci/bci-base:latest

# Remove timezone configuration so we can inherit from host
RUN rm -rf /etc/timezone && rm -rf /etc/localtime

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/lustre-csi-driver .

# Add Cray-built mount.lustre binary to layer
# See mount.lustre description in sbin/README.md
COPY sbin/mount.lustre-cray /sbin/mount.lustre

ENTRYPOINT ["/lustre-csi-driver"]
