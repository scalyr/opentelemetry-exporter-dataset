#!/usr/bin/env bash

# build stress test
d=$(dirname "$0");
cd "${d}/../examples/stress" || exit 1

# build stress test
rm -rfv stress
go build -race -o stress
ls -l stress

EVENTS="${1:=10000}"
BUCKETS="${2:=10000}"

echo "Run stress test for ${EVENTS} events and ${BUCKETS} buckets"
./stress \
  --events="${EVENTS}" \
  --buckets="${BUCKETS}" \
  --sleep=10ms 2>&1 | tee "out-${EVENTS}-${BUCKETS}.log"
