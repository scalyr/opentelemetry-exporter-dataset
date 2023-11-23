#!/usr/bin/env bash
set -e

SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")
# shellcheck disable=SC1090,SC1091
source "${SCRIPT_DIR}/../common.sh"

enable_debug_mode_if_set

source "${SCRIPT_DIR}/../check-for-gnu-tools.sh"

NUM_FILES=4
NUM_LINES=200000
LINES_MU=150
LINES_SIGMA=20
SEED=42
CHECK_LIMIT=300

while [[ $# -gt 0 ]]; do
  case $1 in
    -f|--files|--num-files)
      NUM_FILES="$2"
      shift # past argument
      shift # past value
      ;;
    -l|--lines|--num-lines)
      NUM_LINES="$2"
      shift # past argument
      shift # past value
      ;;
    -m|--mu|--lines-mu|--line-length-mu)
      LINES_MU="$2"
      shift # past argument
      shift # past value
      ;;
    -s|--sigma|--lines-sigma|--line-length-sigma)
      LINES_SIGMA="$2"
      shift # past argument
      shift # past value
      ;;
    -r|--random|--seed)
      SEED="$2"
      shift # past argument
      shift # past value
      ;;
    -u|--server)
      DATASET_URL="$2"
      shift # past argument
      shift # past value
      ;;
    -c|--check-limit)
      CHECK_LIMIT="$2"
      shift # past argument
      shift # past value
      ;;

    -h|--help)
      echo "Script for running tests with"
      exit 0
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

function cleanup() {
  EXIT_CODE=$?
  echo "Cleaning up..."
  docker compose -f docker-compose.yaml down

  exit ${EXIT_CODE}
}

trap "cleanup" EXIT

log "Run tests with - NUM_FILES: ${NUM_FILES}; NUM_LINES: ${NUM_LINES}; LINES_MU: ${LINES_MU}; LINES_SIGMA: ${LINES_SIGMA}; SEED: ${SEED}; DATASET_URL: ${DATASET_URL}; CHECK_LIMIT: ${CHECK_LIMIT}";

if [ "x${DATASET_URL}" == "x" ]; then
  log "Using dummy server as fallback";
  export DATASET_URL="http://dataset-dummy-server-e2e:8000";
  export DATASET_API_KEY="dummy";
fi;
log "DATASET_URL: '${DATASET_URL}'";
log "DATASET_API_KEY: '${DATASET_API_KEY: -4}' (last 4 characters)";
log "TEST_RUN_SERVERHOST: '${TEST_RUN_SERVERHOST}'"

if [ "x${DATASET_API_KEY}" == "x" ]; then
  log "Environmental variable DATASET_API_KEY is not set!";
  exit 1;
fi;

# stop previous runs
docker compose -f docker-compose.yaml down

LOG_FILE_PREFIX=log.log
LOG_FILE_USED=${LOG_FILE_PREFIX}.used

# delete used files
rm -rfv ${LOG_FILE_USED}.*

# prepare log file
LOG_FILE="${LOG_FILE_PREFIX}.NUM_LINES.${NUM_LINES}.LINES_MU.${LINES_MU}.LINES_SIGMA.${LINES_SIGMA}.SEED.${SEED}"
log "Using log file: ${LOG_FILE}";

# generate log file if it's not there
if [ ! -f "${LOG_FILE}" ]; then
  log "Generate file ${LOG_FILE}";
  docker run --rm logs-generator \
    --lines="${NUM_LINES}" \
    --line-length-mu="${LINES_MU}" \
    --line-length-sigma="${LINES_SIGMA}" \
    --seed="${SEED}" > ${LOG_FILE};
else
  log "No need to generate file ${LOG_FILE}";
fi;

# copy generated file into multiple instances
for i in $(seq 1 "${NUM_FILES}"); do
  logFile=${LOG_FILE_USED}.${i}
  # we cannot just copy them
  # filelogreceiver is fingerprinting them, so they have to differ
  # lets include current time in ns at the beginning
  log "Copy ${LOG_FILE} into ${logFile}";
  gdate -Ins 2> /dev/null || date -Ins > "${logFile}";
  cat "${LOG_FILE}" >> "${logFile}"
done;

# get the log size
LOG_SIZE=$( (gdu -b ${LOG_FILE_USED}.* || du -b ${LOG_FILE_USED}.*) | cut -f1 | (gpaste -s -d+ || paste -s -d+) | bc -l )

# prepare stats files
STATS_FILE=stats-line.txt
STATS_FILE_COL_EV=stats-col-events.json
STATS_FILE_COL_BUF=stats-col-buffer.json
STATS_FILE_COL_TR=stats-col-transfer.json
STATS_DOCKER=stats-docker.txt
log "Using stats file: ${STATS_FILE_COL_EV}, ${STATS_FILE_COL_BUF}, ${STATS_FILE_COL_TR}, ${STATS_FILE}, ${STATS_DOCKER}";
rm -rfv ${STATS_FILE_COL_EV} ${STATS_FILE_COL_BUF} ${STATS_FILE_COL_TR} ${STATS_FILE} ${STATS_DOCKER};
touch ${STATS_FILE_COL_EV} ${STATS_FILE_COL_BUF} ${STATS_FILE_COL_TR} ${STATS_FILE} ${STATS_DOCKER};


# start exporter and dummy server
log "Start instances"
startTime=$(date +"%s");
# you can try: --abort-on-container-exit
docker compose -f docker-compose.yaml up --detach --remove-orphans;

log "Read collector logs"
COLLECTOR_LOGS=collector-logs.log;
docker logs -n 20 -f dataset-collector-e2e 2>&1 | tee ${COLLECTOR_LOGS} &

function extract_collector_stats() {
    grep "Events' Queue Stats" ${COLLECTOR_LOGS} | grep '"data_type": "logs"' | tail -n1 | sed -r 's/.*(\{.*)/\1/' | tee ${STATS_FILE_COL_EV};
    grep "Buffers' Queue Stats" ${COLLECTOR_LOGS} | grep '"data_type": "logs"' | tail -n1 | sed -r 's/.*(\{.*)/\1/' | tee ${STATS_FILE_COL_BUF};
    grep "Transfer Stats" ${COLLECTOR_LOGS} | grep '"data_type": "logs"' | tail -n1 | sed -r 's/.*(\{.*)/\1/' | tee ${STATS_FILE_COL_TR};
    echo;
}

# check how large it is
EVENTS_EXPECTED=$( cat ${LOG_FILE_USED}.* | wc -l | tr -d " ");
log "Expected lines: ${EVENTS_EXPECTED}";

# we have to declare array so that we can pass values between functions
declare -a STATS
ST_PROCESSING=0
ST_INPUT_SIZE=$(( ST_PROCESSING + 1 ))
ST_EVENTS_EXPECTED=$(( ST_INPUT_SIZE + 1 ))
ST_EVENTS_PROCESSED=$(( ST_EVENTS_EXPECTED + 1 ))
ST_EVENTS_ENQUEUED=$(( ST_EVENTS_PROCESSED + 1 ))
ST_EVENTS_WAITING=$(( ST_EVENTS_ENQUEUED + 1 ))
ST_BUFFERS_PROCESSED=$(( ST_EVENTS_WAITING + 1 ))
ST_BUFFERS_ENQUEUED=$(( ST_BUFFERS_PROCESSED + 1 ))
ST_BUFFERS_WAITING=$(( ST_BUFFERS_ENQUEUED + 1 ))
ST_BUFFERS_DROPPED=$(( ST_BUFFERS_WAITING + 1 ))
ST_BYTES_SENT_MB=$(( ST_BUFFERS_DROPPED + 1 ))
ST_BYTES_ACCEPTED_MB=$(( ST_BYTES_SENT_MB + 1 ))
ST_BYTES_PER_BUFFER_MB=$(( ST_BYTES_ACCEPTED_MB + 1 ))
ST_OUT_THROUGHPUT_MBPS=$(( ST_BYTES_PER_BUFFER_MB + 1 ))
ST_SUCCESS_RATE=$(( ST_OUT_THROUGHPUT_MBPS + 1 ))
ST_IN_THROUGHPUT_MBPS=$(( ST_SUCCESS_RATE + 1 ))
ST_IN_THROUGHPUT_EVENTSPS=$(( ST_IN_THROUGHPUT_MBPS + 1 ))


# set fields
STATS[$ST_INPUT_SIZE]=${LOG_SIZE}
STATS[$ST_EVENTS_EXPECTED]=${EVENTS_EXPECTED}

function debug_print_stats_array() {
    printf '%s\n' "${STATS[@]}"
}


function compute_stats() {
  if [ -s ${STATS_FILE_COL_EV} ] && [ -s ${STATS_FILE_COL_BUF} ] && [ -s ${STATS_FILE_COL_TR} ]; then
    STATS[${ST_PROCESSING}]=$( jq '.processingS' "${STATS_FILE_COL_EV}" );

    STATS[${ST_EVENTS_PROCESSED}]=$( jq '.processed' "${STATS_FILE_COL_EV}" );
    STATS[${ST_EVENTS_ENQUEUED}]=$( jq '.enqueued' "${STATS_FILE_COL_EV}" );
    STATS[${ST_EVENTS_WAITING}]=$( jq '.waiting' "${STATS_FILE_COL_EV}" );

    STATS[${ST_BUFFERS_PROCESSED}]=$( jq '.processed' "${STATS_FILE_COL_BUF}" );
    STATS[${ST_BUFFERS_ENQUEUED}]=$( jq '.enqueued' "${STATS_FILE_COL_BUF}" );
    STATS[${ST_BUFFERS_WAITING}]=$( jq '.waiting' "${STATS_FILE_COL_BUF}" );
    STATS[${ST_BUFFERS_DROPPED}]=$( jq '.dropped' "${STATS_FILE_COL_BUF}" );

    STATS[${ST_BYTES_SENT_MB}]=$( jq '.bytesSentMB' "${STATS_FILE_COL_TR}" );
    STATS[${ST_BYTES_ACCEPTED_MB}]=$( jq '.bytesSentMB' "${STATS_FILE_COL_TR}" );
    STATS[${ST_BYTES_PER_BUFFER_MB}]=$( jq '.perBufferMB' "${STATS_FILE_COL_TR}" );
    STATS[${ST_OUT_THROUGHPUT_MBPS}]=$( jq '.throughputMBpS' "${STATS_FILE_COL_TR}" );
    STATS[${ST_SUCCESS_RATE}]=$( jq '.successRate' "${STATS_FILE_COL_TR}" );

    approxProcessed=$( echo "${STATS[$ST_INPUT_SIZE]} * ${STATS[${ST_EVENTS_PROCESSED}]} / ${STATS[${ST_EVENTS_EXPECTED}]}" | bc -l );
    STATS[${ST_IN_THROUGHPUT_MBPS}]=$( echo "${approxProcessed} / ${STATS[${ST_PROCESSING}]} / (1024 * 1024)" | bc -l );
    STATS[${ST_IN_THROUGHPUT_EVENTSPS}]=$( echo "${STATS[${ST_EVENTS_PROCESSED}]} / ${STATS[${ST_PROCESSING}]}" | bc -l );

    sizeMB=$( echo "${STATS[$ST_INPUT_SIZE]} / (1024 * 1024)" | bc -l );
    approxProcessedMB=$( echo "${approxProcessed} / (1024 * 1024)" | bc -l );

    processedER=$( echo "${STATS[${ST_EVENTS_PROCESSED}]} / ${STATS[${ST_EVENTS_EXPECTED}]}" | bc -l )
    processedMBR=$( echo "${approxProcessedMB} / ${sizeMB}" | bc -l )

    # print final counts
    # size of the generated input log files in MB
    log "Size (MB): ${sizeMB}";
    # for how many seconds was the exporter processing the data - from the moment when it received first data to the moment when the last data were accepted by the server
    log "Duration (s): ${STATS[${ST_PROCESSING}]}";
    # how many events have been accepted by the server
    log "Processed (events): ${STATS[${ST_EVENTS_PROCESSED}]} / ${STATS[${ST_EVENTS_EXPECTED}]} (${processedER})";
    # estimate how many MB of input logs has been processed, it's based on the input log size, number of processed events, and total number of events
    log "Processed (MB): ${approxProcessedMB} / ${sizeMB} (${processedMBR})";
    # estimated input throughput measured in the input data size (based on Processed (MB), Duration (s))
    log "Throughput - IN (MB / s): ${STATS[${ST_IN_THROUGHPUT_MBPS}]}";
    # estimated input throughput measured in the input data size (based on Processed (events), Duration (s))
    log "Throughput - IN (events / s): ${STATS[${ST_IN_THROUGHPUT_EVENTSPS}]}";
    # estimated output throughput measured in the output data size (based on Produced - Accepted (MB), Duration (s))
    log "Throughput - OUT (MB / s): ${STATS[${ST_OUT_THROUGHPUT_MBPS}]}";
    # amount of accepted raw data by the server, without compression, if the payload is accepted after retry, then it's size is counted once
    log "Produced - Accepted (MB): ${STATS[${ST_BYTES_ACCEPTED_MB}]}";
    # amount of raw data sent to the server, without compression, if the payload is accepted after retry, then it's size is counted twice
    log "Produced - Sent (MB): ${STATS[${ST_BYTES_SENT_MB}]}";
    # number of events that has been added to the buffers
    log "Events - processed: ${STATS[${ST_EVENTS_PROCESSED}]}";
    # number of events that has been accepted by the library
    log "Events - enqueued: ${STATS[${ST_EVENTS_ENQUEUED}]}";
    # number of events being processed by the library (enqueued - processed)
    log "Events - waiting: ${STATS[${ST_EVENTS_WAITING}]}";
    # number of buffers that has been accepted by the server
    log "Buffers - processed: ${STATS[${ST_BUFFERS_PROCESSED}]}";
    # number of buffers that has been created and submitted for further processing
    log "Buffers - enqueued: ${STATS[${ST_BUFFERS_ENQUEUED}]}";
    # number of buffers that are being processed (enqueued - processed)
    log "Buffers - waiting: ${STATS[${ST_BUFFERS_WAITING}]}";
    # number of buffers that were dropped after all the retries have failed
    log "Buffers - dropped: ${STATS[${ST_BUFFERS_DROPPED}]}";
    log "Server: ${STATS[${ST_SERVER}]}";
    echo;

    echo "${STATS[@]}" > ${STATS_FILE};
  fi;
}

# wait until it's processed
for i in $( seq "${CHECK_LIMIT}" ); do
  if [ "$( docker ps | grep -c dataset-collector-e2e )" == 0 ]; then
    log "Container dataset-collector-e2e is dead => exiting";
    exit 1;
  fi;

  extract_collector_stats

  log "Docker stats:"
  docker stats --no-stream | tee ${STATS_DOCKER};
  echo;

  log "Check: ${i} / ${CHECK_LIMIT} ";
  compute_stats;
  if [ -s ${STATS_FILE} ]; then
    read -r -a STATS < ${STATS_FILE};
    eventsProcessed=${STATS[ST_EVENTS_PROCESSED]};
    log "Processed: ${eventsProcessed} / ${EVENTS_EXPECTED}";
    if [ "${eventsProcessed}" -ge "${EVENTS_EXPECTED}" ] && [ "${STATS[${ST_BUFFERS_WAITING}]}" == 0 ]; then
      break;
    fi;
  fi;

  sleep 10;
done;

endTime=$(date +"%s");

# get latest statistics
docker stats --no-stream | tee ${STATS_DOCKER};

# extract information about the source code
LIB_DIR=$(pwd)/../../dataset-go;
if [ -d "${LIB_DIR}" ]; then
  LIB_HASH=$( cd "${LIB_DIR}"; git rev-parse HEAD  );
  LIB_BRANCH=$( cd "${LIB_DIR}"; git branch --show-current );
fi;

OTEL_DIR=$(pwd)/../../scalyr-opentelemetry-collector-contrib;
if [ -d "${OTEL_DIR}" ]; then
  OTEL_HASH=$( cd "${OTEL_DIR}"; git rev-parse HEAD  );
  OTEL_BRANCH=$( cd "${OTEL_DIR}"; git branch --show-current );
fi;

THIS_HASH=$( git rev-parse HEAD  );
THIS_BRANCH=$( git branch --show-current );

# echo "${LIB_BRANCH}; ${LIB_HASH}; ${OTEL_HASH}; ${OTEL_BRANCH}; ${THIS_BRANCH}; ${THIS_HASH}";

# extract information about used files
DOCKER_COMPOSE_HASH=$( md5sum docker-compose.yaml )
CONFIG_HASH=$( md5sum otel-config.yaml )
SCRIPT_HASH=$( md5sum test.sh )

# echo "${DOCKER_COMPOSE_HASH}; ${CONFIG_HASH}; ${SCRIPT_HASH};";

# extract system information
UNAME=$( uname -a );
HOSTNAME=$( hostname );
IP_INFO=$( dig +short myip.opendns.com @resolver1.opendns.com 2> /dev/null || curl https://ipinfo.io/ip )
CPU_INFO="\"\"";
if [ "$( lscpu -J )" ]; then
  CPU_INFO=$( lscpu -J | paste -s );
elif [ "$( sysctl -a machdep )" ]; then
  CPU_INFO=$( sysctl -a machdep | paste -s -d, );
  CPU_INFO="\"${CPU_INFO}\"";
fi;
CONTAINER_INFO=$( docker container inspect dataset-collector-e2e | jq '{Image: .[].Image, Platform: .[].Platform, Memory: .[].HostConfig.Memory, NanoCpus: .[].HostConfig.NanoCpus}' | paste -s )

# echo "${UNAME}; ${HOSTNAME}; ${IP_INFO}; ${CPU_INFO}; ${CONTAINER_INFO}"

# stop everything
log "Stop instances";
docker compose -f docker-compose.yaml down;

log "Compute final stats";
extract_collector_stats;
compute_stats;
read -r -a STATS < ${STATS_FILE};
eventsProcessed=${STATS[ST_EVENTS_PROCESSED]};

log "Processed: ${eventsProcessed} / ${EVENTS_EXPECTED}";

benchmarkLine="{
  \"key\": \"value\",
  \"time\": \"$( gdate -Is 2> /dev/null || date -Is )\",
  \"serverURL\": \"${DATASET_URL}\",
  \"serverApiKey\": \"${DATASET_API_KEY: -4}\",
  \"numFiles\": ${NUM_FILES},
  \"numLines\": ${NUM_LINES},
  \"linesMu\": ${LINES_MU},
  \"linesSigma\": ${LINES_SIGMA},
  \"seed\": ${SEED},

  \"processingTimeS\": ${STATS[${ST_PROCESSING}]},
  \"inputSizeB\": ${STATS[${ST_INPUT_SIZE}]},
  \"eventsExpected\": ${STATS[${ST_EVENTS_EXPECTED}]},
  \"eventsProcessed\": ${STATS[${ST_EVENTS_PROCESSED}]},
  \"eventsEnqueued\": ${STATS[${ST_EVENTS_ENQUEUED}]},
  \"eventsWaiting\": ${STATS[${ST_EVENTS_WAITING}]},
  \"buffersProcessed\": ${STATS[${ST_BUFFERS_PROCESSED}]},
  \"buffersEnqueued\": ${STATS[${ST_BUFFERS_ENQUEUED}]},
  \"buffersWaiting\": ${STATS[${ST_BUFFERS_WAITING}]},
  \"buffersDropped\": ${STATS[${ST_BUFFERS_DROPPED}]},
  \"bytesSentMB\": ${STATS[${ST_BYTES_SENT_MB}]},
  \"bytesAcceptedMB\": ${STATS[${ST_BYTES_ACCEPTED_MB}]},
  \"bytesPerBufferMB\": ${STATS[${ST_BYTES_PER_BUFFER_MB}]},
  \"throughputOutMBpS\": ${STATS[${ST_OUT_THROUGHPUT_MBPS}]},
  \"successRate\": ${STATS[${ST_SUCCESS_RATE}]},
  \"throughputInMBpS\": ${STATS[${ST_IN_THROUGHPUT_MBPS}]},
  \"throughputInEventspS\": ${STATS[${ST_IN_THROUGHPUT_EVENTSPS}]},

  \"startTime\": ${startTime},
  \"endTime\": ${endTime},

  \"thisBranch\": \"${THIS_BRANCH}\",
  \"thisHash\": \"${THIS_HASH}\",
  \"libBranch\": \"${LIB_BRANCH}\",
  \"libHash\": \"${LIB_HASH}\",
  \"otelBranch\": \"${OTEL_BRANCH}\",
  \"otelHash\": \"${OTEL_HASH}\",


  \"dockerComposeHash\": \"${DOCKER_COMPOSE_HASH}\",
  \"otelConfigHash\": \"${CONFIG_HASH}\",
  \"scriptHash\": \"${SCRIPT_HASH}\",

  \"uname\": \"${UNAME}\",
  \"hostname\": \"${HOSTNAME}\",
  \"ipInfo\": \"${IP_INFO}\",
  \"cpuInfo\": ${CPU_INFO},
  \"containerInfo\":  ${CONTAINER_INFO},

  \"last\": \"last\"
}";

echo "${benchmarkLine}" | jq '.';

BENCHMARK_FILE=benchmark.log
echo "${benchmarkLine}" | paste -s >> "${BENCHMARK_FILE}";

# exit with non zero status if values differ
log "${EVENTS_EXPECTED}" == "${eventsProcessed}";
test "${EVENTS_EXPECTED}" == "${eventsProcessed}";
