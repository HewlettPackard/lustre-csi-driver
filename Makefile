
VERSION ?= $(shell sed 1q .version)
IMAGE_TAG_BASE ?= arti.dev.cray.com/rabsw-docker-master-local/cray-dp-lustre-csi-driver
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

# Tell Kustomize to deploy the default config, or an overlay.
# To use the 'lustre' overlay:
#   export KUBECONFIG=/my/craystack/kubeconfig.file
#   make deploy OVERLAY=lustre
OVERLAY ?= base

all: build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

build: fmt vet docker-build
	go build -o bin/lustre-csi-driver

run: fmt vet
	go run ./main.go

docker-build: Dockerfile fmt vet
	# Name the base stages so they are not lost during a cache prune.
	time docker build -t ${IMG}-base --target base .
	time docker build -t ${IMG}-app-base --target application-base .
	time docker build -t ${IMG} .

kind-push:
	kind load docker-image $(IMG)

deploy: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/deploy/base && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/deploy/${OVERLAY} | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/deploy/${OVERLAY} | kubectl delete -f -

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
