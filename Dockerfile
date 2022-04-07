FROM arti.dev.cray.com/baseos-docker-master-local/centos8:centos8 AS base

WORKDIR /

# Install basic dependencies
RUN dnf install -y make git gzip wget gcc tar

# Retrieve lustre-rpms
WORKDIR /tmp/lustre-rpms

RUN wget \
  http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/kmod-lustre-2.14.0-1.el8.x86_64.rpm \
  http://arti.dev.cray.com/artifactory/kj-rpm-unstable-local/predev/centos8.4.2105-lustre-zfs/storage-mirrors/x86_64/lustre-2.14.0-1.el8.x86_64.rpm

WORKDIR /

ENV GO_VERSION=1.17.6
RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && tar -xzf go${GO_VERSION}.linux-amd64.tar.gz


# Start from scratch to make the base stage for the final application.
# Build it here so it won't be invalidated when we COPY the controller source
# code in the next layer.
FROM arti.dev.cray.com/baseos-docker-master-local/centos8:centos8 AS application-base

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

# Set Go environment
ENV GOROOT="/go"
ENV PATH="${PATH}:${GOROOT}/bin" GOPRIVATE="github.hpe.com"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY config/ config/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nnf-csi-driver main.go

ENTRYPOINT ["/bin/sh"]

#FROM builder as testing
#WORKDIR /workspace
#
#COPY Makefile .
#
#RUN go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest && \
#    make manifests && make generate && make fmt &&  make vet && \
#    mkdir -p /workspace/testbin && /bin/bash -c "test -f /workspace/testbin/setup-envtest.sh || curl -sSLo /workspace/testbin/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.2/hack/setup-envtest.sh" && \
#    /bin/bash -c "source /workspace/testbin/setup-envtest.sh; fetch_envtest_tools /workspace/testbin; setup_envtest_env /workspace/testbin"
#
#ENTRYPOINT ["sh", "/workspace/initiateContainerTest.sh"]

# The final application stage.
FROM application-base

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/nnf-csi-driver .

ENTRYPOINT ["/nnf-csi-driver"]

