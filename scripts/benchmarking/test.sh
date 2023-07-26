#!/usr/bin/env bash
set -e

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
# shellcheck disable=SC1090,SC1091
source "${SCRIPT_DIR}/../common.sh"

enable_debug_mode_if_set

source "${SCRIPT_DIR}/../check-for-gnu-tools.sh"

TEST_RUN_SERVERHOST=${TEST_RUN_SERVERHOST:-"otel-cicd-benchmarks-e2e-$(date '+%s')"}

echo "Using TEST_RUN_SERVERHOST=${TEST_RUN_SERVERHOST}"

# We need to ensure we use unique host for each test run so we can differentiate between
# different runs and assert on each run data.
export TEST_RUN_SERVERHOST

function cleanup() {
  EXIT_CODE=$?
  echo "Cleaning up..."
  docker compose -f docker-compose.yaml down;

  exit ${EXIT_CODE}
}

trap "cleanup" EXIT

# stop previous runs
log "Stop previous instances"
docker compose -f docker-compose.yaml down

log "Start new instances"
docker compose -f docker-compose.yaml up --abort-on-container-exit
