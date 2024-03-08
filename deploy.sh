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
set -o pipefail

# Deploy/undeploy controller to the K8s cluster specified in ~/.kube/config.

CMD=$1
KUSTOMIZE=$2
OVERLAY_DIR=$3

if [[ $CMD == 'deploy' ]]; then
    $KUSTOMIZE build "$OVERLAY_DIR" | kubectl apply -f -
fi

if [[ $CMD == 'undeploy' ]]; then
    # Do not touch the namespace resource when deleting this service.
    # Wishing for yq(1)...
    $KUSTOMIZE build "$OVERLAY_DIR" | python3 -c 'import yaml, sys; all_docs = yaml.safe_load_all(sys.stdin); less_docs=[doc for doc in all_docs if doc["kind"] != "Namespace"]; print(yaml.dump_all(less_docs))' |  kubectl delete --ignore-not-found -f -
fi

exit 0

