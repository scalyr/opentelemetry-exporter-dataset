#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
# shellcheck disable=SC1090,SC1091
source "${SCRIPT_DIR}/common.sh"

enable_debug_mode_if_set

source "${SCRIPT_DIR}/check-for-gnu-tools.sh"

ROOT_DIR=$(readlink -f "${SCRIPT_DIR}/..")

FROM_VERSION=""
TO_VERSION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    -f|--from)
      FROM_VERSION="$2"
      shift # past argument
      shift # past value
      ;;
    -t|--to)
      TO_VERSION="$2"
      shift # past argument
      shift # past value
      ;;

    -h|--help)
      echo "Script for updating otel version in scripts"
      exit 0
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

function usage() {
  echo "Usage: $0 -f <from_version> -t <to_version>"
  echo "Updates version in the files from <from_version> to <to_version>"
  echo "Example: $0 -f v0.83.0 -t v0.102.1"
}

if [[ -z "${FROM_VERSION}" ]]; then
  echo "Missing from version - specify with -f"
  usage
  exit 1
fi

if [[ -z "${TO_VERSION}" ]]; then
  echo "Missing to version - specify with -t"
  usage
  exit 1
fi

log "Updating otel versions in files from ${FROM_VERSION} to ${TO_VERSION}"


function update() {
  from=${2//v/};
  to=${3//v/};
  if [ ! -f "${1}" ]; then
    echo "File ${1} does not exist"
    exit 1
  fi
  echo "Updating file ${1}: from ${from} to ${to}"
  sed -i '' "s/${from}/${to}/g" "${1}"
  git --no-pager diff "${1}"
}

update "${ROOT_DIR}/Dockerfile" "${FROM_VERSION}" "${TO_VERSION}"
update "${ROOT_DIR}/Dockerfile.contrib_datasetexporter_latest" "${FROM_VERSION}" "${TO_VERSION}"
update "${ROOT_DIR}/otelcol-builder.datasetexporter-latest.yaml" "${FROM_VERSION}" "${TO_VERSION}"
update "${ROOT_DIR}/otelcol-builder.yaml" "${FROM_VERSION}" "${TO_VERSION}"
