#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -euo pipefail

function usage {
    echo "Usage: $0 [branch|local|url]"
    echo
    echo "branch: The branch from which to install the Azure Lustre CSI Driver to install. Default is 'main'."
    echo "local: Deploy out of local filesystem."
    echo
    echo "Example:"
    echo "$0 # install from remote main"
    echo "$0 main # install from remote branch or reference"
    echo "$0 local # install from locally checked out branch"
    echo "$0 https://raw.githubusercontent.com/csmuell/azurelustre-csi-driver/main # install from given remote repository/branch"
    exit 1
}

if [[ "$#" -gt 1 || ("$#" -gt 0 && "$1" == "--help") ]]; then
  usage
fi

branch="main"
repo="https://raw.githubusercontent.com/kubernetes-sigs/azurelustre-csi-driver/${branch}/deploy"

if [[ "$#" -eq 1 ]]; then
  case "$1" in
    local)
      repo="$(git rev-parse --show-toplevel)/deploy"
      ;;
    http*)
      repo="${1}/deploy"
      ;;
    *)
      branch="${1}"
      repo="https://raw.githubusercontent.com/kubernetes-sigs/azurelustre-csi-driver/${branch}/deploy"
      ;;
  esac
fi

verify="${repo}/install-driver.sh"
if ! [ -f "${verify}" ]; then
  if ! curl -L -Is --fail "${verify}" > /dev/null; then
    echo "Unknown repository: ${repo} ${verify} does not exist."
    usage
  fi
fi

echo
echo "Installing Azure Lustre CSI Driver branch: $branch, repo: $repo ..."

kubectl apply -f "$repo/rbac-csi-azurelustre-controller.yaml"
kubectl apply -f "$repo/rbac-csi-azurelustre-node.yaml"
kubectl apply -f "$repo/csi-azurelustre-driver.yaml"
kubectl apply -f "$repo/csi-azurelustre-controller.yaml"
kubectl apply -f "$repo/csi-azurelustre-node.yaml"

kubectl rollout status deployment csi-hpelustre-controller -n lustre-csi-system --timeout=300s
kubectl rollout status daemonset csi-hpelustre-node -n lustre-csi-system --timeout=1800s
echo 'Azure Lustre CSI driver installed successfully.'
