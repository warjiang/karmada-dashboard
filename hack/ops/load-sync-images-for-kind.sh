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
  echo "Note: You should do login to private registry before execute this script,
        This script is an internal script and is not intended used by end-users."
  echo "Usage: hack/ops/load-sync-images-for-kind.sh <IMAGE_LIST> <KIND_CLUSTER_NAME>"
  echo "Example: hack/ops/load-sync-images-for-kind.sh hack/images/image.list karmada-host"
  echo "Parameters: "
  echo "        IMAGE_LIST               List of image descriptions you want to sync"
  echo "        KIND_CLUSTER_NAME        Cluster name which initialized by kind"
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

IMAGE_FILE_PATH=${1}
KIND_CLUSTER_NAME=${2}

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
    docker pull ${target_image}
    docker tag ${target_image} ${origin_image}
    kind load docker-image ${origin_image} --name=${KIND_CLUSTER_NAME} -v -1
  fi
done

INFO "Load all images in kind successfully."