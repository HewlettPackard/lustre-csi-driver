# Copyright 2021, 2022 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Default container tool to use.
#   To use podman:
#   $ DOCKER=podman make docker-build
DOCKER ?= docker

VERSION ?= $(shell sed 1q .version)
IMAGE_TAG_BASE ?= ghcr.io/hewlettpackard/lustre-csi-driver
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

# Tell Kustomize to deploy the default config, or an overlay.
# To use the 'kind' overlay:
#   export KUBECONFIG=/my/kubeconfig.file
#   make deploy OVERLAY=overlays/kind
# Or, make kind-deploy
# To deploy the base lustre config:
#   make deploy

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
	time ${DOCKER} build -t $(IMG) .

edit-image: ## Replace plugin.yaml image with name "controller" -> $IMG variable
	cd deploy/kubernetes/base && $(KUSTOMIZE) edit set image controller=$(IMG)

kind-push: ## Push image to Kind environment
	kind load docker-image $(IMG)

deploy_overlay: kustomize edit-image ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build deploy/kubernetes/$(OVERLAY) | kubectl apply -f -

deploy: OVERLAY ?= base
deploy: deploy_overlay

kind-deploy: OVERLAY=overlays/kind
kind-deploy: deploy_overlay

undeploy_overlay: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build deploy/kubernetes/$(OVERLAY) | kubectl delete -f -

undeploy: OVERLAY ?= lustre
undeploy: undeploy_overlay

kind-undeploy: OVERLAY=overlays/kind
kind-undeploy: undeploy_overlay

installer-gen: kustomize edit-image
	$(KUSTOMIZE) build deploy/kubernetes/$(OVERLAY) > deploy/kubernetes/lustre-csi-driver$(OVERLAY_LABEL).yaml

installer: ## Generates full .yaml output from Kustomize for base and overlays
	make installer-gen OVERLAY=base && make installer-gen OVERLAY=overlays/kind OVERLAY_LABEL=-kind

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
