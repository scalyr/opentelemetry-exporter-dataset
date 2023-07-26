#!/usr/bin/env sh
set -e

OUTAGE_TIME="${1:-100}"
DRAIN_TIME="${2:-50}"
SLEEP_DELAY="${3:-"0.3"}"

STATS_FILE="/tmp/api-stats.json"
echo "Server will be out for ${OUTAGE_TIME} and then available for ${DRAIN_TIME}";

sleep 5;
# set status to 530, so it's failing
curl -v -m 5 -X GET "http://dataset-dummy-server-rel:8000/params/?http_code=530"

# delete previous stats
rm -rfv ${STATS_FILE};

# keep it failing for some time
for i in $( seq "${OUTAGE_TIME}" ); do
  echo "OUT: ${i} / ${OUTAGE_TIME}:  $(date)";
  curl -s -m 1 -X GET "http://dataset-dummy-server-rel:8000/stats" | tee -a ${STATS_FILE};
  echo "" >> ${STATS_FILE};
  echo ""
  ls -l "${STATS_FILE}"
  echo ""
  sleep "${SLEEP_DELAY}"
done;

# set status to 200
curl -v -m 5 -X GET "http://dataset-dummy-server-rel:8000/params/?http_code=200"

# drain remaining messages
for i in $( seq "${DRAIN_TIME}" ); do
  echo "DRAIN: ${i} / ${DRAIN_TIME}: $(date)";
  curl -s -m 1 -X GET "http://dataset-dummy-server-rel:8000/stats" | tee -a ${STATS_FILE};
  echo "" >> ${STATS_FILE};
  echo ""
  ls -l "${STATS_FILE}"
  echo ""
  sleep "${SLEEP_DELAY}";
done;

exit 0;
