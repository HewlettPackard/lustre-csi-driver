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

ARG VERSION=undefined

WORKDIR /workspace

# Copy the Go Modules manifests
COPY Makefile go.mod go.sum ./

# Copy the go source
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN make hpelustre-direct

FROM ubuntu:22.04

RUN apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y \
   gpg curl ca-certificates iproute2 kmod libc-dev-bin libc6 libc6-dev  \
   libcrypt-dev libgdbm-compat4 libgdbm6 libkmod2     \
   libmysqlclient21 libnetsnmptrapd40 libnl-3-200 libnl-genl-3-200 libnsl-dev libpci-dev libpci3    \
   libperl5.34 libsensors-config libsensors-dev libsensors5 libsnmp-base libsnmp-dev libsnmp40      \
   libssl-dev libtirpc-dev libudev-dev libwrap0 libwrap0-dev libyaml-0-2 linux-base linux-base-sgx  \
   linux-libc-dev mysql-common pci.ids perl perl-modules-5.34 rpcsvc-proto zlib1g-dev -y && \
  apt-get autoremove -y && \
  apt-get clean -y && \
  rm -rf \
  /var/cache/debconf/* \
  /var/lib/apt/lists/* \
  /var/log/* \
  /tmp/* \
  /var/tmp/*

WORKDIR "/app"

#LABEL maintainers="dabradley;t-mialve"
#LABEL description="Azure Lustre CSI driver"
LABEL maintainers="HPE"
LABEL description="HPE Lustre CSI driver"

COPY --from=builder /workspace/bin/hpelustreplugin /app/hpelustreplugin
COPY "./pkg/hpelustreplugin/entrypoint.sh" "/app/entrypoint.sh"

RUN chmod +x "/app/entrypoint.sh"

ENTRYPOINT ["/app/entrypoint.sh", "/app/hpelustreplugin"]
