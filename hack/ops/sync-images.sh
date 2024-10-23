#!/bin/bash
# Copyright 2021 The Karmada Authors.
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

cd "$(dirname "${BASH_SOURCE[0]}")"
SHELL_FOLDER=$(pwd)
REPO_ROOT=$(cd ../../ && pwd)

source "${REPO_ROOT}"/hack/util/init.sh && util:init:init_scripts

function usage() {
  echo "This script is used to sync images from one registry to another."
  echo "Note: This script is an internal script and is not intended used by end-users."
  echo "Usage: hack/ops/sync-images.sh <IMAGE_LIST> [REGISTRY_USERNAME] [REGISTRY_PASSWORD]"
  echo "Example: hack/ops/sync-images.sh hack/images/image.list"
  echo "Parameters: "
  echo "        IMAGE_LIST               List of image descriptions you want to sync"
  echo "        REGISTRY_USERNAME        Username for the registry, optional"
  echo "        REGISTRY_PASSWORD        Password for the registry, optional"
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

IMAGE_FILE_PATH=${1}
REGISTRY_USERNAME=${2:-''}
REGISTRY_PASSWORD=${3:-''}

# unify relative path and absolute path
if [[ ${IMAGE_FILE_PATH} != /* ]]; then
  IMAGE_FILE_PATH="${REPO_ROOT}/${IMAGE_FILE_PATH}"
fi

# ensure the image_file_path exist
INFO "Image file path: ${IMAGE_FILE_PATH}"
if [ ! -f "${IMAGE_FILE_PATH}" ]; then
  ERROR "File ${IMAGE_FILE_PATH} not exits."
  exit 1
fi

export proxy_client="127.0.0.1"
export http_proxy_port="7890"
export socks5_proxy_port="7890"
function proxy() {
    if [ "$1" = "on" ]; then
        export https_proxy=http://$proxy_client:$http_proxy_port
        export http_proxy=http://$proxy_client:$http_proxy_port
        export all_proxy=socks5://$proxy_client:$socks5_proxy_port
        echo proxy on
    else
        unset https_proxy
        unset http_proxy
        unset all_proxy
        echo proxy off
    fi
}
proxy on

# shellcheck disable=SC2002
lines=$(cat "${IMAGE_FILE_PATH}" | grep -v '^\s*$' | grep -v '^#' )
for line in $lines; do
  IFS=';' read -r -a line_items <<< "${line}"
  if [ "${#line_items[@]}" -ne 3 ]; then
    WARN "Line[$line] contains ${#line_items[@]} fields, expect 3 fields, skip it"
    continue
  else
    component_name=${line_items[0]}
    origin_image=${line_items[1]}
    target_image=${line_items[2]}

    if [ -z "${origin_image}" ] || [ -z "${target_image}" ]; then
      WARN "Line[$line] contains empty origin_image or target_image, skip it"
      continue
    fi

    INFO "Start sync [${component_name}] image."
    skopeo copy --multi-arch all --dest-creds ${REGISTRY_USERNAME}:${REGISTRY_PASSWORD} docker://${origin_image}  docker://${target_image}
    INFO "Sync [${component_name}] image finished."
  fi
done

proxy off
INFO "Sync all successfully."