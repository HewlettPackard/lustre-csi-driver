#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
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

#
# Shell script to install Lustre client kernel modules and launch CSI driver
#   $1 is the path to the CSI driver.
#

set -o xtrace
set -o errexit
set -o pipefail
set -o nounset

function add_net_interfaces() {
  echo "$(date -u) Determining ethernet interfaces."
  echo "$(date -u) Route table is:"
  ip route list
  interface_list=$(ip route show | sed -n 's/.*\s\+dev\s\+\([^ ]\+\).*/\1/p' | sort -u)
  ethernet_interfaces=()
  for interface in $interface_list; do
    interface_info=$(ip link show "${interface}")
    # Skip interfaces that are not needed
    if [[ "$interface_info" =~ 'SLAVE' ]]; then
      echo "$(date -u) Not adding slave interface: ${interface}"
    elif [[ "$interface_info" =~ 'link-netns' ]]; then
      echo "$(date -u) Not adding namespaced interface: ${interface}"
    elif [[ "$interface_info" =~ 'UNKNOWN' ]]; then
      echo "$(date -u) Not adding state unknown interface: ${interface}"
    # Add remaining link/ether interface
    elif [[ "$interface_info" =~ 'link/ether' ]]; then
      echo "$(date -u) Including ethernet interface: ${interface}"
      ethernet_interfaces+=("$interface")
    else
      echo "$(date -u) Skipping non-ethernet interface: ${interface}"
    fi
  done
  echo "$(date -u) List of found ethernet interfaces is: ${ethernet_interfaces[*]}"

  if [[ "${#ethernet_interfaces[@]}" -eq 0 ]]; then
    echo "$(date -u) Cannot find any ethernet network interface"
    exit 1
  fi

  for interface in "${ethernet_interfaces[@]}"; do
    if lnetctl net show --net tcp | grep -q "\b${interface}\b"; then
      echo "$(date -u) Interface already added, skipping: ${interface}"
    else
      echo "$(date -u) Adding interface: ${interface}"
      lnetctl net add --net tcp --if "${interface}"
    fi
  done
}

installClientPackages=${AZURELUSTRE_CSI_INSTALL_LUSTRE_CLIENT:-yes}
echo "installClientPackages: ${installClientPackages}"

requiredLustreVersion=${LUSTRE_VERSION:-"2.15.5"}
echo "requiredLustreVersion: ${requiredLustreVersion}"

requiredClientSha=${CLIENT_SHA_SUFFIX:-"41-gc010524"}
echo "requiredClientSha: ${requiredClientSha}"

pkgVersion="${requiredLustreVersion}-${requiredClientSha}"
echo "pkgVersion: ${pkgVersion}"

pkgName="amlfs-lustre-client-${pkgVersion}"
echo "pkgName: ${pkgName}"

if [[ ! -z $(grep -R 'bionic' /etc/os-release) ]]; then
  osReleaseCodeName="bionic"
elif [[ ! -z $(grep -R 'jammy' /etc/os-release) ]]; then
  osReleaseCodeName="jammy"
elif [[ ! -z $(grep -R 'focal' /etc/os-release) ]]; then
  osReleaseCodeName="focal"
else
  echo "Unsupported Linux distro"
  exit 1
fi

echo "$(date -u) Command line arguments: $@"

if [[ "${installClientPackages}" == "yes" ]]; then
  kernelVersion=$(uname -r)

  echo "$(date -u) Installing Lustre client packages for OS=${osReleaseCodeName}, kernel=${kernelVersion} "

  if [ ! -f /etc/apt/sources.list.d/amlfs.list ] ||  ! ls /var/lib/apt/lists  | grep "packages.microsoft.com_repos_amlfs" &> /dev/null; then
    curl -sL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | tee /etc/apt/trusted.gpg.d/microsoft.gpg > /dev/null
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/amlfs-${osReleaseCodeName}/ ${osReleaseCodeName} main" | tee /etc/apt/sources.list.d/amlfs.list
    apt-get update
  fi

  echo "$(date -u) Installing Lustre client modules: ${pkgName}=${kernelVersion}"

  tries=3
  sleep_before_retry=15
  install_success=false
  while [[ tries -gt 0 ]]; do
    # grub issue
    # https://stackoverflow.com/questions/40748363/virtual-machine-apt-get-grub-issue/40751712
    if ! DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends -o DPkg::options::="--force-confdef" -o DPkg::options::="--force-confold" \
      ${pkgName}=${kernelVersion}; then
      echo "$(date -u) Error installing Lustre client modules. Will try removing existing versions"
      # Check if lustre_rmmod is available, attempt to unload the modules if so.
      # If modules are already uninstalled, this will still pass
      if type lustre_rmmod >/dev/null 2>&1 && ! lustre_rmmod; then
        echo "$(date -u) Error: Unable to unload running module. Are there still mounted Lustre filesystems on this node? Old Lustre client version may continue running."
      fi
      if existing_versions=$(dpkg-query --showformat=' ${Package}=${Version}' --show '*lustre-client*'); then
        echo  "$(date -u) The following existing versions of the Lustre client are installed and will be removed:${existing_versions}"
      fi
      echo "$(date -u) Uninstalling existing Lustre client versions."
      apt-get remove --purge -y '*lustre-client*' || true
      tries=$((tries - 1))
      sleep $sleep_before_retry
      sleep_before_retry=$((sleep_before_retry * 2))
    else
      install_success=true
      break
    fi
  done

  echo "$(date -u) Install success: ${install_success}, Tries left: ${tries}"

  if ! ${install_success}; then
    echo "$(date -u) Error: Could not install necessary Lustre drivers for: ${pkgName}=${kernelVersion}"
  else
    echo "$(date -u) Installed Lustre client packages for: ${pkgName}=${kernelVersion}"
  fi

  init_lnet="true"

  if lsmod | grep "^lnet"; then
    if lnetctl net show --net tcp | grep interfaces; then
      echo "$(date -u) LNet is loaded skip the load"
      echo "$(date -u) Adding missing interfaces"
      add_net_interfaces
      init_lnet="false"
    elif lnetctl net show | grep "net type: tcp"; then
    # There may be a default configuration with no interface.
    # This is configured by an old version CSI.
      lnetctl net del --net tcp
    fi
  fi

  if [[ "${init_lnet}" == "true" ]]; then
    echo "$(date -u) Loading the LNet."
    modprobe -v lnet
    modprobe -v ksocklnd skip_mr_route_setup=1
    lnetctl lnet configure

    add_net_interfaces

    echo "$(date -u) Done"
  fi

  echo "$(date -u) Enabling Lustre client kernel modules."
  modprobe -v mgc
  modprobe -v lustre

  echo "$(date -u) Enabled Lustre client kernel modules."

fi

echo "$(date -u) Entering Lustre CSI driver"

echo Executing: $1 ${2-} ${3-} ${4-} ${5-} ${6-} ${7-} ${8-} ${9-}
$1 ${2-} ${3-} ${4-} ${5-} ${6-} ${7-} ${8-} ${9-}

echo "$(date -u) Exiting Lustre CSI driver"
