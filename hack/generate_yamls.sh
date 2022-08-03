#!/usr/bin/env bash

set -o errexit
set -o pipefail

readonly YAML_REPO_ROOT=${1:?"First argument must be the repo root dir"}
readonly YAML_OUTPUT_DIR=${2:?"Second argument must be the dist dir"}
readonly TAG=${3:?"Third argument must be the tag"}

KO_YAML_FLAGS="-B --tags=${TAG} --platform=linux/amd64,linux/arm64"

# Set output directory
if [[ -z "${YAML_OUTPUT_DIR:-}" ]]; then
  readonly YAML_OUTPUT_DIR="$(mktemp -d)"
fi

rm -fr "${YAML_OUTPUT_DIR}"/*.yaml

# Generated Utility component YAML files
readonly INSTALL_YAML=${YAML_OUTPUT_DIR}/install.yaml

readonly KO_YAML_FLAGS="${KO_YAML_FLAGS} ${KO_FLAGS}"

if [ -z "$TAG" ];
then
  TAG=$(svu next --suffix="${PRE_RELEASE_SUFFIX}")
fi

if [[ -n "${TAG}" ]]; then
  LABEL_YAML_CMD=(sed -e "s|app.kubernetes.io/version: devel|app.kubernetes.io/version: \"${TAG:1}\"|")
else
  LABEL_YAML_CMD=(cat)
fi

export KO_DOCKER_REPO="docker.io/kameshsampath"

cd "${YAML_REPO_ROOT}"

echo "Building Drone Tutorial Workshop Helper Kubernetes Manifests"
kustomize build config | ko resolve ${KO_YAML_FLAGS} -R -f - | "${LABEL_YAML_CMD[@]}" > "${INSTALL_YAML}"