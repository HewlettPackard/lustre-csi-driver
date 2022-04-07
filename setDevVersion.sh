#!/bin/bash
# Use `source`` to push the environment variables back into the calling shell
# "source ./setDevVersion.sh"

# Command to access arti and list all of the docker images there filtering out the tag for the
# version matching the current branch's git SHA
unset VERSION
unset IMAGE_TAG_BASE

# Setup some artifactory paths to the developer and master branch locations
ARTI_URL_DEV=https://arti.dev.cray.com/artifactory/rabsw-docker-unstable-local/cray-dp-lustre-csi-driver/
ARTI_URL_MASTER=https://arti.dev.cray.com/artifactory/rabsw-docker-master-local/cray-dp-lustre-csi-driver/

# Retrieve the name of the current branch. If we are in detached HEAD state, assume it is master.
CURRENT_BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)

# Depending on whether we are on the master branch or not, setup for deployment from artifactory
# NOTE: Detached HEAD state is assumed to match 'master'
if [[ "$CURRENT_BRANCH_NAME" == "master" ]] || [[ "$CURRENT_BRANCH_NAME" == "HEAD" ]]; then
    ARTI_URL="$ARTI_URL_MASTER"
else    # not on the master branch
    ARTI_URL="$ARTI_URL_DEV"

    # Deploying a developer build requires the IMAGE_TAG_BASE to change as well.
    # Master branch is the default, so we don't change it when we are on the master branch.
    IMAGE_TAG_BASE=arti.dev.cray.com/rabsw-docker-unstable-local/cray-dp-lustre-csi-driver
    export IMAGE_TAG_BASE
    echo IMAGE_TAG_BASE: "$IMAGE_TAG_BASE"
fi

# Locate the container tags in arti to set the VERSION environment variable
# which allows us to run `make deploy` and pull the correct version from ARTI.
LATEST_LOCAL_COMMIT=$(git rev-parse --short HEAD)
ARTI_TAG=$(wget --spider --recursive --no-parent -l1 "$ARTI_URL" 2>&1 | grep -- ^-- | awk '{print $3}' | grep "$LATEST_LOCAL_COMMIT" | tail -1)
VERSION=$(basename "$ARTI_TAG")
export VERSION
echo VERSION: "$VERSION"
