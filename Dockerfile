# Builder stage for compiling go source code
FROM golang:1.19 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/

# Retrieve go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lustre-csi-driver main.go

ENTRYPOINT ["/bin/sh"]

# The final application stage, minimal image with compiled binary copied in
# We're using openSUSE's bci-base image since it has the mount binary, and
# is what we've built the Cray /sbin/mount.lustre user-space tool for.
FROM registry.suse.com/bci/bci-base:latest

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/lustre-csi-driver .

# Add Cray-built mount.lustre binary to layer
# See mount.lustre description in sbin/README.md
COPY sbin/mount.lustre-cray /sbin/mount.lustre

ENTRYPOINT ["/lustre-csi-driver"]