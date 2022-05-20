FROM golang:1.17 as builder

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
#ENTRYPOINT ["bash", "/workspace/initiateContainerTest.sh"]

# The final application stage.
FROM redhat/ubi8-minimal

WORKDIR /
# Retrieve executable from previous layer
COPY --from=builder /workspace/nnf-csi-driver .

ENTRYPOINT ["/nnf-csi-driver"]

