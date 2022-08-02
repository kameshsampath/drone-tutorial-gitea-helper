#!/usr/bin/env bash

#
# Copyright © 2022  Kamesh Sampath <kamesh.sampath@hotmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
#  limitations under the License.
#

# Documentation about this script and how to use it can be found
# at https://github.com/go-kluster/hack

source $(dirname $0)/build-flags.sh

# Dir where this script is located
basedir() {
    # Default is current directory
    local script=${BASH_SOURCE[0]}

    # Resolve symbolic links
    if [ -L $script ]; then
        if readlink -f $script >/dev/null 2>&1; then
            script=$(readlink -f $script)
        elif readlink $script >/dev/null 2>&1; then
            script=$(readlink $script)
        elif realpath $script >/dev/null 2>&1; then
            script=$(realpath $script)
        else
            echo "ERROR: Cannot resolve symbolic link $script"
            exit 1
        fi
    fi

    local dir full_dir
    dir=$(dirname "$script")
    full_dir=$(cd "${dir}/.." && pwd)
    echo "${full_dir}"
}

function cross_platform() {
  local basedir ld_flags
  basedir=$(basedir)
  ld_flags="$(build_flags $basedir)"

  export CGO_ENABLED=0

  echo "🚧 🐧 Building for Linux (amd64)"
  GOOS=linux GOARCH=amd64 go build -ldflags "${ld_flags}"  -mod=vendor -o ./kluster-linux-amd64 ./cmd/...
  echo "🚧 💪 Building for Linux (arm64)"
  GOOS=linux GOARCH=arm64 go build -ldflags "${ld_flags}"  -mod=vendor -o ./kluster-linux-arm64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -ldflags "${ld_flags}"  -mod=vendor -o ./kluster-darwin-amd64 ./cmd/...
  echo "🚧 🍎 Building for macOS (arm64)"
  GOOS=darwin GOARCH=arm64 go build -ldflags "${ld_flags}"  -mod=vendor -o ./kluster-darwin-arm64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -ldflags "${ld_flags}"  -mod=vendor -o ./kluster-windows-amd64.exe ./cmd/...
  ARTIFACTS_TO_PUBLISH="kluster-darwin-amd64 kluster-darwin-arm64 kluster-linux-amd64 kluster-linux-arm64 kluster-windows-amd64.exe"
  sha256sum ${ARTIFACTS_TO_PUBLISH} > checksums.txt
  ARTIFACTS_TO_PUBLISH="${ARTIFACTS_TO_PUBLISH} checksums.txt"
  echo "🧮     Checksum:"
  cat checksums.txt
}

cross_platform "$@"