FROM centos:centos8 AS base

WORKDIR /

# Install basic dependencies
RUN sed -i -e "s|mirrorlist=|#mirrorlist=|g" /etc/yum.repos.d/CentOS-*
RUN sed -i -e "s|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g" /etc/yum.repos.d/CentOS-*
RUN dnf install -y wget tar

# Retrieve lustre-rpms
WORKDIR /tmp/lustre-rpms

RUN wget https://downloads.whamcloud.com/public/lustre/lustre-2.14.0/el8.3/client/RPMS/x86_64/lustre-client-2.14.0-1.el8.x86_64.rpm \
  https://downloads.whamcloud.com/public/lustre/lustre-2.14.0/el8.3/client/RPMS/x86_64/kmod-lustre-client-2.14.0-1.el8.x86_64.rpm
  # http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/kmod-lustre-2.14.0-1.el8.x86_64.rpm \
  # http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/lustre-2.14.0-1.el8.x86_64.rpm


# Start from scratch to make the base stage for the final application.
# Build it here so it won't be invalidated when we COPY the controller source
# code in the next layer.
FROM centos:centos8 AS application-base

WORKDIR /
# Retrieve executable from previous layer
COPY --from=base /tmp/lustre-rpms/*.rpm /root/

# Retrieve built rpms from previous layer and install Lustre dependencies
WORKDIR /root/
RUN dnf clean all && \
    rpm -Uivh --nodeps lustre-* kmod-* && \
    rm /root/*.rpm

ENTRYPOINT ["/bin/sh"]

#
# Note: The COPY commands below have the potential to invalidate any layer
# that follows.
#

FROM base as builder

# Install go
ENV GO_VERSION=1.17.6
RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Set Go environment
ENV PATH="${PATH}:/usr/local/go/bin" GOPRIVATE="github.hpe.com"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lustre-csi-driver main.go

ENTRYPOINT ["/bin/sh"]


# The final application stage.
FROM application-base

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/lustre-csi-driver .

ENTRYPOINT ["/lustre-csi-driver"]
