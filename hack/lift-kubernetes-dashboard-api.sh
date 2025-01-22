#!/usr/bin/env bash
# Copyright 2022 The Karmada Authors.
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
REPO_ROOT=$(cd ../ && pwd)


#BRANCH_NAME=release/7.10.1
#git clone --depth=1 --branch ${BRANCH_NAME} git@github.com:kubernetes/dashboard.git ${REPO_ROOT}/tmp

mv ${REPO_ROOT}/tmp/modules/api/ ${REPO_ROOT}/cmd/kubernetes-dashboard-api
mv ${REPO_ROOT}/tmp/modules/common ${REPO_ROOT}/pkg/kubernetes-dashboard-common

rm -rf tmp