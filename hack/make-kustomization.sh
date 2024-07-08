#!/bin/bash

# Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

set -e

OVERLAY_DIR=$1
OVERLAY=$2
IMAGE_TAG_BASE=$3
TAG=$4

if [[ ! -d $OVERLAY_DIR ]]
then
    mkdir "$OVERLAY_DIR"
fi

COMPONENT_LABELS="
    - op: add
      path: /metadata/labels/app.kubernetes.io~1version
      value: "$TAG"
    - op: add
      path: /metadata/labels/app.kubernetes.io~1component
      value: lustre-csi-driver
"

NNF_VER_LABELS=""
if [[ -n $NNF_VERSION ]]
then
    NNF_VER_LABELS="
    - op: add
      path: /metadata/labels/app.kubernetes.io~1nnf-version
      value: "$NNF_VERSION"
    - op: add
      path: /metadata/labels/app.kubernetes.io~1part-of
      value: nnf
"
fi

cat <<EOF > "$OVERLAY_DIR"/kustomization.yaml
resources:
- ../$OVERLAY

patches:
- target:
    kind: Deployment
  patch: |-
$COMPONENT_LABELS
$NNF_VER_LABELS
- target:
    kind: DaemonSet
  patch: |-
$COMPONENT_LABELS
$NNF_VER_LABELS

apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
images:
- name: $IMAGE_TAG_BASE
  newTag: $TAG
EOF

