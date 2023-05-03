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
IMAGE_TAG_BASE ?= ghcr.io/hewlettpackard/lustre-csi-driver

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

build: VERSION ?= $(shell cat .version)
build: .version fmt vet
	go mod vendor
	go build -o bin/lustre-csi-driver

run: fmt vet
	go run ./main.go

docker-build: VERSION ?= $(shell cat .version)
docker-build: .version Dockerfile fmt vet
	time ${DOCKER} build -t $(IMAGE_TAG_BASE):$(VERSION) .

edit-image: VERSION ?= $(shell cat .version)
edit-image: .version ## Replace plugin.yaml image with name "controller" -> ghcr tagged container reference
	cd deploy/kubernetes/base && $(KUSTOMIZE) edit set image controller=$(IMAGE_TAG_BASE):$(VERSION)

kind-push: VERSION ?= $(shell cat .version)
kind-push: .version ## Push image to Kind environment
	kind load docker-image $(IMAGE_TAG_BASE):$(VERSION)

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

installer-gen: kustomize edit-image helm-version
	$(KUSTOMIZE) build deploy/kubernetes/$(OVERLAY) > deploy/kubernetes/lustre-csi-driver$(OVERLAY_LABEL).yaml

installer: ## Generates full .yaml output from Kustomize for base and overlays
	make installer-gen OVERLAY=base && make installer-gen OVERLAY=overlays/kind OVERLAY_LABEL=-kind

helm-version: VERSION ?= $(shell cat .version)
helm-version: .version ## Updates the Helm values.yaml with new version
	yq e -i ".deployment.tag=\"$(VERSION)\"" charts/lustre-csi-driver/values.yaml

# Let .version be phony so that a git update to the workarea can be reflected
# in it each time it's needed.
.PHONY: .version
.version: ## Uses the git-version-gen script to generate a tag version
	./git-version-gen --fallback `git rev-parse HEAD` > .version

clean:
	rm -f .version


## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }
