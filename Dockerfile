# -----------------------------------------------------------------
# Dockerfile -
#
# Provides Docker image build instructions for nnf-csi-driver
#
# Author: Nate Roiger
#
# © Copyright 2021 Hewlett Packard Enterprise Development LP
#
# -----------------------------------------------------------------

# Builder
FROM golang:1.17 as builder

WORKDIR /workspace
COPY ./ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPRIVATE=stash.us.cray.com go build -o nnf-csi-driver


# Final application container
FROM arti.dev.cray.com/baseos-docker-master-local/centos:latest

WORKDIR /
COPY --from=builder /workspace/nnf-csi-driver .

ENTRYPOINT [ "/nnf-csi-driver" ]