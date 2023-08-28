# Copyright 2023 SentinelOne, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# build options are from https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/Makefile.Common
# to make out library compatible with open telemetry

SHELL = /bin/bash
ifeq ($(shell uname -s),Windows)
	.SHELLFLAGS = /e /o pipefile /c
else
	.SHELLFLAGS = -e -o pipefail -c
endif

GO_BUILD_TAGS=""
GOTEST_OPT?= -race -timeout 300s -parallel 4 -count=1 --tags=$(GO_BUILD_TAGS)
GOTEST_INTEGRATION_OPT?= -race -timeout 360s -parallel 4
GOTEST_OPT_WITH_COVERAGE = $(GOTEST_OPT) -coverprofile=coverage.txt -covermode=atomic
GOTEST_OPT_WITH_INTEGRATION=$(GOTEST_INTEGRATION_OPT) -tags=integration,$(GO_BUILD_TAGS) -run=Integration -coverprofile=integration-coverage.txt -covermode=atomic
GOCMD?= go
GOTEST=$(GOCMD) test

DIR_EXPORTER=scalyr-opentelemetry-collector-contrib/exporter/datasetexporter/

.PHONY: docker-build
docker-build: exporter-docker-build dummy-server-build logs-generator-build

.PHONY: exporter-docker-build
exporter-docker-build:
	docker build -t otelcol-dataset .
	# when you are modifying the source code you should build it without cache
	# docker build --no-cache -t otelcol-dataset .

.PHONY: exporter-docker-run
exporter-docker-run:
	docker run \
      -v $$( pwd )/config.yaml:/config.yaml \
      -p 4318:4318 \
      -it \
      --rm \
      --env DATASET_URL \
      --env DATASET_API_KEY \
      --name otelcol-dataset \
      otelcol-dataset --config /config.yaml

.PHONY: exporter-docker-official-run
exporter-docker-official-run:
	VERSION=latest; \
	docker pull otel/opentelemetry-collector-contrib:$${VERSION}; \
	docker run \
      -v $$( pwd )/config.yaml:/config.yaml \
       -p 4318:4318 \
       -it \
       --rm \
       --env DATASET_URL \
       --env DATASET_API_KEY \
       --name otelcol-dataset-official \
       otel/opentelemetry-collector-contrib:$${VERSION} --config /config.yaml

.PHONY: exporter-normal-build
exporter-normal-build:
	builder --config=otelcol-builder.yaml

.PHONY: exporter-normal-run
exporter-normal-run:
	if [ $$( ls -1 scripts/e2e/ | grep -c flog.log ) == 0 ]; then \
  		echo "You should generate some test files."; \
  		echo "Run: cd scripts/e2e; ./test.sh"; \
  		exit 1; \
  	  fi; \
	./otelcol-dataset/otelcol-dataset --config config.yaml

.PHONY: dummy-server-build
dummy-server-build:
	cd dummy-server; \
	make docker-build

.PHONY: dummy-server-run
dummy-server-run:
	cd dummy-server; \
	make docker-run

.PHONY: logs-generator-build
logs-generator-build:
	cd logs-generator; \
	make docker-build

.PHONY: logs-generator-run
logs-generator-run:
	cd logs-generator; \
	make docker-run


.PHONY: test
test: test-exporter

# for now our module is not included there => it has no effect
# in the future we have to modify go.mod to point to our local folder
.PHONY: test-otel-collector-contrib
test-otel-collector-contrib:
	cd opentelemetry-collector-contrib/cmd/otelcontribcol; \
	make test


.PHONY: test-exporter
test-exporter:
	cd $(DIR_EXPORTER); \
	$(GOTEST) $(GOTEST_OPT) ./... && \
	$(GOTEST) $(GOTEST_OPT_WITH_INTEGRATION) ./...

.PHONY: test-many-times
test-many-times:
	set -e; \
	if [ "x$(COUNT)" == "x" ]; then \
  		COUNT=50; \
  	else \
  		COUNT=$(COUNT); \
  	fi; \
  	for i in `seq 1 $${COUNT}`; do \
  		echo "Running test $${i} / $${COUNT}"; \
  		make test 2>&1 | tee out-test-$${i}.log; \
  	done;

.PHONY: coverage
coverage: coverage-exporter

.PHONY: coverage-exporter
coverage-exporter:
	cd $(DIR_EXPORTER); \
	$(GOTEST) $(GOTEST_OPT_WITH_COVERAGE) ./...; \
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html;


# https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
# https://formulae.brew.sh/formula/coreutils - expects, that you have install coreutils and they are in your PATH
.PHONY: push-log
# NOTE: BSD date on OSX doesn't support %N modified so we fall back to gnu date (if available)
push-log:
	./scripts/check-for-gnu-tools.sh || exit 1;          \
     if [ "x$(MSG)" = "x" ]; then                            \
      echo "Use make push-log MSG='some msg'";               \
      exit 1;                                                \
    fi;                                                      \
    if [ "x$(TS)" = "x" ]; then                              \
      TS=`gdate "+%s%N" 2> /dev/null || date "+%s%N"`;       \
    else                                                     \
      TS="$(TS)";                                            \
    fi;                                                      \
    S=`date "+%S"`;                                          \
    echo "Using: MSG=$(MSG); TS: $${TS}; S: $${S}";          \
    curl -i http://127.0.0.1:4318/v1/logs -H                 \
      'Content-Type: application/json'                       \
      -d '                                                   \
    {                                                        \
    "resourceLogs": [                                        \
      {                                                      \
        "resource": {                                        \
            "attributes": [                                  \
              {                                              \
                "key": "dummy",                              \
                "value": {                                   \
                  "stringValue": "foo"                       \
                }                                            \
              }                                              \
            ]                                                \
        },                                                   \
        "scopeLogs": [                                       \
          {                                                  \
            "logRecords": [                                  \
              {                                              \
                "timeUnixNano": "'$${TS}'",                  \
                "body": {                                    \
                  "stringValue": "$(MSG) - '$${TS}'"         \
                },                                           \
                "attributes": [                              \
                  {                                          \
                    "key": "app",                            \
                    "value": {                               \
                      "stringValue": "server"                \
                    }                                        \
                  },                                         \
                  {                                          \
                    "key": "instance_num",                   \
                    "value": {                               \
                      "intValue": "1"                        \
                    }                                        \
                  },                                         \
                  {                                          \
                    "key": "container_id",                   \
                    "value": {                               \
                      "stringValue": "cont_'$${S}'"          \
                    }                                        \
                  }                                          \
                ]                                            \
              }                                              \
            ]                                                \
          }                                                  \
        ]                                                    \
      }                                                      \
    ]                                                        \
    }'

.PHONY: push-logs
push-logs:
	echo "Use make push-logs MSG='message' SLEEP=4 COUNT=10 for better control"
	if [ "x$(MSG)" = "x" ]; then                             \
      echo "Use make push-logs MSG='some msg'";               \
      exit 1;                                                \
    fi; \
	if [ "x$(COUNT)" == "x" ]; then \
		COUNT=30; \
	else \
		COUNT=$(COUNT); \
	fi; \
	if [ "x$(SLEEP)" == "x" ]; then \
		SLEEP=10; \
	else \
		SLEEP=$(SLEEP); \
	fi; \
	for i in `seq 1 $${COUNT}`; do              \
	  make push-log MSG="$(MSG) - $${i}"; \
	  sleep "$${SLEEP}";                   \
	done;

.PHONY: push-linked-trace-with-logs
push-linked-trace-with-logs:
	./scripts/check-for-gnu-tools.sh || exit 1;          \
	TS=`gdate "+%s%N" 2> /dev/null || date "+%s%N"`; \
	REP=`echo $${TS} | cut -c 1-9`; \
	sed -r "s/111111111/$${REP}/" data/logs.json > logs.json; \
	sed -r "s/111111111/$${REP}/" data/traces.json > traces.json; \
	curl -X POST -H "Content-Type: application/json" -d @logs.json -i localhost:4318/v1/logs; \
	curl -X POST -H "Content-Type: application/json" -d @traces.json -i localhost:4318/v1/traces; \
	echo; \
	echo "Search for traceId: "; \
	grep -H traceId logs.json traces.json; \
	echo "Search for spanId: "; \
	grep -H spanId logs.json traces.json; \
	echo "Search for : $${REP}";

.PHONY: test-e2e
test-e2e:
	cd scripts/e2e; \
	./test.sh

.PHONY: subtrees-pull
subtrees-pull:
	git subtree pull --prefix dataset-go https://github.com/scalyr/dataset-go.git main --squash
	git subtree pull --prefix opentelemetry-collector-contrib https://github.com/open-telemetry/opentelemetry-collector-contrib.git main --squash
	git subtree pull --prefix opentelemetry-collector-releases https://github.com/open-telemetry/opentelemetry-collector-releases.git main --squash
	git subtree pull --prefix scalyr-opentelemetry-collector-contrib https://github.com/scalyr/opentelemetry-collector-contrib.git datasetexporter-latest --squash
