#!/usr/bin/env bash
set -e

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
# shellcheck disable=SC1090,SC1091
source "${SCRIPT_DIR}/../common.sh"

enable_debug_mode_if_set

source "${SCRIPT_DIR}/../check-for-gnu-tools.sh"

function cleanup() {
  EXIT_CODE=$?
  echo "Cleaning up..."
  docker compose -f docker-compose.yaml down;

  exit ${EXIT_CODE}
}

# stop everything
docker compose -f docker-compose.yaml down;

# prepare empty log file
LOG_FILE=logs-10000.log;
log "Using log file: ${LOG_FILE}";
rm -rfv ${LOG_FILE};
touch ${LOG_FILE};

# Used by otel collector inside the container
mkdir -p otc

# TODO: Use syslog load gen which can generate much more realistic data
# NOTE: It's built by make docker-build step which runs as part of CI/CD step
#pushd ../../logs-generator
#docker build -t logs-generator:latest .
# popd

trap "cleanup" EXIT

# generate 10k log lines
docker run logs-generator:latest --lines 10000 2> /dev/null > ${LOG_FILE}

# start exporter and dummy server
# there is also image, that make the server unavailable first and then available
# make sure, that the combined time is greater then time needed to generate
# logs above
# NOTE: server-control.sh will exit with 0 on completion and that's intentional
set +e
docker compose -f docker-compose.yaml up --remove-orphans --abort-on-container-exit
set -e

# check how large it is
expected=$(wc -l < ${LOG_FILE} | tr -d " ");
log "Expected lines: ${expected}";

# get generated stats from API
API_STATS_FILE="./api-stats.json"
tail -n1 ${API_STATS_FILE} | jq '.';

processed=$( tail -n1 ${API_STATS_FILE} | jq '.events.total.success' )
log "Processed lines: ${processed}";

# stop everything
docker compose -f docker-compose.yaml down;

# print few empty lines to make space
for _ in $( seq 30 ); do
  echo "";
done;

# print final counts
log "FINAL: Processed: ${processed}; Expected: ${expected}";

# exit with non zero status if values differ
test "${expected}" == "${processed}";

log "Tests passed successfully!"
exit 0
