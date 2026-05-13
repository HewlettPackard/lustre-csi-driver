# Copyright 2021-2026 Hewlett Packard Enterprise Development LP
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

CHART_DIR ?= charts

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Tell Kustomize to deploy the default config, or an overlay.
# To use the 'kind' overlay:
#   export KUBECONFIG=/my/kubeconfig.file
#   make deploy OVERLAY=overlays/kind
# Or, make kind-deploy
# To deploy the base lustre config:
#   make deploy
OVERLAY ?= overlays/kind

all: build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

build: VERSION ?= $(shell cat .version)
build: .version fmt vet
	go build -o bin/hpelustreplugin ./pkg/hpelustreplugin

run: fmt vet
	go run ./pkg/hpelustreplugin

# The current context of image building
# The architecture of the image
ARCH ?= amd64
# Output type of docker buildx build
OUTPUT_TYPE ?= registry

PKG = github.com/HewlettPackard/lustre-csi-driver

dockerfile = ./pkg/hpelustreplugin/Dockerfile

.PHONY: hpelustre
hpelustre: VERSION ?= $(shell cat .version)
hpelustre: .version fmt vet
hpelustre: hpelustre-direct

# hpelustre-direct is called by the Dockerfile.
.PHONY: hpelustre-direct
hpelustre-direct:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -a -ldflags="-X '$(PKG)/pkg/hpelustre.driverVersion=$(VERSION)' -s -w -extldflags -static" -mod vendor -o bin/hpelustreplugin ./pkg/hpelustreplugin

.PHONY: docker-build
docker-build: VERSION ?= $(shell cat .version)
docker-build: .version fmt vet
	docker buildx build --platform="linux/$(ARCH)" \
		-t $(IMAGE_TAG_BASE):$(VERSION) --build-arg VERSION=$(VERSION) --build-arg ARCH=$(ARCH) -f $(dockerfile) .

edit-image: VERSION ?= $(shell cat .version)
edit-image: .version ## Replace plugin.yaml image with name "controller" -> ghcr tagged container reference
	$(KUSTOMIZE_IMAGE_TAG) deploy/kubernetes/begin $(OVERLAY) $(IMAGE_TAG_BASE) $(VERSION)

kind-push: VERSION ?= $(shell cat .version)
kind-push: .version ## Push image to Kind environment
	kind load docker-image $(IMAGE_TAG_BASE):$(VERSION)

deploy_overlay: kustomize edit-image ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	./deploy.sh deploy $(KUSTOMIZE) deploy/kubernetes/begin

.PHONY: deploy
deploy: OVERLAY ?= base
deploy: deploy_overlay

kind-deploy: OVERLAY=overlays/kind
kind-deploy: deploy_overlay

undeploy_overlay: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	./deploy.sh undeploy $(KUSTOMIZE) deploy/kubernetes/$(OVERLAY)

undeploy: OVERLAY ?= lustre
undeploy: undeploy_overlay

kind-undeploy: OVERLAY=overlays/kind
kind-undeploy: undeploy_overlay

installer: kustomize edit-image helm-app-version

.PHONY: helm-one-chart
helm-one-chart:
	if [[ `find $(CHART_DIR)/ -maxdepth 1 -mindepth 1 -type d | wc -l | tr -d ' '` != 1 ]]; then \
		echo "Expecting only one chart version in $(CHART_DIR)."; \
		exit 1; \
	fi

# Update the application version in the helm chart.
# Use this prior to using the 'helm-reindex' and 'helm-repackage' targets.
# This expects only one helm chart to exist at a time.
.PHONY: helm-app-version
helm-app-version: VERSION ?= $(shell cat .version)
helm-app-version: CHART_VERSION ?= $(shell /bin/ls -d1 $(CHART_DIR)/v*)
helm-app-version: .version helm-one-chart ## Updates the Helm values.yaml with new version.
	yq -i e -M '(.deployment.tag) = "$(VERSION)"' "$(CHART_VERSION)/lustre-csi-driver/values.yaml"
	yq -i e -M '(.appVersion) = "$(VERSION)"' "$(CHART_VERSION)/lustre-csi-driver/Chart.yaml"

# Update the chart version in the helm chart. Use this if the chart version
# has changed, prior to using the 'helm-repackage' and 'helm-reindex' targets.
# This expects only one helm chart to exist at a time.
.PHONY: helm-chart-version
helm-chart-version: CHART_VERSION ?= $(shell /bin/ls -d1 $(CHART_DIR)/v*)
helm-chart-version: helm-one-chart
	export NEW_VER=`echo "$(CHART_VERSION)" | sed 's,$(CHART_DIR)/v,,'`; \
	yq -i e -M '(.version) = "'"$$NEW_VER"'"' $(CHART_VERSION)/lustre-csi-driver/Chart.yaml

# Repackage the chart. Use this if the chart version has changed
# or if the application version has changed.
# Use the 'helm-app-version' and 'helm-chart-version' targets for any new chart
# prior to running this 'helm-repackage' target.
# This expects only one helm chart to exist at a time.
.PHONY: helm-repackage
helm-repackage: CHART_VERSION ?= $(shell /bin/ls -d1 $(CHART_DIR)/v*)
helm-repackage: helm-one-chart
	find $(CHART_VERSION) -name '*.tgz' -delete
	helm package $(CHART_VERSION)/lustre-csi-driver -d $(CHART_VERSION)

# Reindex the charts/index.yaml. Use this if the chart version has changed or
# if the application version has changed.
# Use the 'helm-app-version', 'helm-chart-version', and 'helm-repackage'
# targets for any new chart prior to running this 'helm-reindex' target.
# This expects only one helm chart to exist at a time.
.PHONY: helm-reindex
helm-reindex: CHART_BRANCH ?= master
helm-reindex: helm-one-chart
	rm -f $(CHART_DIR)/index.yaml
	helm repo index $(CHART_DIR)/ --url https://raw.githubusercontent.com/HewlettPackard/lustre-csi-driver/$(CHART_BRANCH)/$(CHART_DIR)

# Let .version be phony so that a git update to the workarea can be reflected
# in it each time it's needed.
.PHONY: .version
.version: ## Uses the git-version-gen script to generate a tag version
	./git-version-gen --fallback `git rev-parse HEAD` > .version

clean:
	rm -f .version

.PHONY: clean-bin
clean-bin:
	rm -rf $(LOCALBIN)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE_IMAGE_TAG ?= ./hack/make-kustomization.sh
KUSTOMIZE ?= $(LOCALBIN)/kustomize

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(LOCALBIN) ## Download kustomize locally if necessary.
	if [[ ! -s $(LOCALBIN)/kustomize || ! $$($(LOCALBIN)/kustomize version) =~ $(KUSTOMIZE_VERSION) ]]; then \
	  rm -f $(LOCALBIN)/kustomize && \
	  { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }; \
	fi

