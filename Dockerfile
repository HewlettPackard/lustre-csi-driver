FROM golang:1.17 as builder

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
# Start from scratch to make the base stage for the final application.
# Build it here so it won't be invalidated when we COPY the controller source
# code in the next layer.
FROM centos:centos8 AS application-base

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/lustre-csi-driver .

# Set repo mirror and install basic dependencies
RUN sed -i -e "s|mirrorlist=|#mirrorlist=|g" /etc/yum.repos.d/CentOS-*
RUN sed -i -e "s|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g" /etc/yum.repos.d/CentOS-*
RUN dnf install -y wget tar

# Retrieve lustre-rpms
WORKDIR /tmp/lustre-rpms

RUN wget https://downloads.whamcloud.com/public/lustre/lustre-2.14.0/el8.3/client/RPMS/x86_64/lustre-client-2.14.0-1.el8.x86_64.rpm
  # https://downloads.whamcloud.com/public/lustre/lustre-2.14.0/el8.3/client/RPMS/x86_64/kmod-lustre-client-2.14.0-1.el8.x86_64.rpm
  # http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/kmod-lustre-2.14.0-1.el8.x86_64.rpm \
  # http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/lustre-2.14.0-1.el8.x86_64.rpm

RUN dnf clean all && \
    rpm -Uivh --nodeps lustre-* && \
    rm *.rpm

WORKDIR /

# Add in mount.lustre binary
COPY ./sbin/mount.lustre /sbin/mount.lustre


ENTRYPOINT ["/lustre-csi-driver"]
