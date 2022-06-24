# Initial application stage
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

# Final application stage
FROM redhat/ubi8-minimal

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/lustre-csi-driver .

ENTRYPOINT ["/lustre-csi-driver"]

