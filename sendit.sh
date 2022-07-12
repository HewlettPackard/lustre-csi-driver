#!/bin/bash

docker build -t lustre-csi-driver .
docker save -o lustre-csi-driver_image.tar lustre-csi-driver
scp lustre-csi-driver_image.tar shira:~/ccarlson
ssh shira "scp ~/ccarlson/lustre-csi-driver_image.tar mercury01:~/ccarlson"
