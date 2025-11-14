# Copyright 2019 The Kubernetes Authors.
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

FROM golang:1.24 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY Makefile go.mod go.sum ./

# Copy the go source
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
ARG VERSION=undefined
RUN make hpelustre-direct

FROM ubuntu:22.04

WORKDIR "/app"

COPY "./pkg/hpelustreplugin/entrypoint.sh" "/app/entrypoint.sh"
RUN chmod +x "/app/entrypoint.sh"

# Add Cray-built mount.lustre binary to layer
# See mount.lustre description in sbin/README.md
COPY sbin/mount.lustre-cray /sbin/mount.lustre

# By now the plugin build in the "builder" stage should have completed.
COPY --from=builder /workspace/bin/hpelustreplugin /app/hpelustreplugin

ENTRYPOINT ["/app/entrypoint.sh", "/app/hpelustreplugin"]
