#!/bin/bash

POD=${1:-$(kubectl get pods --no-headers | grep csi-nnf-node | awk '{print $1}' | head -n1)}
CONTAINER=${2:-nnf-csi-driver}

kubectl logs pod/$POD $CONTAINER