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
GOTEST_OPT?= -v -race -timeout 300s -parallel 4 -count=1 --tags=$(GO_BUILD_TAGS)
GOTEST_LONG_RUNNING_OPT?= -v -race -timeout 600s -parallel 4 -count=1 -tags=long_running,$(GO_BUILD_TAGS)
GOTEST_OPT_WITH_COVERAGE = $(GOTEST_OPT) -coverprofile=coverage.txt -covermode=atomic
GOTEST_OPT_WITH_COVERAGE_LONG_RUNNING=$(GOTEST_LONG_RUNNING_OPT) -coverprofile=coverage.txt -covermode=atomic
GOCMD?= go
GOTEST=$(GOCMD) test

.DEFAULT_GOAL := pre-commit-run

.PHONY: pre-commit-install
pre-commit-install:
	pre-commit install
	./scripts/install-dev-tools.sh

.PHONY: pre-commit-run
pre-commit-run:
	pre-commit run -a

build:
	echo "Done"

.PHONY: test
test: test-all

.PHONY: test-unit
test-unit:
	time $(GOTEST) $(GOTEST_OPT) ./...

.PHONY: test-all
test-all:
	time $(GOTEST) $(GOTEST_LONG_RUNNING_OPT) ./...

.PHONY: test-many-times
test-many-times:
	set -e; \
	if [ "x$(COUNT)" == "x" ]; then \
		COUNT=50; \
	else \
		COUNT=$(COUNT); \
	fi; \
	prefix=out-test-many-times-; \
	rm -rfv $${prefix}*; \
	for i in `seq 1 $${COUNT}`; do \
		echo "Running test $${i} / $${COUNT} - BEGIN"; \
		make test 2>&1 | tee $${prefix}-$${i}.log | awk '{print "'$${i}'/'$${COUNT}'", $$0; }' || exit 0; \
		echo "Running test $${i} / $${COUNT} - END"; \
	done; \
	echo "Grep for FAIL - no lines should be found"; \
	! grep -H FAIL $${prefix}-*.log;

foo:
	! false

.PHONY: coverage
coverage: coverage-all

.PHONY: coverage-unit
coverage-unit:
	$(GOTEST) $(GOTEST_OPT_WITH_COVERAGE) ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html

.PHONY: coverage-all
coverage-all:
	$(GOTEST) $(GOTEST_OPT_WITH_COVERAGE_LONG_RUNNING) ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html

.PHONY: build-examples
build-examples:
	for d in examples/*; do \
  		echo "Build example: $${d}"; \
		(cd $${d}; go mod tidy && go build -race) || exit 1; \
	done;

.PHONY: test-ssl-certificates
test-ssl-certificates:
	cd scripts; ./test-ssl-certificates.sh

.PHONY: test-ci
test-ci: test-all build-examples test-ssl-certificates
