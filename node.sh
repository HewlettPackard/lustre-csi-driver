#!/bin/bash

CMD=$1
YAML=${2:-lustre}

if [[ "$CMD" == delete ]] || [[ "$CMD" == reset ]]; then
    kustomize build "config/deploy/${YAML}" | kubectl delete -f -
fi

if [[ "$CMD" == create ]] || [[ "$CMD" == reset ]]; then
    kustomize build "config/deploy/${YAML}" | kubectl create -f -
fi