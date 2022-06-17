#!/bin/bash

function print_usage {
  echo -e "DESCRIPTION\n\tHelper script for setting up a basic Lustre filesystem on 3 NVMe namespaces. These namespaces are"
  echo -e "\tassumed to be /dev/nvme0n1, /dev/nvme1n1, and /dev/nvme2n1. A single MGS, MDT, and OST are established."
  echo -e "\nUSAGE\n\tsetup-lustre.sh <fs_name> <mgs_nid>\n"
  echo -e "\tfs_name\tName of the Lustre filesystem"
  echo -e "\tmgs_nid\tNetwork Identifier (NID) of the Lustre MGS host"
  echo -e "\nEXAMPLE\n\t./setup-lustre.sh testfs-lustre lustre-dev-01@tcp\n"
}

# Check correct number of arguments
[[ ! $# -eq 2 ]] && print_usage && exit 1

FS_NAME=$1
MGS_NID=$2

echo "Preparing Lustre file system $FS_NAME..."

mkdir -p /mnt/mgt
mkdir -p /mnt/mdt
mkdir -p /mnt/ost

# MGS
mkfs.lustre --mgs /dev/nvme0n1
mount -t lustre /dev/nvme0n1 /mnt/mgt
lctl dl

# MDT
mkfs.lustre --mdt --fsname="$FS_NAME" --mgsnode="$MGS_NID" --index=0 /dev/nvme1n1
mount -t lustre /dev/nvme1n1 /mnt/mdt
lctl dl

# OST
mkfs.lustre --ost --fsname="$FS_NAME" --mgsnode="$MGS_NID" --index=0 /dev/nvme2n1
mount -t lustre /dev/nvme2n1 /mnt/ost
lctl dl

# Now you are ready to run the 
echo "Lustre setup is complete for filesystem $FS_NAME"