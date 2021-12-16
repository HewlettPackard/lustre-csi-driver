#!/bin/bash

# This assumes 3 available NVMe Namespaces: /dev/nvme0n1, /dev/nvme1n1, /dev/nvme2n1

FSNAME=${1:-lustre}
echo "Preparing Lustre file system $FSNAME..."

mkdir -p /mnt/mgt
mkdir -p /mnt/mdt
mkdir -p /mnt/ost

# MGS
mkfs.lustre --mgs /dev/nvme0n1
mount -t lustre /dev/nvme0n1 /mnt/mgt
lctl dl

# MDT
mkfs.lustre --mdt --fsname=$FSNAME --mgsnode=rabbit-dev-01@tcp --index=0 /dev/nvme1n1
mount -t lustre /dev/nvme1n1 /mnt/mdt
lctl dl

# OST
mkfs.lustre --ost --fsname=$FSNAME --mgsnode=rabbit-dev-01@tcp --index=0 /dev/nvme2n1
mount -t lustre /dev/nvme2n1 /mnt/ost
lctl dl

# Now you are ready to run the 
echo "Lustre setup is complete for file system $FSNAME"